package httpclient

import (
	"context"
	"net/http"

	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/types"
)

type InvocationAPI struct {
	baseAPI
}

func NewInvocationAPI(endpoint string, client http.Client) *InvocationAPI {
	return &InvocationAPI{
		baseAPI: baseAPI{
			endpoint: endpoint,
			client:   client,
		},
	}
}

func (api *InvocationAPI) Invoke(ctx context.Context, spec *types.WorkflowInvocationSpec) (*apiserver.
	WorkflowInvocationIdentifier, error) {
	result := &apiserver.WorkflowInvocationIdentifier{}
	err := call(http.MethodPost, api.formatURL("/invocation"), spec, result)
	return result, err
}

func (api *InvocationAPI) InvokeSync(ctx context.Context, spec *types.WorkflowInvocationSpec) (*types.
	WorkflowInvocation, error) {
	result := &types.WorkflowInvocation{}
	err := call(http.MethodPost, api.formatURL("/invocation/sync"), spec, result)
	return result, err
}

func (api *InvocationAPI) Cancel(ctx context.Context, id string) error {
	return call(http.MethodDelete, api.formatURL("/invocation/"+id), nil, nil)
}

func (api *InvocationAPI) List(ctx context.Context) (*apiserver.WorkflowInvocationList, error) {
	result := &apiserver.WorkflowInvocationList{}
	err := call(http.MethodGet, api.formatURL("/invocation"), nil, result)
	return result, err
}

func (api *InvocationAPI) Get(ctx context.Context, id string) (*types.WorkflowInvocation, error) {
	result := &types.WorkflowInvocation{}
	err := call(http.MethodGet, api.formatURL("/invocation/"+id), nil, result)
	return result, err
}

func (api *InvocationAPI) Validate(ctx context.Context, spec *types.WorkflowInvocationSpec) error {
	return call(http.MethodPost, api.formatURL("/invocation/validate"), spec, nil)
}
