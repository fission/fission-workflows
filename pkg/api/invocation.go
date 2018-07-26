package api

import (
	"errors"
	"fmt"

	"github.com/fission/fission-workflows/pkg/api/aggregates"
	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/fission/fission-workflows/pkg/util"
)

const ErrInvocationCanceled = "workflow invocation was canceled"

// Invocation contains the API functionality for controlling (workflow) invocations.
// This includes starting, stopping, and completing invocations.
type Invocation struct {
	es fes.Backend
}

// NewInvocationAPI creates the Invocation API.
func NewInvocationAPI(esClient fes.Backend) *Invocation {
	return &Invocation{esClient}
}

// Invoke triggers the start of the invocation using the provided specification.
// The function either returns the invocationID of the invocation or an error.
// The error can be a validate.Err, proto marshall error, or a fes error.
func (ia *Invocation) Invoke(spec *types.WorkflowInvocationSpec) (string, error) {
	err := validate.WorkflowInvocationSpec(spec)
	if err != nil {
		return "", err
	}

	id := fmt.Sprintf("wi-%s", util.UID())

	event, err := fes.NewEvent(*aggregates.NewWorkflowInvocationAggregate(id), &events.InvocationCreated{
		Spec: spec,
	})
	if err != nil {
		return "", err
	}
	err = ia.es.Append(event)
	if err != nil {
		return "", err
	}

	return id, nil
}

// Cancel halts an invocation. This does not guarantee that tasks currently running are halted,
// but beyond the invocation will not progress any further than those tasks. The state of the invocation will
// become ABORTED. If the API fails to append the event to the event store, it will return an error.
func (ia *Invocation) Cancel(invocationID string) error {
	if len(invocationID) == 0 {
		return validate.NewError("invocationID", errors.New("id should not be empty"))
	}

	event, err := fes.NewEvent(*aggregates.NewWorkflowInvocationAggregate(invocationID), &events.InvocationCanceled{
		Error: &types.Error{
			Message: ErrInvocationCanceled,
		},
	})
	if err != nil {
		return err
	}
	event.Hints = &fes.EventHints{Completed: true}
	err = ia.es.Append(event)
	if err != nil {
		return err
	}
	return nil
}

// Complete forces the completion of an invocation. This function - used by the controller - is the only way
// to ensure that a workflow invocation turns into the COMPLETED state.
// If the API fails to append the event to the event store, it will return an error.
func (ia *Invocation) Complete(invocationID string, output *types.TypedValue) error {
	if len(invocationID) == 0 {
		return validate.NewError("invocationID", errors.New("id should not be empty"))
	}

	event, err := fes.NewEvent(*aggregates.NewWorkflowInvocationAggregate(invocationID), &events.InvocationCompleted{
		Output: output,
	})
	if err != nil {
		return err
	}
	event.Hints = &fes.EventHints{Completed: true}
	return ia.es.Append(event)
}

// Fail changes the state of the invocation to FAILED.
// Optionally you can provide a custom error message to indicate the specific reason for the FAILED state.
// If the API fails to append the event to the event store, it will return an error.
func (ia *Invocation) Fail(invocationID string, errMsg error) error {
	if len(invocationID) == 0 {
		return validate.NewError("invocationID", errors.New("id should not be empty"))
	}

	var msg string
	if errMsg != nil {
		msg = errMsg.Error()
	}
	event, err := fes.NewEvent(*aggregates.NewWorkflowInvocationAggregate(invocationID), &events.InvocationFailed{
		Error: &types.Error{
			Message: msg,
		},
	})
	if err != nil {
		return err
	}
	event.Hints = &fes.EventHints{Completed: true}
	return ia.es.Append(event)
}

// AddTask provides functionality to add a task to a specific invocation (instead of a workflow).
// This allows users to modify specific invocations (see dynamic API).
// The error can be a validate.Err, proto marshall error, or a fes error.
func (ia *Invocation) AddTask(invocationID string, task *types.Task) error {
	if len(invocationID) == 0 {
		return validate.NewError("invocationID", errors.New("id should not be empty"))
	}
	err := validate.Task(task)
	if err != nil {
		return err
	}

	event, err := fes.NewEvent(*aggregates.NewWorkflowInvocationAggregate(invocationID), &events.InvocationTaskAdded{
		Task: task,
	})
	if err != nil {
		return err
	}
	return ia.es.Append(event)
}
