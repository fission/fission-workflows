package fission

import (
	"bytes"
	"fmt"
	"net/http"

	"encoding/json"
	"io/ioutil"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	executor "github.com/fission/fission/executor/client"
	"github.com/sirupsen/logrus"

	"strings"

	"github.com/fission/fission/router"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
	"net/url"
)

// FunctionEnv adapts the Fission platform to the function execution runtime.
type FunctionEnv struct {
	executor *executor.Client
}

const (
	InputBody         = "body" // or 'default'
	InputHttpMethod   = "http_method"
	InputHeaderPrefix = "header_"
	InputQueryPrefix  = "query_"
	InputContentType  = "content_type" // to force the content type

	defaultHttpMethod  = http.MethodPost
	defaultProtocol    = "http"
	defaultContentType = "text/plain"
)

func NewFunctionEnv(executor *executor.Client) *FunctionEnv {
	return &FunctionEnv{
		executor: executor,
	}
}

func (fe *FunctionEnv) Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocationStatus, error) {
	meta := &metav1.ObjectMeta{
		Name:      spec.GetType().GetSrc(),
		UID:       k8stypes.UID(spec.GetType().GetResolved()),
		Namespace: metav1.NamespaceDefault,
	}
	logrus.WithFields(logrus.Fields{
		"name": meta.Name,
		"UID":  meta.UID,
		"ns":   meta.Namespace,
	}).Info("Invoking Fission function.")

	// Get reqUrl
	serviceUrl, err := fe.executor.GetServiceForFunction(meta)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":  err,
			"meta": meta,
		}).Error("Fission function could not be found!")
		return nil, err
	}
	rawUrl := fmt.Sprintf("%s://%s", defaultProtocol, serviceUrl)
	reqUrl, err := url.Parse(rawUrl)
	if err != nil {
		logrus.Errorf("Failed to parse url: '%v'", rawUrl)
		panic(err)
	}

	// Map body parameter
	var input []byte
	mainInput, _ := getFirstDefinedTypedValue(spec.Inputs, types.INPUT_MAIN, InputBody)
	var short string
	if mainInput != nil {
		input = mainInput.Value
		short = mainInput.Short()
	}

	r := bytes.NewReader(input)
	logrus.Infof("[request][body]: %v", short)

	// Map HTTP method
	httpMethod := httpMethod(spec.Inputs, defaultHttpMethod)
	logrus.Infof("Using HTTP method: %v", httpMethod)

	// Map headers
	headers := headers(spec.Inputs)

	// Determine ContentType
	reqContentType := contentType(spec.Inputs, mainInput, defaultContentType)

	// Map query and add to url
	query := query(spec.Inputs)
	reqUrl.RawQuery = query.Encode()

	// Construct request and add body
	req, err := http.NewRequest(httpMethod, reqUrl.String(), nil)
	if err != nil {
		panic(fmt.Errorf("failed to make request for '%s': %v", serviceUrl, err))
	}

	// Add body
	req.Body = ioutil.NopCloser(r)
	defer req.Body.Close()

	// Add parameters normally added by Fission
	router.MetadataToHeaders(router.HEADERS_FISSION_FUNCTION_PREFIX, meta, req)

	// Set headers
	for k, v := range headers {
		req.Header.Set(k, v)
		// TODO check that no special headers are overwritten
	}
	req.Header.Set("Content-Type", reqContentType)

	// Perform request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error for reqUrl '%s': %v", serviceUrl, err)
	}

	// Parse output
	output := toTypedValue(resp)

	logrus.Infof("[%s][Content-Type]: %v ", meta.Name, resp.Header.Get("Content-Type"))
	logrus.Infof("[%s][output]: %v", meta.Name, output.Short())
	logrus.Infof("[%s][status]: %v", meta.Name, resp.StatusCode)

	// Determine status of the task invocation
	if resp.StatusCode >= 400 {
		msg, _ := typedvalues.Format(output)
		logrus.Warn("[%s] Failed %v: %v", resp.StatusCode, msg)
		return &types.TaskInvocationStatus{
			Status: types.TaskInvocationStatus_FAILED,
			Error: &types.Error{
				Code:    fmt.Sprintf("%v", resp.StatusCode),
				Message: fmt.Sprintf("%s", msg),
			},
		}, nil
	}

	return &types.TaskInvocationStatus{
		Status: types.TaskInvocationStatus_SUCCEEDED,
		Output: output,
	}, nil
}

func toTypedValue(resp *http.Response) *types.TypedValue {
	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var i interface{} = body
	if strings.Contains(contentType, "application/json") || strings.Contains(contentType, "text/json") {
		logrus.Info("Assuming JSON")
		err := json.Unmarshal(body, &i)
		if err != nil {
			logrus.Warnf("Expected JSON response could not be parsed: %v", err)
		}
	}

	tv, err := typedvalues.Parse(i)
	if err != nil {
		panic(err)
	}
	return tv
}

func getFirstDefinedTypedValue(inputs map[string]*types.TypedValue, fields ...string) (*types.TypedValue, string) {
	var result *types.TypedValue
	var f string
	for _, f = range fields {
		val, ok := inputs[f]
		if ok {
			result = val
			break
		}
	}
	return result, f
}

func toString(tv *types.TypedValue) string {
	if tv == nil {
		return ""
	}
	i, err := typedvalues.Format(tv)
	if err != nil {
		logrus.Warn("Failed to format input: %v", err)
	}

	return fmt.Sprintf("%v", i)
}

func httpMethod(inputs map[string]*types.TypedValue, defaultHttpMethod string) string {
	tv, _ := getFirstDefinedTypedValue(inputs, InputHttpMethod)

	httpMethod := toString(tv)
	if httpMethod == "" {
		return defaultHttpMethod
	}
	return httpMethod
}

func contentType(inputs map[string]*types.TypedValue, mainInput *types.TypedValue, defaultContentType string) string {
	// Check if content type is forced
	tv, _ := getFirstDefinedTypedValue(inputs, InputContentType, InputHeaderPrefix+InputContentType)
	contentType := toString(tv)
	if contentType != "" {
		return contentType
	}
	return inferContentType(mainInput, defaultContentType)
}

func inferContentType(mainInput *types.TypedValue, defaultContentType string) string {
	// Infer content type from main input  (TODO Temporary solution)
	if mainInput != nil && strings.HasPrefix(mainInput.Type, "json") {
		return "application/json"
	}

	// Use default content type
	return defaultContentType
}

func headers(inputs map[string]*types.TypedValue) map[string]string { // TODO support multi-headers at some point
	result := map[string]string{}
	for k, v := range inputs {
		if strings.HasPrefix(k, InputHeaderPrefix) {
			result[strings.TrimPrefix(k, InputHeaderPrefix)] = toString(v)
		}
	}
	return result
}

func query(inputs map[string]*types.TypedValue) url.Values { // TODO support multi-headers at some point
	result := url.Values{}
	for k, v := range inputs {
		if strings.HasPrefix(k, InputQueryPrefix) {
			result.Add(strings.TrimPrefix(k, InputQueryPrefix), toString(v))
		}
	}
	return result
}
