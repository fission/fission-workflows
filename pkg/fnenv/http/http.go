package http

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httputil"
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
	mapper := httpconv.DefaultHTTPMapper.Clone()
	mapper.DefaultHTTPMethod = http.MethodGet
	return &Runtime{
		Client:   http.DefaultClient,
		httpconv: mapper,
	}
}

type Runtime struct {
	Client   *http.Client
	httpconv *httpconv.HTTPMapper
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
	err = r.httpconv.FormatRequest(spec.GetInputs(), req)
	if err != nil {
		return nil, err
	}

	logrus.Infof("HTTP request: %s %v", req.Method, req.URL)
	if logrus.GetLevel() == logrus.DebugLevel {
		fmt.Println("--- HTTP Request ---")
		bs, err := httputil.DumpRequest(req, true)
		if err != nil {
			logrus.Error(err)
		}
		fmt.Println(string(bs))
		fmt.Println("--- HTTP Request end ---")
	}
	resp, err := r.Client.Do(req)
	if err != nil {
		return nil, err
	}
	logrus.Infof("HTTP response: %d - %s", resp.StatusCode, resp.Header.Get("Content-Type"))
	if logrus.GetLevel() == logrus.DebugLevel {
		fmt.Println("--- HTTP Response ---")
		bs, err := httputil.DumpResponse(resp, true)
		if err != nil {
			logrus.Error(err)
		}
		fmt.Println(string(bs))
		fmt.Println("--- HTTP Response end ---")
	}

	output, err := r.httpconv.ParseResponse(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		msg, _ := typedvalues.Unwrap(output)
		return &types.TaskInvocationStatus{
			Status: types.TaskInvocationStatus_FAILED,
			Error: &types.Error{
				Message: fmt.Sprintf("HTTP runtime request error: %v", msg),
			},
		}, nil
	}
	return &types.TaskInvocationStatus{
		Status: types.TaskInvocationStatus_SUCCEEDED,
		Output: output,
	}, nil
}
