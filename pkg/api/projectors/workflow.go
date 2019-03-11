package projectors

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/ptypes"
)

type Workflow struct {
}

func NewWorkflow() *Workflow {
	return &Workflow{}
}

func (w *Workflow) Project(base fes.Entity, events ...*fes.Event) (updated fes.Entity, err error) {
	var wf *types.Workflow
	if base == nil {
		wf = &types.Workflow{}
	} else {
		var ok bool
		wf, ok = base.(*types.Workflow)
		if !ok {
			return nil, fmt.Errorf("entity expected workflow, but was %T", base)
		}
		wf = wf.Copy()
	}

	for _, event := range events {
		err := w.project(wf, event)
		if err != nil {
			return wf, err
		}
	}
	return wf, nil
}

func (w *Workflow) project(wf *types.Workflow, event *fes.Event) error {
	if err := w.ensureValidEvent(event); err != nil {
		return err
	}

	eventData, err := fes.ParseEventData(event)
	if err != nil {
		return err
	}

	switch m := eventData.(type) {
	case *events.WorkflowCreated:
		spec := m.GetSpec()
		wf.Metadata = &types.ObjectMetadata{
			Id:        wf.GetMetadata().GetId(),
			Name:      spec.GetName(),
			CreatedAt: event.GetTimestamp(),
		}
		wf.Spec = spec
		wf.Status = &types.WorkflowStatus{
			Status: types.WorkflowStatus_PENDING,
		}
	case *events.WorkflowParsingFailed:
		wf.Status.Error = m.GetError()
		wf.Status.Status = types.WorkflowStatus_FAILED
	case *events.WorkflowParsed:
		wf.Status.Status = types.WorkflowStatus_READY
		//wf.Status.Tasks = m.GetTasks()
		for taskID, status := range m.GetTasks() {
			spec := wf.GetSpec().TaskSpec(taskID)
			if spec == nil {
				return fmt.Errorf("%s: unknown task", taskID)
			}
			wf.Status.AddTask(taskID, &types.Task{
				Metadata: &types.ObjectMetadata{
					Id: taskID,
				},
				Spec:   spec,
				Status: status,
			})
		}
	case *events.WorkflowDeleted:
		wf.Status.Status = types.WorkflowStatus_DELETED
	default:
		return fes.ErrUnsupportedEntityEvent.WithEvent(event)
	}
	wf.Metadata.Generation++
	wf.Status.UpdatedAt = event.GetTimestamp()
	return nil
}

func (w *Workflow) ensureValidEvent(event *fes.Event) error {
	if err := fes.ValidateEvent(event); err != nil {
		return err
	}

	if event.Aggregate.Type != types.TypeWorkflow {
		return fes.ErrUnsupportedEntityEvent.WithEvent(event)
	}
	// TODO check sequence of event
	return nil
}

func (w *Workflow) NewProjection(key fes.Aggregate) (fes.Entity, error) {
	if key.Type != types.TypeWorkflow {
		return nil, fes.ErrInvalidAggregate.WithAggregate(&key)
	}
	return &types.Workflow{
		Metadata: &types.ObjectMetadata{
			Id:        key.Id,
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec:   &types.WorkflowSpec{},
		Status: &types.WorkflowStatus{},
	}, nil
}

func NewWorkflowAggregate(id string) fes.Aggregate {
	return fes.Aggregate{
		Id:   id,
		Type: types.TypeWorkflow,
	}
}
