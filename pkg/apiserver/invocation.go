package apiserver

import (
	"errors"
	"strings"
	"time"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

const (
	invokeSyncTimeout         = time.Duration(10) * time.Minute
	invokeSyncPollingInterval = time.Duration(100) * time.Millisecond
)

type Invocation struct {
	api      *api.Invocation
	wfiCache fes.CacheReader
}

func (gi *Invocation) Validate(ctx context.Context, spec *types.WorkflowInvocationSpec) (*empty.Empty, error) {
	err := validate.WorkflowInvocationSpec(spec)
	if err != nil {
		logrus.Info(strings.Replace(validate.Format(err), "\n", "; ", -1))
		return nil, err
	}
	return &empty.Empty{}, nil
}

func NewInvocation(api *api.Invocation, wfiCache fes.CacheReader) WorkflowInvocationAPIServer {
	return &Invocation{api, wfiCache}
}

func (gi *Invocation) Invoke(ctx context.Context, spec *types.WorkflowInvocationSpec) (*WorkflowInvocationIdentifier, error) {
	eventID, err := gi.api.Invoke(spec)
	if err != nil {
		return nil, err
	}

	return &WorkflowInvocationIdentifier{eventID}, nil
}

func (gi *Invocation) InvokeSync(ctx context.Context, spec *types.WorkflowInvocationSpec) (*types.WorkflowInvocation, error) {
	wfiID, err := gi.api.Invoke(spec)
	if err != nil {
		logrus.Errorf("Failed to invoke workflow: %v", err)
		return nil, err
	}

	timeout, _ := context.WithTimeout(ctx, invokeSyncTimeout)
	var result *types.WorkflowInvocation
	for {
		wi := aggregates.NewWorkflowInvocation(wfiID)
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
			err := gi.api.Cancel(wfiID)
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

func (gi *Invocation) Cancel(ctx context.Context, invocationID *WorkflowInvocationIdentifier) (*empty.Empty, error) {
	err := gi.api.Cancel(invocationID.GetId())
	if err != nil {
		return nil, err
	}

	return &empty.Empty{}, nil
}

func (gi *Invocation) Get(ctx context.Context, invocationID *WorkflowInvocationIdentifier) (*types.WorkflowInvocation, error) {
	wi := aggregates.NewWorkflowInvocation(invocationID.GetId())
	err := gi.wfiCache.Get(wi)
	if err != nil {
		return nil, err
	}
	return wi.WorkflowInvocation, nil
}

func (gi *Invocation) List(context.Context, *empty.Empty) (*WorkflowInvocationList, error) {
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
