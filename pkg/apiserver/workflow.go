package apiserver

import (
	"github.com/fission/fission-workflow/pkg/api"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// TODO move logic to gRPC-less code
type GrpcWorkflowApiServer struct {
	api *api.WorkflowApi
}

func NewGrpcWorkflowApiServer(api *api.WorkflowApi) *GrpcWorkflowApiServer {
	// TODO es: check if available
	wf := &GrpcWorkflowApiServer{
		api: api,
	}

	err := wf.api.Projector.Watch("workflows.>")
	if err != nil {
		logrus.Warnf("Failed to watch for workflows, because '%v'.", err)
	}

	return wf
}

// TODO validate workflow
func (ga *GrpcWorkflowApiServer) Create(ctx context.Context, wf *types.WorkflowSpec) (*WorkflowIdentifier, error) {
	id, err := ga.api.Create(wf)
	if err != nil {
		return nil, err
	}

	return &WorkflowIdentifier{id}, nil
}

func (ga *GrpcWorkflowApiServer) Get(ctx context.Context, workflowId *WorkflowIdentifier) (*types.Workflow, error) {
	return ga.api.Get(workflowId.GetId())
}

func (ga *GrpcWorkflowApiServer) List(ctx context.Context, req *types.Empty) (*SearchWorkflowResponse, error) {
	wfs, err := ga.api.List("*")
	if err != nil {
		return nil, err
	}
	return &SearchWorkflowResponse{wfs}, nil
}
