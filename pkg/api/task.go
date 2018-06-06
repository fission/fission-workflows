package api

import (
	"errors"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/golang/protobuf/proto"
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
func (ap *Task) Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocation, error) {
	err := validate.TaskInvocationSpec(spec)
	if err != nil {
		return nil, err
	}

	aggregate := aggregates.NewWorkflowInvocationAggregate(spec.InvocationId)
	taskID := spec.TaskId // assumption: 1 task == 1 TaskInvocation (How to deal with retries? Same invocation?)
	fn := &types.TaskInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        taskID,
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec: spec,
	}

	fnAny, err := proto.Marshal(fn)
	if err != nil {
		return nil, err
	}

	err = ap.es.Append(&fes.Event{
		Type:      events.Task_TASK_STARTED.String(),
		Parent:    aggregate,
		Aggregate: aggregates.NewTaskInvocationAggregate(taskID),
		Timestamp: ptypes.TimestampNow(),
		Data:      fnAny,
	})
	if err != nil {
		return nil, err
	}

	fnResult, err := ap.runtime[spec.FnRef.Runtime].Invoke(spec)
	if fnResult == nil && err == nil {
		err = errors.New("function crashed")
	}
	if err != nil {
		// TODO improve error handling here (retries? internal or task related error?)
		logrus.WithField("fn", spec.FnRef).
			WithField("wi", spec.InvocationId).
			WithField("task", spec.TaskId).
			Infof("Failed to invoke task: %v", err)
		esErr := ap.es.Append(&fes.Event{
			Type:      events.Task_TASK_FAILED.String(),
			Parent:    aggregate,
			Aggregate: aggregates.NewTaskInvocationAggregate(taskID),
			Timestamp: ptypes.TimestampNow(),
			Data:      fnAny,
		})
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

	fn.Status = fnResult
	fnStatusAny, err := proto.Marshal(fn)
	if err != nil {
		return nil, err
	}

	if fnResult.Status == types.TaskInvocationStatus_SUCCEEDED {
		err = ap.es.Append(&fes.Event{
			Type:      events.Task_TASK_SUCCEEDED.String(),
			Parent:    aggregate,
			Aggregate: aggregates.NewTaskInvocationAggregate(taskID),
			Timestamp: ptypes.TimestampNow(),
			Data:      fnStatusAny,
		})
	} else {
		err = ap.es.Append(&fes.Event{
			Type:      events.Task_TASK_FAILED.String(),
			Parent:    aggregate,
			Aggregate: aggregates.NewTaskInvocationAggregate(taskID),
			Timestamp: ptypes.TimestampNow(),
			Data:      fnStatusAny,
		})
	}
	if err != nil {
		return nil, err
	}

	fn.Status = fnResult
	return fn, nil
}

// Fail forces the failure of a task. This turns the state of a task into FAILED.
// If the API fails to append the event to the event store, it will return an error.
func (ap *Task) Fail(invocationID string, taskID string) error {
	if len(invocationID) == 0 {
		return errors.New("invocationID is required")
	}
	if len(taskID) == 0 {
		return errors.New("taskID is required")
	}

	return ap.es.Append(&fes.Event{
		Type:      events.Task_TASK_FAILED.String(),
		Parent:    aggregates.NewWorkflowInvocationAggregate(invocationID),
		Aggregate: aggregates.NewTaskInvocationAggregate(taskID),
		Timestamp: ptypes.TimestampNow(),
	})
}
