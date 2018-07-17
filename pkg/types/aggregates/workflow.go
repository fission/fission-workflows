package aggregates

import (
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
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

func (wf *Workflow) ApplyEvent(event *fes.Event) error {
	eventData, err := unmarshalEventData(event)
	if err != nil {
		return err
	}

	switch m := eventData.(type) {
	case *events.WorkflowParsingFailed:
		wf.Status.Error = m.GetError()
		wf.Status.UpdatedAt = event.GetTimestamp()
		wf.Status.Status = types.WorkflowStatus_FAILED
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
				// TODO Nest into own state machine
				Status:    types.WorkflowStatus_PENDING,
				UpdatedAt: event.GetTimestamp(),
			},
		}
	case *events.WorkflowParsed:
		wf.Status.UpdatedAt = event.GetTimestamp()
		wf.Status.Status = types.WorkflowStatus_READY
		wf.Status.Tasks = m.GetTasks()
	case *events.WorkflowDeleted:
		wf.Status.Status = types.WorkflowStatus_DELETED
	default:
		log.WithFields(log.Fields{
			"event": event,
		}).Warn("Skipping unimplemented event.")
	}
	return nil
}

func (wf *Workflow) GenericCopy() fes.Entity {
	n := &Workflow{
		Workflow: wf.Copy(),
	}
	n.BaseEntity = wf.CopyBaseEntity(n)
	return n
}

func (wf *Workflow) Copy() *types.Workflow {
	return proto.Clone(wf.Workflow).(*types.Workflow)
}
