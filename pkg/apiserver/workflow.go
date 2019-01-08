package apiserver

import (
	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/api/aggregates"
	"github.com/fission/fission-workflows/pkg/api/store"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
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
	events, err := ga.backend.Get(fes.Aggregate{
		Id:   md.Id,
		Type: aggregates.TypeWorkflow,
	})
	if err != nil {
		return nil, toErrorStatus(err)
	}
	return &ObjectEvents{
		Metadata: md,
		Events:   events,
	}, nil
}
