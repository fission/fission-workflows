package aggregates

import (
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

const (
	TypeTaskInvocation = "task"
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

	eventData, err := unmarshalEventData(event)
	if err != nil {
		return err
	}

	switch m := eventData.(type) {
	case *events.TaskStarted:
		task := m.GetTask()
		ti.TaskInvocation = &types.TaskInvocation{
			Metadata: task.GetMetadata(),
			Spec:     task.GetSpec(),
			Status: &types.TaskInvocationStatus{
				Status:    types.TaskInvocationStatus_IN_PROGRESS,
				UpdatedAt: event.Timestamp,
			},
		}
	case *events.TaskSucceeded:
		ti.Status.Output = m.GetResult().Output
		ti.Status.Status = types.TaskInvocationStatus_SUCCEEDED
		ti.Status.UpdatedAt = event.Timestamp
	case *events.TaskFailed:
		// TODO validate event data
		if ti.Status == nil {
			ti.Status = &types.TaskInvocationStatus{}
		}
		ti.Status.Error = m.GetError()
		ti.Status.UpdatedAt = event.Timestamp
		ti.Status.Status = types.TaskInvocationStatus_FAILED
	case *events.TaskSkipped:
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
