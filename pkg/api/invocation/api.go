package invocation

import (
	"errors"

	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/eventstore/eventids"
	"github.com/fission/fission-workflow/pkg/projector/project"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/events"
	"github.com/golang/protobuf/ptypes"
	"github.com/satori/go.uuid"
)

type Api struct {
	esClient  eventstore.Client
	Projector project.InvocationProjector // TODO move projection functions out?
}

func NewApi(esClient eventstore.Client, projector project.InvocationProjector) *Api {
	return &Api{esClient, projector}
}

// Commands
func (ia *Api) Invoke(invocation *types.WorkflowInvocationSpec) (string, error) {
	if len(invocation.WorkflowId) == 0 {
		return "", errors.New("WorkflowId is required")
	}

	id := uuid.NewV4().String()

	data, err := ptypes.MarshalAny(invocation)
	if err != nil {
		return "", err
	}

	event := eventstore.NewEvent(ia.createSubject(id), events.Invocation_INVOCATION_CREATED.String(), data)

	err = ia.esClient.Append(event)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (ia *Api) Cancel(invocationId string) error {
	if len(invocationId) == 0 {
		return errors.New("invocationId is required")
	}

	event := eventstore.NewEvent(ia.createSubject(invocationId), events.Invocation_INVOCATION_CANCELED.String(), nil)

	err := ia.esClient.Append(event)
	if err != nil {
		return err
	}
	return nil
}

// TODO might want to split out public vs. internal api
func (ia *Api) Success(invocationId string, output string) error {
	if len(invocationId) == 0 {
		return errors.New("invocationId is required")
	}

	status := &types.WorkflowInvocationStatus{
		Output: output,
	}

	data, err := ptypes.MarshalAny(status)
	if err != nil {
		return err
	}

	event := eventstore.NewEvent(ia.createSubject(invocationId), events.Invocation_INVOCATION_COMPLETED.String(), data)

	err = ia.esClient.Append(event)
	if err != nil {
		return err
	}
	return nil
}

func (ia *Api) Fail(invocationId string) {
	panic("not implemented")
}

func (ia *Api) createSubject(invocationId string) *eventstore.EventID {
	return eventids.NewSubject(types.SUBJECT_INVOCATION, invocationId)
}

// TODO move projection functions out?
func (ia *Api) Get(invocationId string) (*types.WorkflowInvocation, error) {
	return ia.Projector.Get(invocationId)
}

func (ia *Api) List(query string) ([]string, error) {
	return ia.Projector.List(query)
}
