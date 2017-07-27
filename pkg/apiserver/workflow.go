package apiserver

import (
	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/types"
	"golang.org/x/net/context"
)

// TODO move logic to gRPC-less code
type GrpcWorkflowApiServer struct {
	esClient eventstore.Client
	//projector *projector.WorkflowProjector
}

func NewGrpcWorkflowApiServer(esClient eventstore.Client) *GrpcWorkflowApiServer {
	// TODO es: check if available
	return &GrpcWorkflowApiServer{
		esClient: esClient,
		//projector: projector.NewWorkflowProjector(esClient),
	}
}

// TODO validate workflow
func (ga *GrpcWorkflowApiServer) Create(ctx context.Context, workflow *types.Workflow) (*types.Workflow, error) {
	//workflow.Id = uuid.NewV4().String()
	//
	//b, err := ptypes.MarshalAny(workflow)
	//if err != nil {
	//	return nil, err
	//}
	//
	//event := events.New(eventids.New([]string{"WORKFLOWS"}, workflow.GetId()),
	//	types.WorkflowEvents_WORKFLOW_CREATED.String(), b)
	//
	////updatedEvent, err := ga.esClient.Append(event)
	//if err != nil {
	//	return nil, err
	//}
	//
	////updatedWorkflow, err := projector.GetWorkflowFromEvent(updatedEvent)
	//if err != nil {
	//	return nil, err
	//}

	return nil, nil
}

func (ga *GrpcWorkflowApiServer) Get(ctx context.Context, workflowId *WorkflowIdentifier) (*types.Workflow, error) {
	//	wf, err := ga.projector.GetWorkflow(workflowId.GetId())
	//	if err != nil {
	//		return nil, err
	//	}
	//	if wf == nil {
	//		return nil, status.Errorf(codes.NotFound, "Workflow with id '%s' not found.", workflowId.GetId())
	//	}
	//	return wf, nil
	panic("not implemented")
}

func (ga *GrpcWorkflowApiServer) Search(ctx context.Context, query *SearchWorkflowRequest) (*SearchWorkflowResponse, error) {
	//wfEvents, err := ga.esClient.Events(eventids.NewSubject("WORKFLOWS"))
	//if err != nil {
	//	return nil, err
	//}
	//var results []*types.Workflow
	//for _, wfEvent := range wfEvents {
	//	//wf, err := projector.GetWorkflowFromEvent(wfEvent)
	//	if err != nil {
	//		return nil, err
	//	}
	//	results = append(results, wf)
	//}
	//return &SearchWorkflowResponse{
	//	Data: results,
	//}, nil
	return nil, nil
}
