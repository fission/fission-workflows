package httpclient

import (
	"context"
	"net/http"

	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/types"
)

type InvocationApi struct {
	BaseApi
}

func NewInvocationApi(endpoint string, client http.Client) *InvocationApi {
	return &InvocationApi{
		BaseApi: BaseApi{
			endpoint: endpoint,
			client:   client,
		},
	}
}

func (api *InvocationApi) Invoke(ctx context.Context, spec *types.WorkflowInvocationSpec) (*apiserver.
	WorkflowInvocationIdentifier, error) {
	result := &apiserver.WorkflowInvocationIdentifier{}
	err := call(http.MethodPost, api.formatUrl("/invocation"), spec, result)
	return result, err
}

func (api *InvocationApi) InvokeSync(ctx context.Context, spec *types.WorkflowInvocationSpec) (*types.
	WorkflowInvocation, error) {
	result := &types.WorkflowInvocation{}
	err := call(http.MethodPost, api.formatUrl("/invocation/sync"), spec, result)
	return result, err
}

func (api *InvocationApi) Cancel(ctx context.Context, id string) error {
	return call(http.MethodDelete, api.formatUrl("/invocation/"+id), nil, nil)
}

func (api *InvocationApi) List(ctx context.Context) (*apiserver.WorkflowInvocationList, error) {
	result := &apiserver.WorkflowInvocationList{}
	err := call(http.MethodGet, api.formatUrl("/invocation"), nil, result)
	return result, err
}

func (api *InvocationApi) Get(ctx context.Context, id string) (*types.WorkflowInvocation, error) {
	result := &types.WorkflowInvocation{}
	err := call(http.MethodGet, api.formatUrl("/invocation/"+id), nil, result)
	return result, err
}

func (api *InvocationApi) Validate(ctx context.Context, spec *types.WorkflowInvocationSpec) error {
	return call(http.MethodPost, api.formatUrl("/invocation/validate"), spec, nil)
}
