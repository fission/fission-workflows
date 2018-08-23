package api

import (
	"errors"

	"github.com/fission/fission-workflows/pkg/api/aggregates"
	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

// TODO move events here

// Task contains the API functionality for controlling the lifecycle of individual tasks.
// This includes starting, stopping and completing tasks.
type Task struct {
	runtime    map[string]fnenv.Runtime // TODO support AsyncRuntime
	es         fes.Backend
	dynamicAPI *Dynamic
}

// NewTaskAPI creates the Task API.
func NewTaskAPI(runtime map[string]fnenv.Runtime, esClient fes.Backend, api *Dynamic) *Task {
	return &Task{
		runtime:    runtime,
		es:         esClient,
		dynamicAPI: api,
	}
}

// Invoke starts the execution of a task, changing the state of the task into RUNNING.
// Currently it executes the underlying function synchronously and manage the execution until completion.
// TODO make asynchronous
func (ap *Task) Invoke(spec *types.TaskInvocationSpec, opts ...CallOption) (*types.TaskInvocation, error) {
	cfg := parseCallOptions(opts)
	err := validate.TaskInvocationSpec(spec)
	if err != nil {
		return nil, err
	}

	taskID := spec.TaskId // assumption: 1 task == 1 TaskInvocation (How to deal with retries? Same invocation?)
	task := &types.TaskInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        taskID,
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec: spec,
	}

	aggregate := aggregates.NewWorkflowInvocationAggregate(spec.InvocationId)
	event, err := fes.NewEvent(*aggregates.NewTaskInvocationAggregate(taskID), &events.TaskStarted{
		Spec: spec,
	})
	event.Parent = aggregate
	if err != nil {
		return nil, err
	}

	// TODO propagate context
	fnResult, err := ap.runtime[spec.FnRef.Runtime].Invoke(spec, fnenv.WithContext(cfg.ctx))
	if fnResult == nil && err == nil {
		err = errors.New("function crashed")
	}
	if err != nil {
		// TODO improve error handling here (retries? internal or task related error?)
		logrus.WithField("task", spec.FnRef).
			WithField("wi", spec.InvocationId).
			WithField("task", spec.TaskId).
			Infof("Failed to invoke task: %v", err)
		esErr := ap.Fail(spec.InvocationId, taskID, err.Error())
		if esErr != nil {
			return nil, esErr
		}
		return nil, err
	}

	// TODO to a middleware component
	if typedvalues.IsControlFlow(typedvalues.ValueType(fnResult.GetOutput().GetType())) {
		logrus.Info("Adding dynamic flow")
		flow, err := typedvalues.FormatControlFlow(fnResult.GetOutput())
		if err != nil {
			return nil, err
		}
		err = ap.dynamicAPI.AddDynamicFlow(spec.InvocationId, taskID, *flow)
		if err != nil {
			return nil, err
		}
	}

	if fnResult.Status == types.TaskInvocationStatus_SUCCEEDED {
		event, err := fes.NewEvent(*aggregates.NewTaskInvocationAggregate(taskID), &events.TaskSucceeded{
			Result: fnResult,
		})
		if err != nil {
			return nil, err
		}
		event.Parent = aggregate
		err = ap.es.Append(event)
	} else {
		err = ap.Fail(spec.InvocationId, taskID, fnResult.Error.GetMessage())
	}
	if err != nil {
		return nil, err
	}

	task.Status = fnResult
	return task, nil
}

// Fail forces the failure of a task. This turns the state of a task into FAILED.
// If the API fails to append the event to the event store, it will return an error.
func (ap *Task) Fail(invocationID string, taskID string, errMsg string) error {
	if len(invocationID) == 0 {
		return validate.NewError("invocationID", errors.New("id should not be empty"))
	}
	if len(taskID) == 0 {
		return validate.NewError("taskID", errors.New("id should not be empty"))
	}

	event, err := fes.NewEvent(*aggregates.NewTaskInvocationAggregate(taskID), &events.TaskFailed{
		Error: &types.Error{Message: errMsg},
	})
	if err != nil {
		return err
	}
	event.Parent = aggregates.NewWorkflowInvocationAggregate(invocationID)
	return ap.es.Append(event)
}
