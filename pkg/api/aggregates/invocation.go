package aggregates

import (
	"errors"

	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

const (
	TypeWorkflowInvocation = "invocation"
)

type WorkflowInvocation struct {
	*fes.BaseEntity
	*types.WorkflowInvocation
}

func NewWorkflowInvocation(invocationID string, wi ...*types.WorkflowInvocation) *WorkflowInvocation {
	wia := &WorkflowInvocation{}
	if len(wi) > 0 {
		wia.WorkflowInvocation = wi[0]
	}

	wia.BaseEntity = fes.NewBaseEntity(wia, *NewWorkflowInvocationAggregate(invocationID))
	return wia
}

func NewWorkflowInvocationAggregate(invocationID string) *fes.Aggregate {
	return &fes.Aggregate{
		Id:   invocationID,
		Type: TypeWorkflowInvocation,
	}
}

func (wi *WorkflowInvocation) ApplyEvent(event *fes.Event) error {
	// If the event is a task event, use the Task Aggregate to resolve it.
	if event.Aggregate.Type == TypeTaskInvocation {
		return wi.applyTaskEvent(event)
	}

	if err := wi.ensureNextEvent(event); err != nil {
		return err
	}
	eventData, err := fes.ParseEventData(event)
	if err != nil {
		return err
	}

	switch m := eventData.(type) {
	case *events.InvocationCreated:
		wi.BaseEntity = fes.NewBaseEntity(wi, *event.Aggregate)
		wi.WorkflowInvocation = &types.WorkflowInvocation{
			Metadata: &types.ObjectMetadata{
				Id:        event.Aggregate.Id,
				CreatedAt: event.Timestamp,
			},
			Spec: m.GetSpec(),
			Status: &types.WorkflowInvocationStatus{
				Status:       types.WorkflowInvocationStatus_IN_PROGRESS,
				Tasks:        map[string]*types.TaskInvocation{},
				DynamicTasks: map[string]*types.Task{},
			},
		}
	case *events.InvocationCanceled:
		wi.Status.Status = types.WorkflowInvocationStatus_ABORTED
		wi.Status.Error = m.GetError()
	case *events.InvocationCompleted:
		wi.Status.Status = types.WorkflowInvocationStatus_SUCCEEDED
		wi.Status.Output = m.GetOutput()
	case *events.InvocationTaskAdded:
		task := m.GetTask()
		if wi.Status.DynamicTasks == nil {
			wi.Status.DynamicTasks = map[string]*types.Task{}
		}
		wi.Status.DynamicTasks[task.ID()] = task
	case *events.InvocationFailed:
		wi.Status.Error = m.GetError()
		wi.Status.Status = types.WorkflowInvocationStatus_FAILED
	default:
		key := wi.Aggregate()
		logrus.Debugf("------> %T", m)
		return fes.ErrUnsupportedEntityEvent.WithAggregate(&key).WithEvent(event)
	}
	wi.Metadata.Generation++
	wi.Status.UpdatedAt = event.GetTimestamp()
	return err
}

func (wi *WorkflowInvocation) applyTaskEvent(event *fes.Event) error {
	if wi.Aggregate() != *event.Parent {
		return errors.New("function does not belong to invocation")
	}
	taskID := event.Aggregate.Id
	task, ok := wi.Status.Tasks[taskID]
	if !ok {
		task = types.NewTaskInvocation(taskID)
	}
	ti := NewTaskInvocation(taskID, task)
	err := ti.ApplyEvent(event)
	if err != nil {
		return err
	}

	if wi.Status.Tasks == nil {
		wi.Status.Tasks = map[string]*types.TaskInvocation{}
	}
	wi.Status.Tasks[taskID] = ti.TaskInvocation

	return nil
}

func (wi *WorkflowInvocation) CopyEntity() fes.Entity {
	n := &WorkflowInvocation{
		WorkflowInvocation: wi.Copy(),
	}
	n.BaseEntity = wi.CopyBaseEntity(n)
	return n
}

func (wi *WorkflowInvocation) Copy() *types.WorkflowInvocation {
	return proto.Clone(wi.WorkflowInvocation).(*types.WorkflowInvocation)
}

func NewInvocationEntity() fes.Entity {
	return NewWorkflowInvocation("")
}

func (wi *WorkflowInvocation) ensureNextEvent(event *fes.Event) error {
	if err := fes.ValidateEvent(event); err != nil {
		return err
	}

	if event.Aggregate.Type != TypeWorkflowInvocation {
		logrus.Info("workflow check")
		return fes.ErrUnsupportedEntityEvent.WithEntity(wi).WithEvent(event)
	}
	// TODO check sequence of event
	return nil
}
