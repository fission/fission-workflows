package aggregates

import (
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/gogo/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

const (
	TYPE_WORKFLOW = "workflow"
)

type Workflow struct {
	*fes.AggregatorMixin
	*types.Workflow
}

func NewWorkflow(workflowId string, wi *types.Workflow) *Workflow {
	wia := &Workflow{
		Workflow: wi,
	}

	wia.AggregatorMixin = fes.NewAggregatorMixin(wia, *NewWorkflowAggregate(workflowId))
	return wia
}

func NewWorkflowAggregate(workflowId string) *fes.Aggregate {
	return &fes.Aggregate{
		Id:   workflowId,
		Type: TYPE_WORKFLOW,
	}
}

func (wf *Workflow) ApplyEvent(event *fes.Event) error {
	wfEvent, err := events.ParseWorkflow(event.Type)
	if err != nil {
		return err
	}
	switch wfEvent {
	case events.Workflow_WORKFLOW_PARSING_FAILED:
		wfErr := &types.Error{}
		err := proto.Unmarshal(event.Data, wfErr)
		if err != nil {
			wfErr.Code = "error"
			wfErr.Message = err.Error()
			log.Errorf("failed to unmarshal event: '%v' (%v)", event, err)
		}

		wf.Status.Error = wfErr
		wf.Status.UpdatedAt = event.GetTimestamp()
		wf.Status.Status = types.WorkflowStatus_FAILED
	case events.Workflow_WORKFLOW_CREATED:
		spec := &types.WorkflowSpec{}
		err := proto.Unmarshal(event.Data, spec)
		if err != nil {
			return err
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
				Status:    types.WorkflowStatus_UNKNOWN, // TODO Nest into own state machine maybe
				UpdatedAt: event.GetTimestamp(),
			},
		}
	case events.Workflow_WORKFLOW_PARSED:
		status := &types.WorkflowStatus{}
		err := proto.Unmarshal(event.Data, status)
		if err != nil {
			return err
		}
		wf.Status.UpdatedAt = event.GetTimestamp()
		wf.Status.Status = types.WorkflowStatus_READY
		wf.Status.ResolvedTasks = status.GetResolvedTasks()
	default:
		log.WithFields(log.Fields{
			"event": event,
		}).Warn("Skipping unimplemented event.")
	}
	return nil
}
