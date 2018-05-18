package httpclient

import (
	"context"
	"net/http"

	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/version"
)

type AdminApi struct {
	BaseApi
}

func NewAdminApi(endpoint string, client http.Client) *AdminApi {
	return &AdminApi{
		BaseApi: BaseApi{
			endpoint: endpoint,
			client:   client,
		},
	}
}

func (api *AdminApi) Status(ctx context.Context) (*apiserver.Health, error) {
	result := &apiserver.Health{}
	err := call(http.MethodGet, api.formatUrl("/healthz"), nil, result)
	return result, err
}

func (api *AdminApi) Version(ctx context.Context) (*version.Info, error) {
	result := &version.Info{}
	err := call(http.MethodGet, api.formatUrl("/version"), nil, result)
	return result, err
}

func (api *AdminApi) Resume(ctx context.Context) error {
	err := call(http.MethodGet, api.formatUrl("/resume"), nil, nil)
	return err
}

func (api *AdminApi) Halt(ctx context.Context) error {
	err := call(http.MethodGet, api.formatUrl("/halt"), nil, nil)
	return err
}
