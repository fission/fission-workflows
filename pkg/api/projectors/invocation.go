package projectors

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/ptypes"
)

type WorkflowInvocation struct {
	taskRunProjector *TaskRun
}

func NewWorkflowInvocation() *WorkflowInvocation {
	return &WorkflowInvocation{
		taskRunProjector: NewTaskRun(),
	}
}

func (i *WorkflowInvocation) Project(base fes.Entity, events ...*fes.Event) (updated fes.Entity, err error) {
	var invocation *types.WorkflowInvocation
	if base == nil {
		invocation = &types.WorkflowInvocation{}
	} else {
		var ok bool
		invocation, ok = base.(*types.WorkflowInvocation)
		if !ok {
			return nil, fmt.Errorf("entity expected workflow, but was %T", base)
		}
		invocation = invocation.Copy()
	}

	for _, event := range events {
		err := i.project(invocation, event)
		if err != nil {
			return nil, err
		}
	}
	return invocation, nil
}

func (i *WorkflowInvocation) project(wi *types.WorkflowInvocation, event *fes.Event) error {
	// If the event is a task event, use the Task Aggregate to resolve it.
	if event.Aggregate.Type == types.TypeTaskRun {
		return i.applyTaskEvent(wi, event)
	}

	if err := i.ensureNextEvent(event); err != nil {
		return err
	}
	eventData, err := fes.ParseEventData(event)
	if err != nil {
		return err
	}

	switch m := eventData.(type) {
	case *events.InvocationCreated:
		wi.Metadata = &types.ObjectMetadata{
			Id:        event.Aggregate.Id,
			CreatedAt: event.Timestamp,
		}
		wi.Spec = m.GetSpec()
		wi.Status = &types.WorkflowInvocationStatus{
			Status:       types.WorkflowInvocationStatus_IN_PROGRESS,
			Tasks:        map[string]*types.TaskInvocation{},
			DynamicTasks: map[string]*types.Task{},
		}
	case *events.InvocationCanceled:
		wi.Status.Status = types.WorkflowInvocationStatus_ABORTED
		wi.Status.Error = m.GetError()
	case *events.InvocationCompleted:
		wi.Status.Status = types.WorkflowInvocationStatus_SUCCEEDED
		wi.Status.Output = m.GetOutput()
		wi.Status.OutputHeaders = m.GetOutputHeaders()
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
		//key := wi.Aggregate()
		return fes.ErrUnsupportedEntityEvent.WithEvent(event)
	}
	wi.Metadata.Generation++
	wi.Status.UpdatedAt = event.GetTimestamp()
	return err
}

func (i *WorkflowInvocation) NewProjection(key fes.Aggregate) (fes.Entity, error) {
	if key.Type != types.TypeInvocation {
		return nil, fes.ErrInvalidAggregate.WithAggregate(&key)
	}
	return &types.WorkflowInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        key.Id,
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec:   &types.WorkflowInvocationSpec{},
		Status: &types.WorkflowInvocationStatus{},
	}, nil
}

func (i *WorkflowInvocation) ensureNextEvent(event *fes.Event) error {
	if err := fes.ValidateEvent(event); err != nil {
		return err
	}

	if event.Aggregate.Type != types.TypeInvocation {
		return fes.ErrUnsupportedEntityEvent.WithEvent(event)
	}
	// TODO check sequence of event
	return nil
}

func (i *WorkflowInvocation) applyTaskEvent(wi *types.WorkflowInvocation, event *fes.Event) error {
	wiAggregate := fes.GetAggregate(wi)
	if wiAggregate != *event.Parent {
		return fmt.Errorf("event does not belong to invocation: (expected: %v, value: %v)", wiAggregate, *event.Parent)
	}
	taskID := event.Aggregate.Id
	task, ok := wi.Status.Tasks[taskID]
	if !ok {
		entity, _ := i.taskRunProjector.NewProjection(*event.Aggregate)
		task, _ = entity.(*types.TaskInvocation)
	}
	task = task.Copy()

	err := i.taskRunProjector.project(task, event)
	if err != nil {
		return err
	}

	if wi.Status.Tasks == nil {
		wi.Status.Tasks = map[string]*types.TaskInvocation{}
	}
	wi.Status.Tasks[taskID] = task
	return nil
}

func NewInvocationAggregate(invocationID string) fes.Aggregate {
	return fes.Aggregate{
		Id:   invocationID,
		Type: types.TypeInvocation,
	}
}
