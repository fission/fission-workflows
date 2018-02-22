package invocation

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

type Api struct {
	es fes.Backend
}

func NewApi(esClient fes.Backend) *Api {
	return &Api{esClient}
}

func (ia *Api) Invoke(invocation *types.WorkflowInvocationSpec) (string, error) {
	err := validate.WorkflowInvocationSpec(invocation)
	if err != nil {
		return "", err
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

	return ia.es.Append(&fes.Event{
		Type:      events.Invocation_INVOCATION_COMPLETED.String(),
		Aggregate: aggregates.NewWorkflowInvocationAggregate(invocationId),
		Timestamp: ptypes.TimestampNow(),
		Data:      data,
		Hints:     &fes.EventHints{Completed: true},
	})
}

func (ia *Api) Fail(invocationId string, errMsg error) error {
	if len(invocationId) == 0 {
		return errors.New("invocationId is required")
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
		Aggregate: aggregates.NewWorkflowInvocationAggregate(invocationId),
		Timestamp: ptypes.TimestampNow(),
		Data:      data,
		Hints:     &fes.EventHints{Completed: true},
	})
}

func (ia *Api) AddTask(invocationId string, task *types.Task) error {
	if len(invocationId) == 0 {
		return errors.New("invocationId is required")
	}
	err := validate.Task(task)
	if err != nil {
		return err
	}

	data, err := proto.Marshal(task)
	if err != nil {
		return err
	}

	// Submit dynamic task
	return ia.es.Append(&fes.Event{
		Type:      events.Invocation_INVOCATION_TASK_ADDED.String(),
		Aggregate: aggregates.NewWorkflowInvocationAggregate(invocationId),
		Timestamp: ptypes.TimestampNow(),
		Data:      data,
	})
}
