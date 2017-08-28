package workflow

import (
	"errors"

	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/events"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

type reduceFunc func(currentState types.Workflow, event *eventstore.Event) (newState *types.Workflow, err error)

var eventMapping = map[events.Workflow]reduceFunc{
	events.Workflow_WORKFLOW_CREATED: created,
	events.Workflow_WORKFLOW_DELETED: deleted,
	events.Workflow_WORKFLOW_PARSED:  parsed,
	events.Workflow_WORKFLOW_UPDATED: skip,
}

func Initial() *types.Workflow {
	return &types.Workflow{}
}

func From(events ...*eventstore.Event) (currentState *types.Workflow, err error) {
	return Apply(*Initial(), events...)
}

func Apply(currentState types.Workflow, seqEvents ...*eventstore.Event) (newState *types.Workflow, err error) {
	// Check if it is indeed next event (maybe wrap in a projectionContainer)
	newState = &currentState
	for _, event := range seqEvents {

		eventType, err := events.ParseWorkflow(event.GetType())
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

func created(currentState types.Workflow, event *eventstore.Event) (newState *types.Workflow, err error) {
	if currentState != (types.Workflow{}) {
		return nil, errors.New("invalid event") // TODO fix errors
	}

	spec := &types.WorkflowSpec{}
	err = ptypes.UnmarshalAny(event.Data, spec)
	if err != nil {
		return nil, err
	}

	return &types.Workflow{
		Metadata: &types.ObjectMetadata{
			Id:        event.GetEventId().GetSubjects()[1], // TODO remove this hardcoding
			CreatedAt: event.GetTime(),
		},
		Spec: spec,
		Status: &types.WorkflowStatus{
			Status:    types.WorkflowStatus_UNKNOWN, // TODO Nest into own state machine maybe
			UpdatedAt: event.GetTime(),
		},
	}, nil
}

func deleted(currentState types.Workflow, event *eventstore.Event) (newState *types.Workflow, err error) {
	if currentState != (types.Workflow{}) {
		return nil, errors.New("invalid event") // TODO fix errors
	}

	currentState.Status.UpdatedAt = event.GetTime()
	currentState.Status.Status = types.WorkflowStatus_DELETED

	return &currentState, nil
}

func parsed(currentState types.Workflow, event *eventstore.Event) (newState *types.Workflow, err error) {
	status := &types.WorkflowStatus{}
	err = ptypes.UnmarshalAny(event.Data, status)
	if err != nil {
		return nil, err
	}

	currentState.Status.UpdatedAt = event.GetTime()
	currentState.Status.Status = types.WorkflowStatus_READY
	currentState.Status.ResolvedTasks = status.GetResolvedTasks()

	return &currentState, nil
}

func skip(currentState types.Workflow, event *eventstore.Event) (newState *types.Workflow, err error) {
	logrus.WithFields(logrus.Fields{
		"currentState": currentState,
		"event":        event,
	}).Debug("Skipping unimplemented event.")
	return &currentState, nil
}
