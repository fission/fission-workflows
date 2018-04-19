package builtin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
)

const (
	Http             = "http"
	HttpInputMethod  = "method"
	HttpInputUri     = "uri"
	HttpInputBody    = "body"
	HttpInputHeaders = "headers"

	httpDefaultProtocol    = "http"
	httpDefaultContentType = "application/json"
)

// FunctionHttp supports simple HTTP calls.
type FunctionHttp struct{}

func (fn *FunctionHttp) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {
	method, err := fn.getMethod(spec.Inputs)
	uri, err := fn.getUri(spec.Inputs)
	body, err := fn.getBody(spec.Inputs)
	headers, err := fn.getHeaders(spec.Inputs)
	if err != nil {
		return nil, err
	}

	r, err := http.NewRequest(method, uri, body)
	if err != nil {
		return nil, err
	}
	r.Header = headers
	// TODO support other content types
	if body != nil {
		r.Header.Set("Content-Type", httpDefaultContentType)
	}
	logrus.Infof("Executing request: %s %s - headers: {%v} - Body: %d", r.Method, r.URL, r.Header, body)
	resp, err := http.DefaultClient.Do(r)
	if err != nil {
		return nil, err
	}
	logrus.Infof("Received response: %s %s - headers: {%v} - Body: %d", resp.Status, resp.Request.URL, resp.Header, resp.ContentLength)

	return parseBody(resp)
}

func (fn *FunctionHttp) getMethod(inputs map[string]*types.TypedValue) (string, error) {
	// Verify and get condition
	tv, ok := inputs[HttpInputMethod]
	if !ok {
		return http.MethodGet, nil
	}
	method, err := typedvalues.FormatString(tv)
	if err != nil {
		return "", err
	}
	return strings.ToUpper(method), nil
}

func (fn *FunctionHttp) getUri(inputs map[string]*types.TypedValue) (string, error) {
	_, tv := getFirstDefinedTypedValue(inputs, HttpInputUri, types.INPUT_MAIN)
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

func (fn *FunctionHttp) getHeaders(inputs map[string]*types.TypedValue) (http.Header, error) {
	hs := http.Header{}
	tv, ok := inputs[HttpInputHeaders]
	if !ok {
		return hs, nil
	}

	m, err := typedvalues.FormatMap(tv)
	if err != nil {
		return nil, err
	}

	for k, v := range m {
		s, ok := v.(string)
		if !ok {
			continue
		}
		hs.Set(k, s)
	}

	return hs, nil
}

func (fn *FunctionHttp) getBody(inputs map[string]*types.TypedValue) (io.ReadCloser, error) {
	var input []byte
	_, mainInput := getFirstDefinedTypedValue(inputs, HttpInputBody)
	if mainInput != nil {
		// TODO ensure that it is a byte-representation 1-1 of actual value not the representation in TypedValue
		input = mainInput.Value
	}

	return ioutil.NopCloser(bytes.NewReader(input)), nil
}

// parseBody maps the body from a request to the "main" key in the target map
func parseBody(resp *http.Response) (*types.TypedValue, error) {
	contentType := resp.Header.Get("Content-Type")
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("failed to read body")
	}
	defer resp.Body.Close()

	var i interface{} = bs
	// TODO fix this, remove the hardcoded JSON transform
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/json") {
		err = json.Unmarshal(bs, &i)
		if err != nil {
			logrus.WithField("body", len(bs)).Infof("Input is not json: %v", err)
			i = bs
		}
	}

	parsedInput, err := typedvalues.Parse(i)
	if err != nil {
		logrus.Errorf("Failed to parse body: %v", err)
		return parsedInput, errors.New("failed to parse body")
	}
	return parsedInput, nil
}
