package apiserver

import (
	"errors"

	"github.com/fission/fission-workflow/pkg/api/workflow"
	"github.com/fission/fission-workflow/pkg/projector/project"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

type GrpcWorkflowApiServer struct {
	api       *workflow.Api
	validator *workflow.Validator
	projector project.WorkflowProjector
}

func NewGrpcWorkflowApiServer(api *workflow.Api, validator *workflow.Validator, projector project.WorkflowProjector) *GrpcWorkflowApiServer {
	wf := &GrpcWorkflowApiServer{
		api:       api,
		validator: validator,
		projector: projector,
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
		return nil, errors.New("No id provided")
	}

	return ga.projector.Get(id)
}

func (ga *GrpcWorkflowApiServer) List(ctx context.Context, req *empty.Empty) (*SearchWorkflowResponse, error) {
	wfs, err := ga.projector.List("*")
	if err != nil {
		return nil, err
	}
	return &SearchWorkflowResponse{wfs}, nil
}

func (ga *GrpcWorkflowApiServer) Validate(ctx context.Context, spec *types.WorkflowSpec) (*empty.Empty, error) {
	err := ga.validator.Validate(spec)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}
