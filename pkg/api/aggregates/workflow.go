package aggregates

import (
	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/proto"
)

const (
	TypeWorkflow = "workflow"
)

type Workflow struct {
	*fes.BaseEntity
	*types.Workflow
}

func NewWorkflow(workflowID string, wi ...*types.Workflow) *Workflow {
	wia := &Workflow{}
	if len(wi) > 0 {
		wia.Workflow = wi[0]
	}

	wia.BaseEntity = fes.NewBaseEntity(wia, *NewWorkflowAggregate(workflowID))
	return wia
}

func NewWorkflowAggregate(workflowID string) *fes.Aggregate {
	return &fes.Aggregate{
		Id:   workflowID,
		Type: TypeWorkflow,
	}
}

// ApplyEvent applies an event to this entity, typically mutating the state of the entity
//
// If unsuccessful no state will be mutated, and one of the following errors will be returned:
// - fes.ErrUnsupportedEntityEvent: the entity does not support the event type
func (wf *Workflow) ApplyEvent(event *fes.Event) error {

	if err := wf.ensureNextEvent(event); err != nil {
		return err
	}
	eventData, err := fes.ParseEventData(event)
	if err != nil {
		return err
	}

	switch m := eventData.(type) {
	case *events.WorkflowCreated:
		// Setup object
		wf.BaseEntity = fes.NewBaseEntity(wf, *event.Aggregate)
		wf.Workflow = &types.Workflow{
			Metadata: &types.ObjectMetadata{
				Id:        wf.Aggregate().Id,
				CreatedAt: event.GetTimestamp(),
			},
			Spec: m.GetSpec(),
			Status: &types.WorkflowStatus{
				Status: types.WorkflowStatus_PENDING,
			},
		}
	case *events.WorkflowParsingFailed:
		wf.Status.Error = m.GetError()
		wf.Status.Status = types.WorkflowStatus_FAILED
	case *events.WorkflowParsed:
		wf.Status.Status = types.WorkflowStatus_READY
		wf.Status.Tasks = m.GetTasks()
	case *events.WorkflowDeleted:
		wf.Status.Status = types.WorkflowStatus_DELETED
	default:
		key := wf.Aggregate()
		return fes.ErrUnsupportedEntityEvent.WithAggregate(&key).WithEvent(event)
	}
	wf.Metadata.Generation++
	wf.Status.UpdatedAt = event.GetTimestamp()
	return nil
}

func (wf *Workflow) CopyEntity() fes.Entity {
	n := &Workflow{
		Workflow: wf.Copy(),
	}
	n.BaseEntity = wf.CopyBaseEntity(n)
	return n
}

func (wf *Workflow) Copy() *types.Workflow {
	return proto.Clone(wf.Workflow).(*types.Workflow)
}

func (wf *Workflow) ensureNextEvent(event *fes.Event) error {
	if err := fes.ValidateEvent(event); err != nil {
		return err
	}

	if event.Aggregate.Type != TypeWorkflow {
		return fes.ErrUnsupportedEntityEvent.WithEntity(wf).WithEvent(event)
	}
	// TODO check sequence of event
	return nil
}

func NewWorkflowEntity() fes.Entity {
	return NewWorkflow("")
}
