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
	es fes.EventStore
}

func NewApi(esClient fes.EventStore) *Api {
	return &Api{esClient}
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

func (ia *Api) Cancel(invocationId string) error {
	if len(invocationId) == 0 {
		return errors.New("invocationId is required")
	}

	event := &fes.Event{
		Type:      events.Invocation_INVOCATION_CANCELED.String(),
		Aggregate: aggregates.NewWorkflowInvocationAggregate(invocationId),
		Timestamp: ptypes.TimestampNow(),
		Hints:     &fes.EventHints{Completed: true},
	}
	err := ia.es.Append(event)
	if err != nil {
		return err
	}
	return nil
}

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

	err = ia.es.Append(&fes.Event{
		Type:      events.Invocation_INVOCATION_COMPLETED.String(),
		Aggregate: aggregates.NewWorkflowInvocationAggregate(invocationId),
		Timestamp: ptypes.TimestampNow(),
		Data:      data,
		Hints:     &fes.EventHints{Completed: true},
	})
	if err != nil {
		return err
	}
	return nil
}

func (ia *Api) MarkFailed(invocationId string) {
	panic("not implemented")
}
