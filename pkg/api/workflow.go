package api

import (
	"errors"
	"fmt"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

var (
	ErrWorkflowAlreadyExists = errors.New("workflow already exists")
)

// Workflow contains the API functionality for controlling workflow definitions.
// This includes creating and parsing workflows.
type Workflow struct {
	es       fes.Backend
	resolver fnenv.Resolver
}

// NewWorkflowAPI creates the Workflow API.
func NewWorkflowAPI(esClient fes.Backend, resolver fnenv.Resolver) *Workflow {
	return &Workflow{esClient, resolver}
}

// Create creates a new workflow based on the provided workflowSpec.
// The function either returns the id of the workflow or an error.
// The error can be a validate.Err, proto marshall error, or a fes error.
func (wa *Workflow) Create(workflow *types.WorkflowSpec) (string, error) {
	err := validate.WorkflowSpec(workflow)
	if err != nil {
		return "", err
	}

	// If no id is provided generate an id
	id := workflow.ForceId
	if len(id) == 0 {
		id = fmt.Sprintf("wf-%s", util.UID())
	}
	key := aggregates.NewWorkflowAggregate(id)

	// Check if the workflow already exists
	// TODO improve performance (e.g. caching, exists() option)
	e, _ := wa.es.Get(key)
	if e != nil && len(e) != 0 {
		return id, ErrWorkflowAlreadyExists
	}

	data, err := proto.Marshal(workflow)
	if err != nil {
		return "", err
	}

	err = wa.es.Append(&fes.Event{
		Type:      events.Workflow_WORKFLOW_CREATED.String(),
		Aggregate: key,
		Timestamp: ptypes.TimestampNow(),
		Data:      data,
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

// Delete marks a workflow as deleted, making it unavailable to any future interactions.
// This also means that subsequent invocations for this workflow will fail.
// If the API fails to append the event to the event store, it will return an error.
func (wa *Workflow) Delete(workflowID string) error {
	if len(workflowID) == 0 {
		return validate.NewError("workflowID", errors.New("id should not be empty"))
	}

	return wa.es.Append(&fes.Event{
		Type:      events.Workflow_WORKFLOW_DELETED.String(),
		Aggregate: aggregates.NewWorkflowAggregate(workflowID),
		Timestamp: ptypes.TimestampNow(),
		Hints:     &fes.EventHints{Completed: true},
	})
}

// Parse processes the workflow to resolve any ambiguity.
// Currently, this means that all the function references are resolved to function identifiers. For convenience
// this function returns the new WorkflowStatus. If the API fails to append the event to the event store,
// it will return an error.
func (wa *Workflow) Parse(workflow *types.Workflow) (*types.WorkflowStatus, error) {
	if err := validate.WorkflowSpec(workflow.Spec); err != nil {
		return nil, err
	}

	resolvedFns, err := fnenv.ResolveTasks(wa.resolver, workflow.Spec.Tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve tasks in workflow: %v", err)
	}

	wfStatus := types.NewWorkflowStatus()
	for id, t := range workflow.Spec.Tasks {
		wfStatus.AddTaskStatus(id, &types.TaskStatus{
			UpdatedAt: ptypes.TimestampNow(),
			FnRef:     resolvedFns[t.FunctionRef],
			Status:    types.TaskStatus_READY,
		})
	}

	parsedData, err := proto.Marshal(wfStatus)
	if err != nil {
		return nil, err
	}

	err = wa.es.Append(&fes.Event{
		Type:      events.Workflow_WORKFLOW_PARSED.String(),
		Aggregate: aggregates.NewWorkflowAggregate(workflow.Metadata.Id),
		Timestamp: ptypes.TimestampNow(),
		Data:      parsedData,
	})
	if err != nil {
		return nil, err
	}

	return wfStatus, nil
}
