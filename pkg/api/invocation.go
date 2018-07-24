package api

import (
	"errors"
	"fmt"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
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
func (ia *Invocation) Invoke(invocation *types.WorkflowInvocationSpec) (string, error) {
	err := validate.WorkflowInvocationSpec(invocation)
	if err != nil {
		return "", err
	}

	id := fmt.Sprintf("wi-%s", util.UID())
	data, err := proto.Marshal(invocation)
	if err != nil {
		return "", err
	}

	err = ia.es.Append(&fes.Event{
		Type:      events.Invocation_INVOCATION_CREATED.String(),
		Aggregate: aggregates.NewWorkflowInvocationAggregate(id),
		Timestamp: ptypes.TimestampNow(),
		Data:      data,
	})
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

	data, err := proto.Marshal(&types.Error{
		Message: ErrInvocationCanceled,
	})
	if err != nil {
		data = []byte{}
	}
	return ia.es.Append(&fes.Event{
		Type:      events.Invocation_INVOCATION_CANCELED.String(),
		Aggregate: aggregates.NewWorkflowInvocationAggregate(invocationID),
		Timestamp: ptypes.TimestampNow(),
		Data:      data,
		Hints:     &fes.EventHints{Completed: true},
	})
}

// Complete forces the completion of an invocation. This function - used by the controller - is the only way
// to ensure that a workflow invocation turns into the COMPLETED state.
// If the API fails to append the event to the event store, it will return an error.
func (ia *Invocation) Complete(invocationID string, output *types.TypedValue) error {
	if len(invocationID) == 0 {
		return validate.NewError("invocationID", errors.New("id should not be empty"))
	}

	data, err := proto.Marshal(&types.WorkflowInvocationStatus{
		Output: output,
	})
	if err != nil {
		return err
	}

	return ia.es.Append(&fes.Event{
		Type:      events.Invocation_INVOCATION_COMPLETED.String(),
		Aggregate: aggregates.NewWorkflowInvocationAggregate(invocationID),
		Timestamp: ptypes.TimestampNow(),
		Data:      data,
		Hints:     &fes.EventHints{Completed: true},
	})
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
	data, err := proto.Marshal(&types.Error{
		Message: msg,
	})
	if err != nil {
		data = []byte{}
	}
	return ia.es.Append(&fes.Event{
		Type:      events.Invocation_INVOCATION_FAILED.String(),
		Aggregate: aggregates.NewWorkflowInvocationAggregate(invocationID),
		Timestamp: ptypes.TimestampNow(),
		Data:      data,
		Hints:     &fes.EventHints{Completed: true},
	})
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

	data, err := proto.Marshal(task)
	if err != nil {
		return err
	}

	return ia.es.Append(&fes.Event{
		Type:      events.Invocation_INVOCATION_TASK_ADDED.String(),
		Aggregate: aggregates.NewWorkflowInvocationAggregate(invocationID),
		Timestamp: ptypes.TimestampNow(),
		Data:      data,
	})
}
