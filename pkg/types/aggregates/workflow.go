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
	*fes.AggregatorMixin
	*types.Workflow
}

func NewWorkflow(workflowID string, wi ...*types.Workflow) *Workflow {
	wia := &Workflow{}
	if len(wi) > 0 {
		wia.Workflow = wi[0]
	}

	wia.AggregatorMixin = fes.NewAggregatorMixin(wia, *NewWorkflowAggregate(workflowID))
	return wia
}

func NewWorkflowAggregate(workflowID string) *fes.Aggregate {
	return &fes.Aggregate{
		Id:   workflowID,
		Type: TypeWorkflow,
	}
}

func (wf *Workflow) ApplyEvent(event *fes.Event) error {
	wfEvent, err := events.ParseWorkflow(event.Type)
	if err != nil {
		return err
	}
	switch wfEvent {
	case events.Workflow_WORKFLOW_PARSING_FAILED: // Expects payload: types.Error
		wfErr := &types.Error{}
		err := proto.Unmarshal(event.Data, wfErr)
		if err != nil {
			wfErr.Message = err.Error()
			log.Errorf("failed to unmarshal event: '%v' (%v)", event, err)
		}

		wf.Status.Error = wfErr
		wf.Status.UpdatedAt = event.GetTimestamp()
		wf.Status.Status = types.WorkflowStatus_FAILED
	case events.Workflow_WORKFLOW_CREATED: // Expects payload: types.WorkflowSpec
		spec := &types.WorkflowSpec{}
		err := proto.Unmarshal(event.Data, spec)
		if err != nil {
			return err
		}

		// Ensure that this is the first workflow created event
		if wf.Workflow != nil && wf.Workflow.Spec != nil {
			return ErrIllegalEvent
		}

		// Setup object
		wf.AggregatorMixin = fes.NewAggregatorMixin(wf, *event.Aggregate)
		wf.Workflow = &types.Workflow{
			Metadata: &types.ObjectMetadata{
				Id:        wf.Aggregate().Id,
				CreatedAt: event.GetTimestamp(),
			},
			Spec: spec,
			Status: &types.WorkflowStatus{
				Status:    types.WorkflowStatus_PENDING,
				UpdatedAt: event.GetTimestamp(),
			},
		}
	case events.Workflow_WORKFLOW_PARSED: // Expects payload: types.WorkflowStatus
		status := &types.WorkflowStatus{}
		err := proto.Unmarshal(event.Data, status)
		if err != nil {
			return err
		}
		wf.Status.UpdatedAt = event.GetTimestamp()
		wf.Status.Status = types.WorkflowStatus_READY
		wf.Status.Tasks = status.Tasks
	case events.Workflow_WORKFLOW_DELETED: // Expects payload: nil
		if wf.Workflow == nil || wf.Workflow.Status == nil {
			return ErrIllegalEvent
		}
		wf.Status.Status = types.WorkflowStatus_DELETED
	default:
		log.WithFields(log.Fields{
			"event": event,
		}).Warn("Skipping unimplemented event.")
	}
	return nil
}

func (wf *Workflow) GenericCopy() fes.Aggregator {
	n := &Workflow{
		Workflow: wf.Copy(),
	}
	n.AggregatorMixin = wf.CopyAggregatorMixin(n)
	return n
}

func (wf *Workflow) Copy() *types.Workflow {
	return proto.Clone(wf.Workflow).(*types.Workflow)
}
