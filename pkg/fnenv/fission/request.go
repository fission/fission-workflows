package fission

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/fission/fission-workflows/pkg/fnenv/common/httpconv"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
)

const (
	inputBody        = "body" // or 'default'
	inputHTTPMethod  = "method"
	inputContentType = "content_type" // to force the content type

	defaultContentType = "text/plain"
	headerContentType  = "Content-Type"
)

// Format maps values of the source map to the (Fission) request.
func formatRequest(r *http.Request, source map[string]*types.TypedValue) error {
	// Map headers inputs to request
	formatHeaders(r, source) // TODO move error handling here

	// Map HTTP method inputs to request, or use default
	formatMethod(r, source)

	// Map query inputs to request
	formatQuery(r.URL, source) // TODO move error handling here

	// Set the Content-Type
	r.Header.Set(headerContentType, formatContentType(source, defaultContentType))

	// Map body inputs to request
	r.Body = formatBody(source)

	return nil
}

// Parse maps a (Fission) request to a target map.
func parseRequest(r *http.Request, target map[string]*types.TypedValue) error {
	// Content-Type is a common problem, so log this for every request
	contentType := r.Header.Get(headerContentType)
	logrus.WithField("url", r.URL).WithField(headerContentType, contentType).Info("Request Content-Type")

	// Map body to "main" input
	bodyInput, err := httpconv.ParseBody(r.Body, contentType)
	defer r.Body.Close()
	if err != nil {
		return fmt.Errorf("failed to parse request: %v", err)
	}
	target[types.INPUT_MAIN] = &bodyInput

	// Map query to "query.x"
	err = parseQuery(r, target)
	if err != nil {
		return fmt.Errorf("failed to parse request: %v", err)
	}

	// Map headers to "headers.x"
	err = parseHeaders(r, target)
	if err != nil {
		return fmt.Errorf("failed to parse request: %v", err)
	}

	// Map http method to "method"
	err = parseMethod(r, target)
	if err != nil {
		return fmt.Errorf("failed to parse request: %v", err)
	}

	return nil
}

// parseHeaders maps the headers from a request to the "headers" key in the target map
func parseHeaders(r *http.Request, target map[string]*types.TypedValue) error {
	// For now we do not support multi-valued headers
	headers := flattenMultimap(r.Header)

	tv, err := typedvalues.Parse(headers)
	if err != nil {
		logrus.Errorf("Failed to parse headers: %v", err)
		return fmt.Errorf("failed to parse headers: %v", err)
	}
	target[types.INPUT_HEADERS] = tv
	return nil
}

// parseQuery maps the query params from a request to the "query" key in the target map
func parseQuery(r *http.Request, target map[string]*types.TypedValue) error {
	// For now we do not support multi-valued query params
	query := flattenMultimap(r.URL.Query())

	tv, err := typedvalues.Parse(query)
	if err != nil {
		logrus.Errorf("Failed to parse query: %v", err)
		return fmt.Errorf("failed to parse query: %v", err)
	}
	target[types.INPUT_QUERY] = tv
	return nil
}

// parseMethod maps the method param from a request to the "method" key in the target map
func parseMethod(r *http.Request, target map[string]*types.TypedValue) error {
	method, err := typedvalues.Parse(r.Method)
	if err != nil {
		logrus.Errorf("Failed to parse the http method: %v", err)
		return errors.New("failed to parse http method")
	}
	target[types.INPUT_METHOD] = method
	return nil
}

func flattenMultimap(mm map[string][]string) map[string]interface{} {
	target := map[string]interface{}{}
	for k, v := range mm {
		target[k] = v[0]
	}
	return target
}

// formatting logic

func formatMethod(target *http.Request, inputs map[string]*types.TypedValue) {
	_, tv := getFirstDefinedTypedValue(inputs, inputHTTPMethod)
	httpMethod := toString(tv)
	if httpMethod != "" {
		target.Method = httpMethod
	}
}

// TODO support multivalued query params at some point
func formatQuery(targetURL *url.URL, inputs map[string]*types.TypedValue) {
	queryInput := inputs[types.INPUT_QUERY]
	if queryInput == nil {
		return
	}

	i, err := typedvalues.Format(queryInput)
	if err != nil {
		logrus.Errorf("Failed to format headers: %v", err)
	}

	switch i.(type) {
	case map[string]interface{}:
		origQuery := targetURL.Query()
		for k, v := range i.(map[string]interface{}) {
			origQuery.Add(k, fmt.Sprintf("%v", v))
		}
		targetURL.RawQuery = origQuery.Encode()
	default:
		logrus.Warnf("Ignoring invalid type of query input (expected map[string]interface{}, was %v)",
			reflect.TypeOf(i))
	}
}

// TODO support multi-headers at some point
func formatHeaders(target *http.Request, inputs map[string]*types.TypedValue) {
	rawHeaders := inputs[types.INPUT_HEADERS]
	if rawHeaders == nil {
		return
	}

	i, err := typedvalues.Format(rawHeaders)
	if err != nil {
		logrus.Errorf("Failed to format headers: %v", err)
	}

	switch i.(type) {
	case map[string]interface{}:
		if target.Header == nil {
			target.Header = http.Header{}
		}
		for k, v := range i.(map[string]interface{}) {
			target.Header.Add(k, fmt.Sprintf("%v", v))
		}
	default:
		logrus.Warnf("Ignoring invalid type of headers input (expected map[string]interface{}, was %v)",
			reflect.TypeOf(i))
	}
}

func formatBody(inputs map[string]*types.TypedValue) io.ReadCloser {
	var input []byte
	_, mainInput := getFirstDefinedTypedValue(inputs, types.INPUT_MAIN, inputBody)
	if mainInput != nil {
		// TODO ensure that it is a byte-representation 1-1 of actual value not the representation in TypedValue
		input = mainInput.Value
	}

	return ioutil.NopCloser(bytes.NewReader(input))
}

func formatContentType(inputs map[string]*types.TypedValue, defaultContentType string) string {
	// Check if content type is forced
	_, tv := getFirstDefinedTypedValue(inputs, inputContentType) // TODO lookup in headers?
	contentType := toString(tv)
	if contentType != "" {
		return contentType
	}
	return inferContentType(inputs[types.INPUT_MAIN], defaultContentType)
}

func inferContentType(mainInput *types.TypedValue, defaultContentType string) string {
	// Infer content type from main input  (TODO Temporary solution)
	if mainInput != nil && strings.HasPrefix(mainInput.Type, "json") {
		return "application/json"
	}

	// Use default content type
	return defaultContentType
}

// Util functions

// getFirstDefinedTypedValue returns the first input and key of the inputs argument that matches a field in fields.
// For example, given inputs { a : b, c : d }, getFirstDefinedTypedValue(inputs, z, x, c, a) would return (c, d)
func getFirstDefinedTypedValue(inputs map[string]*types.TypedValue, fields ...string) (string, *types.TypedValue) {
	var result *types.TypedValue
	var key string
	for _, key = range fields {
		val, ok := inputs[key]
		if ok {
			result = val
			break
		}
	}
	return key, result
}

// toString is a utility function to do an unsafe conversion of a TypedValue to a String. fmt.Sprintf is used to
// convert other types to their string representation.
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
