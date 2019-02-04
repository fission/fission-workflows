package builtin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fission/fission-workflows/pkg/fnenv/http"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/proto"
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
type FunctionHTTP struct {
	runtime *http.Runtime
}

func NewFunctionHTTP() *FunctionHTTP {
	return &FunctionHTTP{
		runtime: http.New(),
	}
}

func (fn *FunctionHTTP) Invoke(spec *types.TaskInvocationSpec) (*typedvalues.TypedValue, error) {
	// Get the actual url
	targetUrl, err := fn.determineTargetURL(spec.Inputs)
	if err != nil {
		return nil, err
	}
	fnref, err := types.ParseFnRef(targetUrl)
	if err != nil {
		return nil, err
	}
	clonedSpec := proto.Clone(spec).(*types.TaskInvocationSpec)
	t := clonedSpec.GetTask()
	if t == nil {
		t = &types.Task{}
		clonedSpec.Task = t
	}
	ts := t.GetStatus()
	if ts == nil {
		ts = &types.TaskStatus{}
		t.Status = ts
	}
	ts.FnRef = &fnref

	result, err := fn.runtime.Invoke(clonedSpec)
	if err != nil {
		return nil, err
	}
	if result.GetStatus() == types.TaskInvocationStatus_FAILED {
		return nil, result.GetError()
	}
	return result.GetOutput(), nil
}

func (fn *FunctionHTTP) determineTargetURL(inputs map[string]*typedvalues.TypedValue) (string, error) {
	_, tv := getFirstDefinedTypedValue(inputs, HttpInputUrl, types.InputMain)
	if tv == nil {
		return "", errors.New("target URL is required for HTTP function")
	}
	s, err := typedvalues.UnwrapString(tv)
	if err != nil {
		return "", err
	}

	if !strings.HasPrefix(s, "http") {
		s = fmt.Sprintf("%s://%s", httpDefaultProtocol, s)
	}

	return s, err
}
