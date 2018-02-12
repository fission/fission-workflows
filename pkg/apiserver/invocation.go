package apiserver

import (
	"errors"
	"time"

	"github.com/fission/fission-workflows/pkg/api/invocation"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	invokeSyncTimeout         = time.Duration(10) * time.Minute
	invokeSyncPollingInterval = time.Duration(100) * time.Millisecond
)

type grpcInvocationApiServer struct {
	api      *invocation.Api
	wfiCache fes.CacheReader
}

func NewGrpcInvocationApiServer(api *invocation.Api, wfiCache fes.CacheReader) WorkflowInvocationAPIServer {
	return &grpcInvocationApiServer{api, wfiCache}
}

func (gi *grpcInvocationApiServer) Invoke(ctx context.Context, spec *types.WorkflowInvocationSpec) (*WorkflowInvocationIdentifier, error) {
	eventId, err := gi.api.Invoke(spec)
	if err != nil {
		return nil, err
	}

	return &WorkflowInvocationIdentifier{eventId}, nil
}

func (gi *grpcInvocationApiServer) InvokeSync(ctx context.Context, spec *types.WorkflowInvocationSpec) (*types.WorkflowInvocation, error) {
	wfiId, err := gi.api.Invoke(spec)
	if err != nil {
		logrus.Errorf("Failed to invoke workflow: %v", err)
		return nil, err
	}

	timeout, _ := context.WithTimeout(ctx, invokeSyncTimeout)
	var result *types.WorkflowInvocation
	for {
		wi := aggregates.NewWorkflowInvocation(wfiId, &types.WorkflowInvocation{})
		err := gi.wfiCache.Get(wi)
		if err != nil {
			logrus.Warnf("Failed to get workflow invocation from cache: %v", err)
		}
		if wi != nil && wi.GetStatus() != nil && wi.GetStatus().Finished() {
			result = wi.WorkflowInvocation
			break
		}

		select {
		case <-timeout.Done():
			err := gi.api.Cancel(wfiId)
			if err != nil {
				logrus.Errorf("Failed to cancel workflow invocation: %v", err)
			}
			return nil, errors.New("timeout occurred")
		default:
			// TODO polling is a temporary shortcut; needs optimizing.
			time.Sleep(invokeSyncPollingInterval)
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
	wi := aggregates.NewWorkflowInvocation(invocationId.GetId(), &types.WorkflowInvocation{})
	err := gi.wfiCache.Get(wi)
	if err != nil {
		return nil, err
	}
	return wi.WorkflowInvocation, nil
}

func (gi *grpcInvocationApiServer) List(context.Context, *empty.Empty) (*WorkflowInvocationList, error) {
	var invocations []string
	as := gi.wfiCache.List()
	for _, a := range as {
		if a.Type != aggregates.TypeWorkflowInvocation {
			return nil, errors.New("invalid type in invocation cache")
		}

		invocations = append(invocations, a.Id)
	}
	return &WorkflowInvocationList{invocations}, nil
}
