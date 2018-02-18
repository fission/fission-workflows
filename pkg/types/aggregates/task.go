package aggregates

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/gogo/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

const (
	TypeTaskInvocation = "function"
)

type TaskInvocation struct {
	*fes.AggregatorMixin
	*types.TaskInvocation
}

func NewTaskInvocation(id string, fi *types.TaskInvocation) *TaskInvocation {
	tia := &TaskInvocation{
		TaskInvocation: fi,
	}

	tia.AggregatorMixin = fes.NewAggregatorMixin(tia, *NewTaskInvocationAggregate(id))

	return tia
}

func NewTaskInvocationAggregate(id string) *fes.Aggregate {
	return &fes.Aggregate{
		Id:   id,
		Type: TypeTaskInvocation,
	}
}

func (ti *TaskInvocation) ApplyEvent(event *fes.Event) error {

	eventType, err := events.ParseFunction(event.Type)
	if err != nil {
		return err
	}

	switch eventType {
	case events.Function_TASK_STARTED:
		fn := &types.TaskInvocation{}
		err = proto.Unmarshal(event.Data, fn)
		if err != nil {
			return fmt.Errorf("failed to unmarshal event: '%v' (%v)", event, err)
		}

		ti.TaskInvocation = &types.TaskInvocation{
			Metadata: fn.Metadata,
			Spec:     fn.Spec,
			Status: &types.TaskInvocationStatus{
				Status:    types.TaskInvocationStatus_IN_PROGRESS,
				UpdatedAt: event.Timestamp,
			},
		}
	case events.Function_TASK_SUCCEEDED:
		invoc := &types.TaskInvocation{}
		err = proto.Unmarshal(event.Data, invoc)
		if err != nil {
			return fmt.Errorf("failed to unmarshal event: '%v' (%v)", event, err)
		}

		ti.Status.Output = invoc.Status.Output
		ti.Status.Status = types.TaskInvocationStatus_SUCCEEDED
		ti.Status.UpdatedAt = event.Timestamp
	case events.Function_TASK_ABORTED:
		ti.Status.Status = types.TaskInvocationStatus_ABORTED
		ti.Status.UpdatedAt = event.Timestamp
	case events.Function_TASK_FAILED:
		fnErr := &types.Error{}
		err = proto.Unmarshal(event.Data, fnErr)
		if err != nil {
			fnErr.Code = "error"
			fnErr.Message = err.Error()
			log.Errorf("failed to unmarshal event: '%v' (%v)", event, err)
		}

		ti.Status.Status = types.TaskInvocationStatus_FAILED
		ti.Status.Error = fnErr
		ti.Status.UpdatedAt = event.Timestamp
	case events.Function_TASK_SKIPPED:
		ti.Status.Status = types.TaskInvocationStatus_SKIPPED
		ti.Status.UpdatedAt = event.Timestamp
	default:
		log.WithFields(log.Fields{
			"event": event,
		}).Warn("Skipping unimplemented event.")
	}
	return nil
}

func (ti *TaskInvocation) GenericCopy() fes.Aggregator {
	n := &TaskInvocation{
		TaskInvocation: ti.Copy(),
	}
	n.AggregatorMixin = ti.CopyAggregatorMixin(n)
	return n
}

func (ti *TaskInvocation) Copy() *types.TaskInvocation {
	return proto.Clone(ti.TaskInvocation).(*types.TaskInvocation)
}
