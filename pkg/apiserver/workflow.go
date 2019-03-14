package apiserver

import (
	"errors"
	"time"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/api/projectors"
	"github.com/fission/fission-workflows/pkg/api/store"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

const (
	CreateSyncPollInterval = 100 * time.Millisecond
)

// Workflow is responsible for all functionality related to managing workflows.
type Workflow struct {
	api     *api.Workflow
	store   *store.Workflows
	backend fes.Backend
}

func NewWorkflow(api *api.Workflow, store *store.Workflows, backend fes.Backend) *Workflow {
	return &Workflow{
		api:     api,
		store:   store,
		backend: backend,
	}
}

func (ga *Workflow) Create(ctx context.Context, spec *types.WorkflowSpec) (*types.ObjectMetadata, error) {
	id, err := ga.api.Create(spec, api.WithContext(ctx))
	if err != nil {
		return nil, toErrorStatus(err)
	}

	return &types.ObjectMetadata{Id: id}, nil
}

func (ga *Workflow) CreateSync(ctx context.Context, spec *types.WorkflowSpec) (*types.Workflow, error) {
	// Start workflow create
	metadata, err := ga.Create(ctx, spec)
	if err != nil {
		return nil, err
	}

	// Poll for completed state of the workflow.
	ticker := time.NewTicker(CreateSyncPollInterval)
	var lastWorkflowErr error
	for {
		select {
		case <-ctx.Done():
			if lastWorkflowErr == nil {
				lastWorkflowErr = ctx.Err()
			}
			return nil, toErrorStatus(lastWorkflowErr)
		case <-ticker.C:
			// Fetch the current workflow from the workflows store.
			wf, err := ga.store.GetWorkflow(metadata.GetId())
			if err != nil {
				continue
			}

			// Decide if to wait further based on the state of the workflow.
			switch wf.GetStatus().GetStatus() {
			case types.WorkflowStatus_READY:
				return wf, nil
			case types.WorkflowStatus_DELETED:
				return nil, toErrorStatus(errors.New("workflow was deleted"))
			case types.WorkflowStatus_QUEUED:
				continue
			case types.WorkflowStatus_FAILED:
				lastWorkflowErr = wf.GetStatus().GetError()
				continue
			}
		}
	}
}

func (ga *Workflow) Get(ctx context.Context, workflowID *types.ObjectMetadata) (*types.Workflow, error) {
	wf, err := ga.store.GetWorkflow(workflowID.GetId())
	if err != nil {
		return nil, toErrorStatus(err)
	}
	return wf, nil
}

func (ga *Workflow) Delete(ctx context.Context, workflowID *types.ObjectMetadata) (*empty.Empty, error) {
	err := ga.api.Delete(workflowID.GetId())
	if err != nil {
		return nil, toErrorStatus(err)
	}
	return &empty.Empty{}, nil
}

func (ga *Workflow) List(ctx context.Context, req *empty.Empty) (*WorkflowList, error) {
	var results []string
	wfs := ga.store.List()
	for _, result := range wfs {
		results = append(results, result.Id)
	}
	return &WorkflowList{results}, nil
}

func (ga *Workflow) Validate(ctx context.Context, spec *types.WorkflowSpec) (*empty.Empty, error) {
	err := validate.WorkflowSpec(spec)
	if err != nil {
		return nil, toErrorStatus(err)
	}
	return &empty.Empty{}, nil
}

func (ga *Workflow) Events(ctx context.Context, md *types.ObjectMetadata) (*ObjectEvents, error) {
	events, err := ga.backend.Get(projectors.NewWorkflowAggregate(md.Id))
	if err != nil {
		return nil, toErrorStatus(err)
	}
	return &ObjectEvents{
		Metadata: md,
		Events:   events,
	}, nil
}
