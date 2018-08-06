package apiserver

import (
	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/api/aggregates"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

// Workflow is responsible for all functionality related to managing workflows.
type Workflow struct {
	api   *api.Workflow
	cache fes.CacheReader
}

func NewWorkflow(api *api.Workflow, cache fes.CacheReader) *Workflow {
	wf := &Workflow{
		api:   api,
		cache: cache,
	}

	return wf
}

func (ga *Workflow) Create(ctx context.Context, spec *types.WorkflowSpec) (*WorkflowIdentifier, error) {
	id, err := ga.api.Create(spec, api.WithContext(ctx))
	if err != nil {
		return nil, toErrorStatus(err)
	}

	return &WorkflowIdentifier{id}, nil
}

func (ga *Workflow) Get(ctx context.Context, workflowID *WorkflowIdentifier) (*types.Workflow, error) {
	entity := aggregates.NewWorkflow(workflowID.GetId())
	err := ga.cache.Get(entity)
	if err != nil {
		return nil, toErrorStatus(err)
	}
	return entity.Workflow, nil
}

func (ga *Workflow) Delete(ctx context.Context, workflowID *WorkflowIdentifier) (*empty.Empty, error) {
	err := ga.api.Delete(workflowID.GetId())
	if err != nil {
		return nil, toErrorStatus(err)
	}
	return &empty.Empty{}, nil
}

func (ga *Workflow) List(ctx context.Context, req *empty.Empty) (*SearchWorkflowResponse, error) {
	var results []string
	wfs := ga.cache.List()
	for _, result := range wfs {
		results = append(results, result.Id)
	}
	return &SearchWorkflowResponse{results}, nil
}

func (ga *Workflow) Validate(ctx context.Context, spec *types.WorkflowSpec) (*empty.Empty, error) {
	err := validate.WorkflowSpec(spec)
	if err != nil {
		return nil, toErrorStatus(err)
	}
	return &empty.Empty{}, nil
}
