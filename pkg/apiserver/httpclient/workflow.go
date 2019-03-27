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

func (api *WorkflowAPI) Create(ctx context.Context, spec *types.WorkflowSpec) (*types.ObjectMetadata, error) {
	result := &types.ObjectMetadata{}
	err := callWithJSON(ctx, http.MethodPost, api.formatURL("/workflow"), spec, result)
	return result, err
}

func (api *WorkflowAPI) CreateSync(ctx context.Context, spec *types.WorkflowSpec) (*types.Workflow, error) {
	wf := &types.Workflow{}
	err := callWithJSON(ctx, http.MethodPost, api.formatURL("/workflow/sync"), spec, wf)
	return wf, err
}

func (api *WorkflowAPI) List(ctx context.Context) (*apiserver.WorkflowList, error) {
	result := &apiserver.WorkflowList{}
	err := callWithJSON(ctx, http.MethodGet, api.formatURL("/workflow"), nil, result)
	return result, err
}

func (api *WorkflowAPI) Get(ctx context.Context, id string) (*types.Workflow, error) {
	result := &types.Workflow{}
	err := callWithJSON(ctx, http.MethodGet, api.formatURL("/workflow/"+id), nil, result)
	return result, err
}

func (api *WorkflowAPI) Delete(ctx context.Context, id string) error {
	err := callWithJSON(ctx, http.MethodDelete, api.formatURL("/workflow/"+id), nil, nil)
	return err
}

func (api *WorkflowAPI) Events(ctx context.Context, id string) (*apiserver.ObjectEvents, error) {
	result := &apiserver.ObjectEvents{}
	err := callWithJSON(ctx, http.MethodGet, api.formatURL("/workflow/"+id+"/events"), nil, result)
	return result, err
}
