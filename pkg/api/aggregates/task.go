package aggregates

import (
	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
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
	if err := ti.ensureNextEvent(event); err != nil {
		return err
	}
	eventData, err := fes.ParseEventData(event)
	if err != nil {
		return err
	}

	switch m := eventData.(type) {
	case *events.TaskStarted:
		ti.TaskInvocation = &types.TaskInvocation{
			Metadata: &types.ObjectMetadata{
				Id:         m.GetSpec().TaskId,
				CreatedAt:  event.Timestamp,
				Generation: 1,
			},
			Spec: m.GetSpec(),
			Status: &types.TaskInvocationStatus{
				Status: types.TaskInvocationStatus_IN_PROGRESS,
			},
		}
	case *events.TaskSucceeded:
		ti.Status.Output = m.GetResult().Output
		ti.Status.Status = types.TaskInvocationStatus_SUCCEEDED
	case *events.TaskFailed:
		ti.Status.Error = m.GetError()
		ti.Status.Status = types.TaskInvocationStatus_FAILED
	case *events.TaskSkipped:
		// TODO ensure that object (spec/status) is present
		ti.Status.Status = types.TaskInvocationStatus_SKIPPED
	default:
		key := ti.Aggregate()
		logrus.Debugf("task ------> %T", m)
		return fes.ErrUnsupportedEntityEvent.WithAggregate(&key).WithEvent(event)
	}
	ti.Metadata.Generation++
	ti.Status.UpdatedAt = event.GetTimestamp()
	return nil
}

func (ti *TaskInvocation) CopyEntity() fes.Entity {
	n := &TaskInvocation{
		TaskInvocation: ti.Copy(),
	}
	n.BaseEntity = ti.CopyBaseEntity(n)
	return n
}

func (ti *TaskInvocation) Copy() *types.TaskInvocation {
	return proto.Clone(ti.TaskInvocation).(*types.TaskInvocation)
}

func (ti *TaskInvocation) ensureNextEvent(event *fes.Event) error {
	if err := fes.ValidateEvent(event); err != nil {
		return err
	}

	if event.Aggregate.Type != TypeTaskInvocation {
		logrus.Info("task check")
		return fes.ErrUnsupportedEntityEvent.WithEntity(ti).WithEvent(event)
	}
	// TODO check sequence of event
	return nil
}
