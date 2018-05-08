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

/*
HttpFunction is a general utility function to perform simple HTTP requests.
It is useful for prototyping and managing low overhead HTTP requests.
To this end it offers basic functionality, such as setting headers, query, method, url, and body inputs.

**Specification**

**input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
url/default     | yes      | string            | URL of the request.
headers         | no       | map[string|string | The action to perform for every element.
content-type    | no       | string            | Force a specific content-type for the request.
method          | no       | string            | HTTP Method of the request. (default: GET)
body            | no       | *                 | The body of the request. (default: application/octet-stream)

Unless the content type is specified explicitly, the workflow engine will infer the content-type based on the body.

**output** (*) the body of the response.

Note: currently you cannot access the metadata of the response.

**Example**

```yaml
# ...
httpExample:
  run: http
  inputs:
    url: http://fission.io
    method: post
    body: "foo"
# ...
```

A complete example of this function can be found in the [httpwhale](../examples/whales/httpwhale.wf.yaml) example.
*/
type FunctionHttp struct{}

func (fn *FunctionHttp) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {
	// Setup request
	contentType := httpconv.DetermineContentTypeInputs(spec.Inputs)
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
