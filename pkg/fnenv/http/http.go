package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/typedvalues/httpconv"
	"github.com/sirupsen/logrus"
)

var (
	ErrUnsupportedScheme = errors.New("fnenv/http: unsupported scheme")
)

func New() *Runtime {
	return &Runtime{Client: http.DefaultClient}
}

type Runtime struct {
	Client *http.Client
}

// Example: https://us-east1-personal-erwinvaneyk.cloudfunctions.net/helloworld
func (r *Runtime) Resolve(ref types.FnRef) (string, error) {
	if err := types.ValidateFnRef(ref, false); err != nil {
		return "", err
	}
	targetUrl, err := url.Parse(ref.Format())
	if err != nil {
		return "", err
	}
	logrus.Info(targetUrl)
	if targetUrl.Scheme != "http" && targetUrl.Scheme != "https" {
		return "", ErrUnsupportedScheme
	}
	id := targetUrl.String()
	logrus.Infof("Resolved http function %s to %s", ref.ID, id)
	return id, nil
}

func (r *Runtime) Invoke(spec *types.TaskInvocationSpec, opts ...fnenv.InvokeOption) (*types.TaskInvocationStatus, error) {
	cfg := fnenv.ParseInvokeOptions(opts)
	req := (&http.Request{}).WithContext(cfg.Ctx)

	// Parse URL
	fnref := spec.FnRef
	fnUrl, err := url.Parse(fnref.Format())
	if err != nil {
		return nil, err
	}
	req.URL = fnUrl

	// Pass task inputs to HTTP request
	err = httpconv.FormatRequest(spec.GetInputs(), req)
	if err != nil {
		return nil, err
	}
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}

	output, err := httpconv.ParseResponse(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		msg, _ := typedvalues.Format(&output)
		return &types.TaskInvocationStatus{
			Status: types.TaskInvocationStatus_FAILED,
			Error: &types.Error{
				Message: fmt.Sprintf("http function error: %v", msg),
			},
		}, nil
	}
	return &types.TaskInvocationStatus{
		Status: types.TaskInvocationStatus_SUCCEEDED,
		Output: &output,
	}, nil
}
