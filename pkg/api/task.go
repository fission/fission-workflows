package api

import (
	"errors"
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/fission/fission-workflows/pkg/api/projectors"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues/controlflow"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

// Task contains the API functionality for controlling the lifecycle of individual tasks.
// This includes starting, stopping and completing tasks.
type Task struct {
	runtime    map[string]fnenv.Runtime
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
func (ap *Task) Invoke(spec *types.TaskInvocationSpec, opts ...CallOption) (*types.TaskInvocation, error) {
	log := logrus.WithField("fn", spec.FnRef).WithField("wi", spec.InvocationId).WithField("task", spec.TaskId)
	cfg := parseCallOptions(opts)
	err := validate.TaskInvocationSpec(spec)
	if err != nil {
		return nil, err
	}

	// Ensure that the task run has a task in its spec.
	// This is not part of the default validation, because it is (for now) in most cases fine to pass just the
	// invocationID, whereas here the task run really needs the task.
	if spec.GetTask() == nil {
		return nil, errors.New("task-run does not contain the task to be run")
	}

	// The assumption that we make for now every task has only one task invocation.
	// Therefore we use the same (task) ID for the task run.
	taskID := spec.TaskId
	task := &types.TaskInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        taskID,
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec: spec,
	}

	aggregate := projectors.NewInvocationAggregate(spec.InvocationId)
	event, err := fes.NewEvent(projectors.NewTaskRunAggregate(taskID), &events.TaskStarted{
		Spec: spec,
	})
	event.Parent = &aggregate
	if err != nil {
		return nil, err
	}

	fnResult, err := ap.runtime[spec.FnRef.Runtime].Invoke(spec, fnenv.WithContext(cfg.ctx),
		fnenv.AwaitWorklow(cfg.awaitWorkflow))
	if fnResult == nil && err == nil {
		err = errors.New("function crashed")
	}
	if err != nil {
		// TODO improve error handling here (retries? internal or task related error?)
		log.Infof("Failed to invoke task: %v", err)
		esErr := ap.Fail(spec.InvocationId, taskID, err.Error())
		if esErr != nil {
			return nil, esErr
		}
		return nil, err
	}

	// TODO to a middleware component
	if controlflow.IsControlFlow(fnResult.GetOutput()) {
		log.Info("Adding dynamic flow")
		flow, err := controlflow.UnwrapControlFlow(fnResult.GetOutput())
		if err != nil {
			return nil, err
		}
		err = ap.dynamicAPI.AddDynamicFlow(spec.InvocationId, taskID, *flow)
		if err != nil {
			return nil, err
		}
	}
	task.Status = fnResult

	if cfg.postTransformer != nil {
		err = cfg.postTransformer(task)
		if err != nil {
			return nil, err
		}
	}

	if fnResult.Status == types.TaskInvocationStatus_SUCCEEDED {
		event, err := fes.NewEvent(projectors.NewTaskRunAggregate(taskID), &events.TaskSucceeded{
			Result: fnResult,
		})
		if err != nil {
			return nil, err
		}
		event.Parent = &aggregate
		err = ap.es.Append(event)
	} else {
		err = ap.Fail(spec.InvocationId, taskID, fnResult.Error.GetMessage())
	}
	if err != nil {
		return nil, err
	}
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

	event, err := fes.NewEvent(projectors.NewTaskRunAggregate(taskID), &events.TaskFailed{
		Error: &types.Error{Message: errMsg},
	})
	if err != nil {
		return err
	}
	aggregate := projectors.NewInvocationAggregate(invocationID)
	event.Parent = &aggregate
	return ap.es.Append(event)
}

func (ap *Task) Prepare(spec *types.TaskInvocationSpec, expectedAt time.Time, opts ...CallOption) error {
	runtime, ok := ap.runtime[spec.GetFnRef().GetRuntime()]
	if !ok {
		return fmt.Errorf("could not find runtime for %s", spec.GetFnRef().Format())
	}

	// check if the runtime supports prewarming
	preparer, ok := runtime.(fnenv.Preparer)
	if !ok {
		return fmt.Errorf("runtime does not support prewarming")
	}

	return preparer.Prepare(*spec.FnRef, expectedAt)
}
