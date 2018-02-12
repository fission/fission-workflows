package apiserver

import (
	"errors"

	"github.com/fission/fission-workflows/pkg/api/workflow"
	"github.com/fission/fission-workflows/pkg/api/workflow/parse"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

type GrpcWorkflowApiServer struct {
	api       *workflow.Api
	validator *parse.Validator
	cache     fes.CacheReader
}

func NewGrpcWorkflowApiServer(api *workflow.Api, validator *parse.Validator, cache fes.CacheReader) *GrpcWorkflowApiServer {
	wf := &GrpcWorkflowApiServer{
		api:       api,
		validator: validator,
		cache:     cache,
	}

	return wf
}

func (ga *GrpcWorkflowApiServer) Create(ctx context.Context, wf *types.WorkflowSpec) (*WorkflowIdentifier, error) {
	err := ga.validator.Validate(wf)
	if err != nil {
		return nil, err
	}

	id, err := ga.api.Create(wf)
	if err != nil {
		return nil, err
	}

	return &WorkflowIdentifier{id}, nil
}

func (ga *GrpcWorkflowApiServer) Get(ctx context.Context, workflowId *WorkflowIdentifier) (*types.Workflow, error) {
	id := workflowId.GetId()
	if len(id) == 0 {
		return nil, errors.New("no id provided")
	}

	entity := aggregates.NewWorkflow(id, nil)
	err := ga.cache.Get(entity)
	if err != nil {
		return nil, err
	}
	return entity.Workflow, nil
}

func (ga *GrpcWorkflowApiServer) List(ctx context.Context, req *empty.Empty) (*SearchWorkflowResponse, error) {
	var results []string
	wfs := ga.cache.List()
	for _, result := range wfs {
		results = append(results, result.Id)
	}
	return &SearchWorkflowResponse{results}, nil
}

func (ga *GrpcWorkflowApiServer) Validate(ctx context.Context, spec *types.WorkflowSpec) (*empty.Empty, error) {
	err := ga.validator.Validate(spec)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}
