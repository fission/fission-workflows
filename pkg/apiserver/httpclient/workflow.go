package httpclient

import (
	"context"
	"net/http"

	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/types"
)

type WorkflowApi struct {
	BaseApi
}

func NewWorkflowApi(endpoint string, client http.Client) *WorkflowApi {
	return &WorkflowApi{
		BaseApi: BaseApi{
			endpoint: endpoint,
			client:   client,
		},
	}
}

func (api *WorkflowApi) Create(ctx context.Context, spec *types.WorkflowSpec) (*apiserver.WorkflowIdentifier, error) {
	result := &apiserver.WorkflowIdentifier{}
	err := call(http.MethodPost, api.formatUrl("/workflow"), spec, result)
	return result, err
}

func (api *WorkflowApi) List(ctx context.Context) (*apiserver.SearchWorkflowResponse, error) {
	result := &apiserver.SearchWorkflowResponse{}
	err := call(http.MethodGet, api.formatUrl("/workflow"), nil, result)
	return result, err
}

func (api *WorkflowApi) Get(ctx context.Context, id string) (*types.Workflow, error) {
	result := &types.Workflow{}
	err := call(http.MethodGet, api.formatUrl("/workflow/"+id), nil, result)
	return result, err
}

func (api *WorkflowApi) Delete(ctx context.Context, id string) error {
	err := call(http.MethodDelete, api.formatUrl("/workflow/"+id), nil, nil)
	return err
}
