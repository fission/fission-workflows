package invocation

import (
	"errors"

	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/invocationevent"
	"github.com/golang/protobuf/ptypes"
)

/*
	Invocation Projection
*/
type reduceFunc func(invocation types.WorkflowInvocationContainer, event *eventstore.Event) (newState *types.WorkflowInvocationContainer, err error)

var eventMapping = map[types.InvocationEvent]reduceFunc{
	types.InvocationEvent_INVOCATION_CREATED:   created,
	types.InvocationEvent_INVOCATION_CANCELED:  canceled,
	types.InvocationEvent_INVOCATION_COMPLETED: completed,
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
		return nil, errors.New("invalid event") // TODO fix errors
	}

	spec := &types.WorkflowInvocationSpec{}
	err = ptypes.UnmarshalAny(event.Data, spec)

	return &types.WorkflowInvocationContainer{
		Metadata: &types.ObjectMetadata{
			Id:        event.GetEventId().GetSubjects()[1], // TODO remove this hardcoding,
			CreatedAt: event.GetTime(),
		},
		Spec: spec,
		Status: &types.WorkflowInvocationStatus{
			Status:    types.WorkflowInvocationStatus_UNKNOWN,
			UpdatedAt: ptypes.TimestampNow(),
		},
	}, nil
}

func canceled(currentState types.WorkflowInvocationContainer, event *eventstore.Event) (newState *types.WorkflowInvocationContainer, err error) {
	// Canceling non-existent / already invocationCanceled invocation does nothing
	if currentState == (types.WorkflowInvocationContainer{}) {
		return nil, errors.New("Unknown state") // TODO fix errors
	}
	currentState.GetStatus().Status = types.WorkflowInvocationStatus_ABORTED
	return &currentState, nil
}

func completed(currentState types.WorkflowInvocationContainer, event *eventstore.Event) (newState *types.WorkflowInvocationContainer, err error) {
	// TODO do some state checking

	currentState.GetStatus().Status = types.WorkflowInvocationStatus_SUCCEEDED
	return &currentState, nil
}
