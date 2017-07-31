package apiserver

import (
	"github.com/fission/fission-workflow/pkg/api"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/golang/protobuf/ptypes/empty"
	"golang.org/x/net/context"
)

// Events all belong to the same invocation ID, but have different sequence numbers. EventID: <InvocationID>#<sequenceID>
// TODO might need to optimize the bucket use in boltdb
type grpcInvocationApiServer struct {
	api *api.InvocationApi
}

func NewGrpcInvocationApiServer(api *api.InvocationApi) WorkflowInvocationAPIServer {
	return &grpcInvocationApiServer{api}
}

func (gi *grpcInvocationApiServer) Invoke(ctx context.Context, invocation *types.WorkflowInvocationSpec) (*WorkflowInvocationIdentifier, error) {
	eventId, err := gi.api.Invoke(invocation)
	if err != nil {
		return nil, err
	}

	return &WorkflowInvocationIdentifier{eventId}, nil
}

func (gi *grpcInvocationApiServer) Cancel(ctx context.Context, invocationId *WorkflowInvocationIdentifier) (*empty.Empty, error) {
	err := gi.api.Cancel(invocationId.GetId())
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (gi *grpcInvocationApiServer) Get(ctx context.Context, invocationId *WorkflowInvocationIdentifier) (*types.WorkflowInvocationContainer, error) {
	return gi.api.Get(invocationId.GetId())
}

func (gi *grpcInvocationApiServer) List(context.Context, *empty.Empty) (*WorkflowInvocationList, error) {
	invocations, err := gi.api.List("*")
	if err != nil {
		return nil, err
	}
	return &WorkflowInvocationList{invocations}, nil
}
