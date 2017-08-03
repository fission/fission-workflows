package api

import (
	"errors"
	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/eventstore/eventids"
	"github.com/fission/fission-workflow/pkg/eventstore/events"
	"github.com/fission/fission-workflow/pkg/projector/project"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/golang/protobuf/ptypes"
	"github.com/satori/go.uuid"
)

const (
	INVOCATION_SUBJECT = "invocation"
)

type InvocationApi struct {
	esClient  eventstore.Client
	Projector project.InvocationProjector
}

func NewInvocationApi(esClient eventstore.Client, projector project.InvocationProjector) *InvocationApi {
	return &InvocationApi{esClient, projector}
}

// Commands
func (ia *InvocationApi) Invoke(invocation *types.WorkflowInvocationSpec) (string, error) {
	if len(invocation.WorkflowId) == 0 {
		return "", errors.New("WorkflowId is required")
	}

	// TODO validation
	id := uuid.NewV4().String()

	data, err := ptypes.MarshalAny(invocation)
	if err != nil {
		return "", err
	}

	event := events.New(ia.createSubject(id), types.InvocationEvent_INVOCATION_CREATED.String(), data)

	err = ia.esClient.Append(event)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (ia *InvocationApi) Cancel(invocationId string) error {
	// TODO validation

	event := events.New(ia.createSubject(invocationId), types.InvocationEvent_INVOCATION_CANCELED.String(), nil)

	err := ia.esClient.Append(event)
	if err != nil {
		return err
	}
	return nil
}

func (ia *InvocationApi) createSubject(invocationId string) *eventstore.EventID {
	return eventids.NewSubject(INVOCATION_SUBJECT, invocationId)
}

func (ia *InvocationApi) Get(invocationId string) (*types.WorkflowInvocationContainer, error) {
	return ia.Projector.Get(invocationId)
}

func (ia *InvocationApi) List(query string) ([]string, error) {
	return ia.Projector.List(query)
}
