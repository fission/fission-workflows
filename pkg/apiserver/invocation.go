package apiserver

import (
	"errors"
	"time"

	"github.com/fission/fission-workflows/pkg/api/invocation"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// Events all belong to the same invocation ID, but have different sequence numbers. EventID: <InvocationID>#<sequenceID>
type grpcInvocationApiServer struct {
	api *invocation.Api
}

func NewGrpcInvocationApiServer(api *invocation.Api) WorkflowInvocationAPIServer {
	return &grpcInvocationApiServer{api}
}

// TODO simplify inputs by allowing inputs without type (string vs. { type, value })
func (gi *grpcInvocationApiServer) Invoke(ctx context.Context, spec *types.WorkflowInvocationSpec) (*WorkflowInvocationIdentifier, error) {
	eventId, err := gi.api.Invoke(spec)
	if err != nil {
		return nil, err
	}

	return &WorkflowInvocationIdentifier{eventId}, nil
}

func (gi *grpcInvocationApiServer) InvokeSync(ctx context.Context, spec *types.WorkflowInvocationSpec) (*types.WorkflowInvocation, error) {
	eventId, err := gi.api.Invoke(spec)
	if err != nil {
		return nil, err
	}

	timeout := time.After(time.Duration(60) * time.Second)
	var result *types.WorkflowInvocation
	for {
		wi, err := gi.api.Get(eventId)
		if err != nil {
			logrus.Warn(err)
		}
		if wi != nil && wi.GetStatus() != nil && wi.GetStatus().Status.Finished() {
			result = wi
			break
		}

		select {
		case <-timeout:
			return nil, errors.New("timeout occurred")
		default:
			// TODO temporary shortcut; needs optimizing.
			time.Sleep(time.Duration(1) * time.Second)
		}
	}

	return result, nil
}

func (gi *grpcInvocationApiServer) Cancel(ctx context.Context, invocationId *WorkflowInvocationIdentifier) (*empty.Empty, error) {
	err := gi.api.Cancel(invocationId.GetId())
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (gi *grpcInvocationApiServer) Get(ctx context.Context, invocationId *WorkflowInvocationIdentifier) (*types.WorkflowInvocation, error) {
	return gi.api.Get(invocationId.GetId())
}

func (gi *grpcInvocationApiServer) List(context.Context, *empty.Empty) (*WorkflowInvocationList, error) {
	invocations, err := gi.api.List("*")
	if err != nil {
		return nil, err
	}
	return &WorkflowInvocationList{invocations}, nil
}
