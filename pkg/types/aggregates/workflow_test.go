package aggregates

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestWorkflow_ApplyEventCreated(t *testing.T) {
	wf := NewWorkflow("wf-1")
	spec := types.NewWorkflowSpec().
		SetOutput("task1").
		AddTask("task1", types.NewTaskSpec("someFn"))
	spec.ForceId = "forcedId"
	data, err := proto.Marshal(spec)
	assert.NoError(t, err)

	a := wf.Aggregate()
	events := []*fes.Event{
		fes.NewEvent("1", events.Workflow_WORKFLOW_CREATED.String(), a, data),
	}

	err = ApplyEvents(wf, events...)
	assert.NoError(t, err)
	assert.Equal(t, spec, wf.Spec)
	assert.Equal(t, types.WorkflowStatus_PENDING, wf.Status.Status)
}

func TestWorkflow_ApplyEventDeleted(t *testing.T) {
	wf := NewWorkflow("wf-1")
	spec := types.NewWorkflowSpec().
		SetOutput("task1").
		AddTask("task1", types.NewTaskSpec("someFn"))
	spec.ForceId = "forcedId"
	data, err := proto.Marshal(spec)
	assert.NoError(t, err)

	a := wf.Aggregate()
	wfEvents := []*fes.Event{
		fes.NewEvent("1", events.Workflow_WORKFLOW_CREATED.String(), a, data),
		fes.NewEvent("2", events.Workflow_WORKFLOW_DELETED.String(), a, nil),
	}

	err = ApplyEvents(wf, wfEvents...)
	assert.NoError(t, err)
	assert.Equal(t, spec, wf.Spec)
	assert.Equal(t, types.WorkflowStatus_DELETED, wf.Status.Status)
}

func TestWorkflow_ApplyEventDeletedNonExistent(t *testing.T) {
	wf := NewWorkflow("wf-1")

	a := wf.Aggregate()
	wfEvents := []*fes.Event{
		fes.NewEvent("2", events.Workflow_WORKFLOW_DELETED.String(), a, nil),
	}

	err := ApplyEvents(wf, wfEvents...)
	assert.Equal(t, ErrIllegalEvent, err)
}
