package httpclient

import (
	"context"
	"net/http"

	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/version"
)

type AdminAPI struct {
	baseAPI
}

func NewAdminAPI(endpoint string, client http.Client) *AdminAPI {
	return &AdminAPI{
		baseAPI: baseAPI{
			endpoint: endpoint,
			client:   client,
		},
	}
}

func (api *AdminAPI) Status(ctx context.Context) (*apiserver.Health, error) {
	result := &apiserver.Health{}
	err := call(http.MethodGet, api.formatURL("/healthz"), nil, result)
	return result, err
}

func (api *AdminAPI) Version(ctx context.Context) (*version.Info, error) {
	result := &version.Info{}
	err := call(http.MethodGet, api.formatURL("/version"), nil, result)
	return result, err
}
