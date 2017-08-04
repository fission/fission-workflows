package invocation

import (
	"errors"

	"fmt"

	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/invocationevent"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

/*
	Invocation Projection
*/
type reduceFunc func(invocation types.WorkflowInvocationContainer, event *eventstore.Event) (newState *types.WorkflowInvocationContainer, err error)

var eventMapping = map[types.InvocationEvent]reduceFunc{
	types.InvocationEvent_INVOCATION_CREATED:      created,       // Data: WorkflowInvocationContainer
	types.InvocationEvent_INVOCATION_CANCELED:     canceled,      // Data: {}
	types.InvocationEvent_INVOCATION_COMPLETED:    completed,     // Data: {}
	types.InvocationEvent_TASK_STARTED:            taskStarted,   // Data: FunctionInvocation
	types.InvocationEvent_TASK_FAILED:             taskFailed,    // Data: FunctionInvocation
	types.InvocationEvent_TASK_SUCCEEDED:          taskSucceeded, // Data: FunctionInvocation
	types.InvocationEvent_TASK_ABORTED:            taskAborted,
	types.InvocationEvent_TASK_HEARTBEAT_REQUEST:  skip,
	types.InvocationEvent_TASK_HEARTBEAT_RESPONSE: skip,
	types.InvocationEvent_TASK_SKIPPED:            skip,
}

func Initial() *types.WorkflowInvocationContainer {
	return &types.WorkflowInvocationContainer{}
}

func From(events ...*eventstore.Event) (currentState *types.WorkflowInvocationContainer, err error) {
	return Apply(*Initial(), events...)
}

func Apply(currentState types.WorkflowInvocationContainer, events ...*eventstore.Event) (newState *types.WorkflowInvocationContainer, err error) {
	// Check if it is indeed next event (maybe wrap in a projectionContainer)
	newState = &currentState
	for _, event := range events {

		eventType, err := invocationevent.Parse(event.GetType())
		if err != nil {
			return nil, err
		}
		newState, err = eventMapping[eventType](currentState, event)
		if err != nil {
			return nil, err
		}
	}

	return newState, nil
}

func created(currentState types.WorkflowInvocationContainer, event *eventstore.Event) (newState *types.WorkflowInvocationContainer, err error) {
	// Check if state
	if currentState != (types.WorkflowInvocationContainer{}) {
		//return nil, fmt.Errorf("invalid event '%v' for state '%v'", event, currentState) // TODO fix errors
		logrus.Warnf("invalid event '%v' for state '%v'", event, currentState)
		return &currentState, nil
	}

	spec := &types.WorkflowInvocationSpec{}
	err = ptypes.UnmarshalAny(event.Data, spec)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal event: '%v' (%v)", event, err)
	}

	return &types.WorkflowInvocationContainer{
		Metadata: &types.ObjectMetadata{
			Id:        event.GetEventId().GetSubjects()[1], // TODO remove this hardcoding,
			CreatedAt: event.GetTime(),
		},
		Spec: spec,
		Status: &types.WorkflowInvocationStatus{
			Status:    types.WorkflowInvocationStatus_IN_PROGRESS,
			Tasks:     map[string]*types.FunctionInvocation{},
			UpdatedAt: event.Time,
		},
	}, nil
}

func canceled(currentState types.WorkflowInvocationContainer, event *eventstore.Event) (newState *types.WorkflowInvocationContainer, err error) {
	// Canceling non-existent / already invocationCanceled invocation does nothing
	if currentState == (types.WorkflowInvocationContainer{}) {
		return nil, errors.New("Unknown state") // TODO fix errors
	}
	currentState.GetStatus().Status = types.WorkflowInvocationStatus_ABORTED
	currentState.GetStatus().UpdatedAt = event.Time
	return &currentState, nil
}

func completed(currentState types.WorkflowInvocationContainer, event *eventstore.Event) (newState *types.WorkflowInvocationContainer, err error) {
	// TODO do some state checking

	currentState.GetStatus().Status = types.WorkflowInvocationStatus_SUCCEEDED
	currentState.GetStatus().UpdatedAt = event.Time
	return &currentState, nil
}

func skip(currentState types.WorkflowInvocationContainer, event *eventstore.Event) (newState *types.WorkflowInvocationContainer, err error) {
	logrus.WithFields(logrus.Fields{
		"currentState": currentState,
		"event":        event,
	}).Debug("Skipping unimplemented event.")
	return &currentState, nil
}

func taskStarted(currentState types.WorkflowInvocationContainer, event *eventstore.Event) (newState *types.WorkflowInvocationContainer, err error) {
	fn := &types.FunctionInvocation{}
	err = ptypes.UnmarshalAny(event.Data, fn)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal event: '%v' (%v)", event, err)
	}

	if existing, ok := currentState.GetStatus().Tasks[fn.Spec.TaskId]; ok {
		logrus.Debug("Overwriting existing function invocation '%v'", existing)
	}

	currentState.GetStatus().Tasks[fn.Spec.TaskId] = &types.FunctionInvocation{
		Metadata: fn.Metadata,
		Spec:     fn.Spec,
		Status: &types.FunctionInvocationStatus{
			Status:    types.FunctionInvocationStatus_SCHEDULED,
			UpdatedAt: event.Time,
		},
	}
	return &currentState, nil
}

func taskFailed(currentState types.WorkflowInvocationContainer, event *eventstore.Event) (newState *types.WorkflowInvocationContainer, err error) {
	fn := &types.FunctionInvocation{}
	err = ptypes.UnmarshalAny(event.Data, fn)

	logrus.WithFields(logrus.Fields{
		"fn":           fn,
		"currentState": currentState,
		"event":        event,
	}).Infof(">>>")

	existing, ok := currentState.GetStatus().Tasks[fn.Spec.TaskId]
	if !ok {
		logrus.Warn("Non-recorded function failed '%v'", fn)
		return &currentState, nil
	}

	existing.Status.Status = types.FunctionInvocationStatus_FAILED
	existing.Status.UpdatedAt = event.Time

	return &currentState, nil
}

func taskSucceeded(currentState types.WorkflowInvocationContainer, event *eventstore.Event) (newState *types.WorkflowInvocationContainer, err error) {
	fn := &types.FunctionInvocation{}
	err = ptypes.UnmarshalAny(event.Data, fn)

	existing, ok := currentState.GetStatus().Tasks[fn.Spec.TaskId]
	if !ok {
		logrus.Warn("Non-recorded function succeeded '%v'", fn)
		return &currentState, nil
	}

	existing.Status.Status = types.FunctionInvocationStatus_SUCCEEDED
	existing.Status.UpdatedAt = event.Time

	return &currentState, nil
}

func taskAborted(currentState types.WorkflowInvocationContainer, event *eventstore.Event) (newState *types.WorkflowInvocationContainer, err error) {
	fn := &types.FunctionInvocation{}
	err = ptypes.UnmarshalAny(event.Data, fn)

	existing, ok := currentState.GetStatus().Tasks[fn.Spec.TaskId]
	if !ok {
		logrus.Warn("Non-recorded function aborted '%v'", fn)
		return &currentState, nil
	}

	existing.Status.Status = types.FunctionInvocationStatus_ABORTED
	existing.Status.UpdatedAt = event.Time

	return &currentState, nil
}
