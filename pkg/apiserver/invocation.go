package apiserver

import (
	"errors"
	"strings"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fnenv/workflows"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

type Invocation struct {
	api      *api.Invocation
	wfiCache fes.CacheReader
	fnenv    *workflows.Runtime
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
	return &Invocation{api, wfiCache, workflows.NewRuntime(api, wfiCache)}
}

func (gi *Invocation) Invoke(ctx context.Context, spec *types.WorkflowInvocationSpec) (*WorkflowInvocationIdentifier, error) {
	eventID, err := gi.api.Invoke(spec)
	if err != nil {
		return nil, err
	}

	return &WorkflowInvocationIdentifier{eventID}, nil
}

func (gi *Invocation) InvokeSync(ctx context.Context, spec *types.WorkflowInvocationSpec) (*types.WorkflowInvocation, error) {
	return gi.fnenv.InvokeWorkflow(spec)
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
