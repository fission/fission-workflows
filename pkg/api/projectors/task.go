package projectors

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/ptypes"
)

type TaskRun struct {
}

func NewTaskRun() *TaskRun {
	return &TaskRun{}
}

func (t *TaskRun) Project(base fes.Entity, events ...*fes.Event) (updated fes.Entity, err error) {
	var taskRun *types.TaskInvocation
	if base == nil {
		taskRun = &types.TaskInvocation{}
	} else {
		var ok bool
		taskRun, ok = base.(*types.TaskInvocation)
		if !ok {
			return nil, fmt.Errorf("entity expected workflow, but was %T", base)
		}
		taskRun = taskRun.Copy()
	}

	for _, event := range events {
		err := t.project(taskRun, event)
		if err != nil {
			return nil, err
		}
	}
	return taskRun, nil
}

func (t *TaskRun) project(taskRun *types.TaskInvocation, event *fes.Event) error {
	if err := t.ensureNextEvent(event); err != nil {
		return err
	}
	eventData, err := fes.ParseEventData(event)
	if err != nil {
		return err
	}

	switch m := eventData.(type) {
	case *events.TaskStarted:
		taskRun.Metadata = &types.ObjectMetadata{
			Id:         m.GetSpec().TaskId,
			CreatedAt:  event.Timestamp,
			Generation: 1,
		}
		taskRun.Spec = m.GetSpec()
		taskRun.Status = &types.TaskInvocationStatus{
			Status: types.TaskInvocationStatus_IN_PROGRESS,
		}
	case *events.TaskSucceeded:
		taskRun.Status.Output = m.GetResult().Output
		taskRun.Status.OutputHeaders = m.GetResult().OutputHeaders
		taskRun.Status.Status = types.TaskInvocationStatus_SUCCEEDED
	case *events.TaskFailed:
		taskRun.Status.Error = m.GetError()
		taskRun.Status.Status = types.TaskInvocationStatus_FAILED
	case *events.TaskSkipped:
		// TODO ensure that object (spec/status) is present
		taskRun.Status.Status = types.TaskInvocationStatus_SKIPPED
	default:
		key := fes.GetAggregate(taskRun)
		return fes.ErrUnsupportedEntityEvent.WithAggregate(&key).WithEvent(event)
	}
	taskRun.Metadata.Generation++
	taskRun.Status.UpdatedAt = event.GetTimestamp()
	return nil

}

func (t *TaskRun) NewProjection(key fes.Aggregate) (fes.Entity, error) {
	if key.Type != types.TypeTaskRun {
		return nil, fes.ErrInvalidAggregate.WithAggregate(&key)
	}
	return &types.TaskInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        key.Id,
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec:   &types.TaskInvocationSpec{},
		Status: &types.TaskInvocationStatus{},
	}, nil
}

func (t *TaskRun) ensureNextEvent(event *fes.Event) error {
	if err := fes.ValidateEvent(event); err != nil {
		return err
	}

	if event.Aggregate.Type != types.TypeTaskRun {
		return fes.ErrUnsupportedEntityEvent.WithEvent(event)
	}
	// TODO check sequence of event
	return nil
}

func NewTaskRunAggregate(id string) fes.Aggregate {
	return fes.Aggregate{
		Id:   id,
		Type: types.TypeTaskRun,
	}
}
