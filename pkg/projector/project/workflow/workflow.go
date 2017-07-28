package workflow

import (
	"errors"

	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

type reduceFunc func(currentState types.Workflow, event *eventstore.Event) (newState *types.Workflow, err error)

var eventMapping = map[types.WorkflowEvent]reduceFunc{
	types.WorkflowEvent_WORKFLOW_CREATED: created,
	types.WorkflowEvent_WORKFLOW_DELETED: skip,
	types.WorkflowEvent_WORKFLOW_UPDATED: skip,
}

func Initial() *types.Workflow {
	return &types.Workflow{}
}

func From(events ...*eventstore.Event) (currentState *types.Workflow, err error) {
	return Apply(*Initial(), events...)
}

func Apply(currentState types.Workflow, events ...*eventstore.Event) (newState *types.Workflow, err error) {
	// Check if it is indeed next event (maybe wrap in a projectionContainer)
	newState = &currentState
	for _, event := range events {

		eventType, err := types.ParseWorkflowEvent(event.GetType())
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

	return &types.Workflow{
		Metadata: &types.ObjectMetadata{
			Id:        event.GetEventId().GetSubjects()[1], // TODO remove this hardcoding
			CreatedAt: event.GetTime(),
		},
		Spec: spec,
	}, nil
}

func skip(currentState types.Workflow, event *eventstore.Event) (newState *types.Workflow, err error) {
	logrus.WithFields(logrus.Fields{
		"currentState": currentState,
		"event":        event,
	}).Debug("Skipping unimplemented event.")
	return &currentState, nil
}
