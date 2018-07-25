package aggregates

import (
	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

const (
	TypeTaskInvocation = "task"
)

type TaskInvocation struct {
	*fes.BaseEntity
	*types.TaskInvocation
}

func NewTaskInvocation(id string, fi *types.TaskInvocation) *TaskInvocation {
	tia := &TaskInvocation{
		TaskInvocation: fi,
	}

	tia.BaseEntity = fes.NewBaseEntity(tia, *NewTaskInvocationAggregate(id))

	return tia
}

func NewTaskInvocationAggregate(id string) *fes.Aggregate {
	return &fes.Aggregate{
		Id:   id,
		Type: TypeTaskInvocation,
	}
}

func (ti *TaskInvocation) ApplyEvent(event *fes.Event) error {

	eventData, err := fes.UnmarshalEventData(event)
	if err != nil {
		return err
	}

	switch m := eventData.(type) {
	case *events.TaskStarted:
		ti.TaskInvocation = &types.TaskInvocation{
			Metadata: types.NewObjectMetadata(m.GetSpec().TaskId),
			Spec:     m.GetSpec(),
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
			"aggregate": ti.Aggregate(),
		}).Warnf("Skipping unimplemented event: %T", eventData)
	}
	return nil
}

func (ti *TaskInvocation) GenericCopy() fes.Entity {
	n := &TaskInvocation{
		TaskInvocation: ti.Copy(),
	}
	n.BaseEntity = ti.CopyBaseEntity(n)
	return n
}

func (ti *TaskInvocation) Copy() *types.TaskInvocation {
	return proto.Clone(ti.TaskInvocation).(*types.TaskInvocation)
}
