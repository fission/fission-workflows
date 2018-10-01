package apiserver

import (
	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/api/aggregates"
	"github.com/fission/fission-workflows/pkg/api/store"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/fnenv/workflows"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
)

// Invocation is responsible for all functionality related to managing store.
type Invocation struct {
	api   *api.Invocation
	store *store.Invocations
	fnenv *workflows.Runtime
}

func NewInvocation(api *api.Invocation, store *store.Invocations) WorkflowInvocationAPIServer {
	return &Invocation{
		api:   api,
		store: store,
		fnenv: workflows.NewRuntime(api, store),
	}
}

func (gi *Invocation) Validate(ctx context.Context, spec *types.WorkflowInvocationSpec) (*empty.Empty, error) {
	err := validate.WorkflowInvocationSpec(spec)
	if err != nil {
		return nil, toErrorStatus(err)
	}
	return &empty.Empty{}, nil
}

func (gi *Invocation) Invoke(ctx context.Context, spec *types.WorkflowInvocationSpec) (*types.ObjectMetadata, error) {
	eventID, err := gi.api.Invoke(spec, api.WithContext(ctx))
	if err != nil {
		return nil, toErrorStatus(err)
	}

	return &types.ObjectMetadata{Id: eventID}, nil
}

func (gi *Invocation) InvokeSync(ctx context.Context, spec *types.WorkflowInvocationSpec) (*types.WorkflowInvocation, error) {
	wfi, err := gi.fnenv.InvokeWorkflow(spec, fnenv.WithContext(ctx))
	if err != nil {
		return nil, toErrorStatus(err)
	}
	return wfi, nil
}

func (gi *Invocation) Cancel(ctx context.Context, objectMetadata *types.ObjectMetadata) (*empty.Empty, error) {
	err := gi.api.Cancel(objectMetadata.GetId())
	if err != nil {
		return nil, toErrorStatus(err)
	}

	return &empty.Empty{}, nil
}

func (gi *Invocation) Get(ctx context.Context, objectMetadata *types.ObjectMetadata) (*types.WorkflowInvocation, error) {
	wi, err := gi.store.GetInvocation(objectMetadata.GetId())
	if err != nil {
		return nil, toErrorStatus(err)
	}
	return wi, nil
}

func (gi *Invocation) List(ctx context.Context, query *InvocationListQuery) (*WorkflowInvocationList, error) {
	var invocations []string
	as := gi.store.List()
	for _, aggregate := range as {
		if aggregate.Type != aggregates.TypeWorkflowInvocation {
			logrus.Errorf("Invalid type in invocation store: %v", aggregate.Format())
			continue
		}

		if len(query.Workflows) > 0 {
			// TODO make more efficient (by moving list queries to store)
			entity, err := gi.store.GetAggregate(aggregate)
			if err != nil {
				logrus.Errorf("List: failed to fetch %v from store: %v", aggregate, err)
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

func (gi *Invocation) AddTask(ctx context.Context, req *AddTaskRequest) (*empty.Empty, error) {
	invocation, err := gi.store.GetInvocation(req.GetInvocationID())
	if err != nil {
		return nil, toErrorStatus(err)
	}
	if err := gi.api.AddTask(invocation.ID(), req.Task); err != nil {
		return nil, err
	}
	return &empty.Empty{}, nil
}

func contains(haystack []string, needle string) bool {
	for i := 0; i < len(haystack); i++ {
		if haystack[i] == needle {
			return true
		}
	}
	return false
}
