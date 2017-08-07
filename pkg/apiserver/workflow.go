package apiserver

import (
	"errors"

	"github.com/fission/fission-workflow/pkg/api/workflow"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

// TODO move logic to gRPC-less code
type GrpcWorkflowApiServer struct {
	api       *workflow.Api
	validator *workflow.Validator
}

func NewGrpcWorkflowApiServer(api *workflow.Api, validator *workflow.Validator) *GrpcWorkflowApiServer {
	// TODO es: check if available
	wf := &GrpcWorkflowApiServer{
		api:       api,
		validator: validator,
	}

	return wf
}

// TODO validate workflow
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

	return ga.api.Get(id)
}

func (ga *GrpcWorkflowApiServer) List(ctx context.Context, req *empty.Empty) (*SearchWorkflowResponse, error) {
	wfs, err := ga.api.List("*")
	if err != nil {
		return nil, err
	}
	return &SearchWorkflowResponse{wfs}, nil
}

// TODO option to do a dummy parse to validate successful parse
func (ga *GrpcWorkflowApiServer) Validate(ctx context.Context, spec *types.WorkflowSpec) (*empty.Empty, error) {
	err := ga.validator.Validate(spec)
	if err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}
