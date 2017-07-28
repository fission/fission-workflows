package api

import (
	"github.com/fission/fission-workflow/pkg/cache"
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

func NewInvocationApi(esClient eventstore.Client) *InvocationApi {
	return &InvocationApi{esClient, project.NewInvocationProjector(esClient, cache.NewMapCache())}
}

// Commands
func (ia *InvocationApi) Invoke(invocation *types.WorkflowInvocationSpec) (string, error) {
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

// Get(invocationID)

// Future: Search/List (time-based, workflow-based, status-based)
