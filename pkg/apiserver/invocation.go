package apiserver

import (
	"errors"

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

// Invocation is responsible for all functionality related to managing invocations.
type Invocation struct {
	api      *api.Invocation
	wfiCache fes.CacheReader
	fnenv    *workflows.Runtime
}

func (gi *Invocation) Validate(ctx context.Context, spec *types.WorkflowInvocationSpec) (*empty.Empty, error) {
	err := validate.WorkflowInvocationSpec(spec)
	if err != nil {
		return nil, toErrorStatus(err)
	}
	return &empty.Empty{}, nil
}

func NewInvocation(api *api.Invocation, wfiCache fes.CacheReader) WorkflowInvocationAPIServer {
	return &Invocation{api, wfiCache, workflows.NewRuntime(api, wfiCache)}
}

func (gi *Invocation) Invoke(ctx context.Context, spec *types.WorkflowInvocationSpec) (*WorkflowInvocationIdentifier, error) {
	eventID, err := gi.api.Invoke(spec)
	if err != nil {
		return nil, toErrorStatus(err)
	}

	return &WorkflowInvocationIdentifier{eventID}, nil
}

func (gi *Invocation) InvokeSync(ctx context.Context, spec *types.WorkflowInvocationSpec) (*types.WorkflowInvocation, error) {
	wfi, err := gi.fnenv.InvokeWorkflow(ctx, spec)
	if err != nil {
		return nil, toErrorStatus(err)
	}
	return wfi, nil
}

func (gi *Invocation) Cancel(ctx context.Context, invocationID *WorkflowInvocationIdentifier) (*empty.Empty, error) {
	err := gi.api.Cancel(invocationID.GetId())
	if err != nil {
		return nil, toErrorStatus(err)
	}

	return &empty.Empty{}, nil
}

func (gi *Invocation) Get(ctx context.Context, invocationID *WorkflowInvocationIdentifier) (*types.WorkflowInvocation, error) {
	wi := aggregates.NewWorkflowInvocation(invocationID.GetId())
	err := gi.wfiCache.Get(wi)
	if err != nil {
		return nil, toErrorStatus(err)
	}
	return wi.WorkflowInvocation, nil
}

func (gi *Invocation) List(ctx context.Context, query *InvocationListQuery) (*WorkflowInvocationList, error) {
	var invocations []string
	as := gi.wfiCache.List()
	for _, aggregate := range as {
		if aggregate.Type != aggregates.TypeWorkflowInvocation {
			return nil, toErrorStatus(errors.New("invalid type in invocation cache"))
		}

		if len(query.Workflows) > 0 {
			// TODO make more efficient (by moving list queries to cache)
			entity, err := gi.wfiCache.GetAggregate(aggregate)
			if err != nil {
				logrus.Error("List: failed to fetch %v from cache: %v", aggregate, err)
				continue
			}
			wfi := entity.(*aggregates.WorkflowInvocation)
			if !contains(query.Workflows, wfi.GetSpec().GetWorkflowId()) {
				continue
			}
		}

		invocations = append(invocations, aggregate.Id)
	}
	return &WorkflowInvocationList{invocations}, nil
}

func contains(haystack []string, needle string) bool {
	for i := 0; i < len(haystack); i++ {
		if haystack[i] == needle {
			return true
		}
	}
	return false
}
