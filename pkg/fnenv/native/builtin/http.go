package builtin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fission/fission-workflows/pkg/fnenv/common/httpconv"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
)

const (
	Http                = "http"
	HttpInputUrl        = "url"
	httpDefaultProtocol = "http"
)

// FunctionHttp supports simple HTTP calls.
type FunctionHttp struct{}

func (fn *FunctionHttp) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {
	// Setup request
	contentType := httpconv.DetermineContentType(spec.Inputs)
	headers := httpconv.FormatHeaders(spec.Inputs)
	method := httpconv.FormatMethod(spec.Inputs)
	// Get the actual url
	uri, err := fn.getUri(spec.Inputs)
	if err != nil {
		return nil, err
	}
	var body io.ReadCloser

	bodyTv, ok := spec.Inputs[types.INPUT_MAIN]
	if ok && bodyTv != nil {
		bs, err := httpconv.FormatBody(*bodyTv, contentType)
		if err != nil {
			return nil, err
		}
		body = ioutil.NopCloser(bytes.NewReader(bs))
	}

	r, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}
	r.Header = headers
	r.Header.Set("Content-Type", contentType)

	// Execute HTTP call
	logrus.Infof("Executing request: %s %s - headers: {%v} - Body: %d", r.Method, r.URL, r.Header, body)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	logrus.Infof("Received response: %s %s - headers: {%v} - Body: %d", resp.Status, resp.Request.URL, resp.Header, resp.ContentLength)

	// Parse output
	output, err := httpconv.ParseBody(resp.Body, r.Header.Get("Content-Type"))
	return &output, err
}

func (fn *FunctionHttp) getUri(inputs map[string]*types.TypedValue) (string, error) {
	_, tv := getFirstDefinedTypedValue(inputs, HttpInputUrl, types.INPUT_MAIN)
	if tv == nil {
		return "", errors.New("target URI is required for HTTP function")
	}
	s, err := typedvalues.FormatString(tv)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(s, "http") {
		s = fmt.Sprintf("%s://%s", httpDefaultProtocol, s)
	}

	return s, err
}
