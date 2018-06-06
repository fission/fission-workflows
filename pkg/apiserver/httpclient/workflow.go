package httpclient

import (
	"context"
	"net/http"

	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/types"
)

type WorkflowAPI struct {
	baseAPI
}

func NewWorkflowAPI(endpoint string, client http.Client) *WorkflowAPI {
	return &WorkflowAPI{
		baseAPI: baseAPI{
			endpoint: endpoint,
			client:   client,
		},
	}
}

func (api *WorkflowAPI) Create(ctx context.Context, spec *types.WorkflowSpec) (*apiserver.WorkflowIdentifier, error) {
	result := &apiserver.WorkflowIdentifier{}
	err := call(http.MethodPost, api.formatURL("/workflow"), spec, result)
	return result, err
}

func (api *WorkflowAPI) List(ctx context.Context) (*apiserver.SearchWorkflowResponse, error) {
	result := &apiserver.SearchWorkflowResponse{}
	err := call(http.MethodGet, api.formatURL("/workflow"), nil, result)
	return result, err
}

func (api *WorkflowAPI) Get(ctx context.Context, id string) (*types.Workflow, error) {
	result := &types.Workflow{}
	err := call(http.MethodGet, api.formatURL("/workflow/"+id), nil, result)
	return result, err
}

func (api *WorkflowAPI) Delete(ctx context.Context, id string) error {
	err := call(http.MethodDelete, api.formatURL("/workflow/"+id), nil, nil)
	return err
}
