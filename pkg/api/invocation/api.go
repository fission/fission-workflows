package invocation

import (
	"errors"

	"fmt"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

type Api struct {
	es          fes.EventStore
	invocations fes.CacheReader // TODO move projection functions out?
}

func NewApi(esClient fes.EventStore, invocations fes.CacheReader) *Api {
	return &Api{esClient, invocations}
}

func (ia *Api) Invoke(invocation *types.WorkflowInvocationSpec) (string, error) {
	if len(invocation.WorkflowId) == 0 {
		return "", errors.New("workflowId is required")
	}

	id := fmt.Sprintf("wi-%s", util.Uid())

	data, err := proto.Marshal(invocation)
	if err != nil {
		return "", err
	}

	event := &fes.Event{
		Type:      events.Invocation_INVOCATION_CREATED.String(),
		Aggregate: aggregates.NewWorkflowInvocationAggregate(id),
		Timestamp: ptypes.TimestampNow(),
		Data:      data,
	}

	err = ia.es.HandleEvent(event)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (ia *Api) Cancel(invocationId string) error {
	if len(invocationId) == 0 {
		return errors.New("invocationId is required")
	}

	event := &fes.Event{
		Type:      events.Invocation_INVOCATION_CANCELED.String(),
		Aggregate: aggregates.NewWorkflowInvocationAggregate(invocationId),
		Timestamp: ptypes.TimestampNow(),
	}
	err := ia.es.HandleEvent(event)
	if err != nil {
		return err
	}
	return nil
}

// TODO might want to split out public vs. internal api
func (ia *Api) MarkCompleted(invocationId string, output *types.TypedValue) error {
	if len(invocationId) == 0 {
		return errors.New("invocationId is required")
	}

	status := &types.WorkflowInvocationStatus{
		Output: output,
	}

	data, err := proto.Marshal(status)
	if err != nil {
		return err
	}

	event := &fes.Event{
		Type:      events.Invocation_INVOCATION_COMPLETED.String(),
		Aggregate: aggregates.NewWorkflowInvocationAggregate(invocationId),
		Timestamp: ptypes.TimestampNow(),
		Data:      data,
	}

	err = ia.es.HandleEvent(event)
	if err != nil {
		return err
	}
	return nil
}

func (ia *Api) MarkFailed(invocationId string) {
	panic("not implemented")
}

func (ia *Api) Get(invocationId string) (*types.WorkflowInvocation, error) {
	wi := aggregates.NewWorkflowInvocation(invocationId, &types.WorkflowInvocation{})
	err := ia.invocations.Get(wi)
	if err != nil {
		return nil, err
	}

	return wi.WorkflowInvocation, nil
}

func (ia *Api) List(query string) ([]string, error) {
	results := []string{}
	as := ia.invocations.List()
	for _, a := range as {
		if a.Type != aggregates.TYPE_WORKFLOW_INVOCATION {
			return nil, errors.New("invalid type in invocation cache")
		}

		results = append(results, a.Id)
	}
	return results, nil
}
