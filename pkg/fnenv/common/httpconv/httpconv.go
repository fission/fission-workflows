package httpconv

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

const (
	inputContentType    = "content-type"
	headerContentType   = "Content-Type"
	contentTypeJSON     = "application/json"
	contentTypeBytes    = "application/octet-stream"
	contentTypeText     = "text/plain"
	contentTypeTask     = "application/vnd.fission.workflows.workflow" // Default format: protobuf, +json for json
	contentTypeWorkflow = "application/vnd.fission.workflows.task"     // Default format: protobuf, +json for json
	contentTypeProtobuf = "application/protobuf"                       // Default format: protobuf, +json for json
	contentTypeDefault  = contentTypeText
	methodDefault       = http.MethodPost
)

// ParseRequest maps a HTTP request to a target map.
func ParseRequest(r *http.Request) (map[string]*types.TypedValue, error) {
	target := map[string]*types.TypedValue{}
	// Content-Type is a common problem, so log this for every request
	contentType := r.Header.Get(headerContentType)

	// Map body to "main" input
	bodyInput, err := ParseBody(r.Body, contentType)
	defer r.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("failed to parse request: %v", err)
	}

	target[types.InputBody] = &bodyInput

	// Deprecated: body is mapped to 'default'
	target[types.InputMain] = &bodyInput

	// Map query to "query.x"
	query := ParseQuery(r)
	target[types.InputQuery] = &query

	// Map headers to "headers.x"
	headers := ParseHeaders(r)
	target[types.InputHeaders] = &headers

	// Map http method to "method"
	method := ParseMethod(r)
	target[types.InputMethod] = &method

	return target, nil
}

func ParseBody(data io.Reader, contentType string) (types.TypedValue, error) {
	if len(contentType) == 0 {
		contentType = contentTypeDefault
	}

	tv := types.TypedValue{}
	tv.SetLabel(headerContentType, contentType)

	bs, err := ioutil.ReadAll(data)
	if err != nil {
		return tv, err
	}

	// Attempt to parse body according to provided ContentType
	switch normalizeContentType(contentType) {
	case contentTypeJSON:
		var i interface{}
		err := json.Unmarshal(bs, &i)
		if err != nil {
			logrus.Warnf("Failed to parse JSON data (len: %v, data: '%.50s' cause: %v), skipping further parsing.",
				len(bs), string(bs), err)
			tv = *typedvalues.ParseBytes(bs)
		} else {
			tv = *typedvalues.MustParse(i)
		}
	case contentTypeText:
		tv = *typedvalues.ParseString(string(bs))
	case contentTypeProtobuf:
		fallthrough
	case contentTypeTask:
		fallthrough
	case contentTypeWorkflow:
		// TODO support json-encoded workflow/task
		var m proto.Message
		err := proto.Unmarshal(bs, m)
		if err != nil {
			return tv, err
		}
		t, err := typedvalues.Parse(m)
		if err != nil {
			return tv, err
		}
		tv = *t
	default:
		// In other cases do not attempt to interpret the data
		fallthrough
	case contentTypeBytes:
		tv = *typedvalues.ParseBytes(bs)
	}

	return tv, nil
}

// ParseMethod maps the method param from a request to a TypedValue
func ParseMethod(r *http.Request) types.TypedValue {
	return *typedvalues.ParseString(r.Method)
}

// ParseHeaders maps the headers from a request to the "headers" key in the target map
func ParseHeaders(r *http.Request) types.TypedValue {
	// For now we do not support multi-valued headers
	headers := flattenMultimap(r.Header)

	tv := typedvalues.MustParse(headers)
	return *tv
}

// ParseQuery maps the query params from a request to the "query" key in the target map
func ParseQuery(r *http.Request) types.TypedValue {
	// For now we do not support multi-valued query params
	query := flattenMultimap(r.URL.Query())

	tv := typedvalues.MustParse(query)
	return *tv
}

// formatting logic
func FormatResponse(w http.ResponseWriter, output *types.TypedValue, outputErr *types.Error) {
	if w == nil {
		panic("cannot format response to nil")
	}

	if outputErr != nil {
		// TODO provide different http codes based on error
		http.Error(w, outputErr.Error(), http.StatusInternalServerError)
		return
	}

	if output == nil {
		w.WriteHeader(http.StatusNoContent)
		output = typedvalues.ParseNil()
		return
	}

	w.WriteHeader(http.StatusOK)
	contentType := DetermineContentType(output)
	w.Header().Set(headerContentType, contentType)
	bs, err := FormatBody(*output, contentType)
	if err != nil {
		FormatResponse(w, nil, &types.Error{
			Message: fmt.Sprintf("Failed to format response body: %v", err),
		})
	}
	w.Write(bs)
	return
}

func FormatRequest(source map[string]*types.TypedValue, target *http.Request) error {
	if target == nil {
		panic("cannot format request to nil")
	}

	// Map content-type to the request's content-type
	contentType := DetermineContentTypeInputs(source)

	// Map 'body' input to the body of the request
	mainInput, ok := source[types.InputBody]
	if !ok {
		mainInput, ok = source[types.InputMain]
	}
	if ok && mainInput != nil {
		bs, err := FormatBody(*mainInput, contentType)
		if err != nil {
			return err
		}
		target.Body = ioutil.NopCloser(bytes.NewReader(bs))
		target.ContentLength = int64(len(bs))
	}

	// Map method input to HTTP method
	method := FormatMethod(source)
	target.Method = method

	// Map query input to URL query
	query := FormatQuery(source)
	if query != nil {
		if target.URL == nil {
			panic("request has no URL")
		}
		target.URL.RawQuery = query.Encode()
	}

	// Map headers input to HTTP headers
	headers := FormatHeaders(source)
	if target.Header == nil {
		target.Header = headers
	} else {
		for k, v := range headers {
			if len(v) > 0 {
				target.Header.Set(k, v[0])
			}
		}
	}
	target.Header.Set(headerContentType, contentType)

	return nil
}

func FormatMethod(inputs map[string]*types.TypedValue) string {
	tv, ok := inputs[types.InputMethod]
	if ok && tv != nil {
		contentType, err := typedvalues.FormatString(tv)
		if err == nil {
			return contentType
		}
		logrus.Error("Invalid method in inputs: %+v", tv)
	}
	return methodDefault
}

// TODO support multivalued query params at some point
func FormatQuery(inputs map[string]*types.TypedValue) url.Values {
	queryInput := inputs[types.InputQuery]
	if queryInput == nil {
		return nil
	}
	target := url.Values{}

	i, err := typedvalues.Format(queryInput)
	if err != nil {
		logrus.Errorf("Failed to format headers: %v", err)
	}

	switch i.(type) {
	case map[string]interface{}:
		for k, v := range i.(map[string]interface{}) {
			target.Add(k, fmt.Sprintf("%v", v))
		}
	default:
		logrus.Warnf("Ignoring invalid type of query input (expected map[string]interface{}, was %v)",
			reflect.TypeOf(i))
	}
	return target
}

func FormatBody(value types.TypedValue, contentType string) ([]byte, error) {
	if len(contentType) == 0 {
		contentType = contentTypeDefault
	}

	i, err := typedvalues.Format(&value)
	if err != nil {
		return nil, err
	}

	// Attempt to parse body according to provided ContentType
	var bs []byte
	switch normalizeContentType(contentType) {
	case contentTypeJSON:
		bs, err = json.Marshal(i)
		if err != nil {
			return nil, err
		}
	case contentTypeText:
		switch t := i.(type) {
		case string:
			bs = []byte(t)
		case []byte:
			bs = t
		default:
			bs = []byte(fmt.Sprintf("%v", t))
		}
	case contentTypeProtobuf:
		fallthrough
	case contentTypeTask:
		fallthrough
	case contentTypeWorkflow:
		// TODO support json
		m, ok := i.(proto.Message)
		if !ok {
			return nil, fmt.Errorf("illegal content type '%T', should be protobuf", i)
		}
		bs, err = proto.Marshal(m)
		if err != nil {
			return nil, err
		}
	default:
		fallthrough
	case contentTypeBytes:
		var ok bool
		bs, ok = i.([]byte)
		if !ok {
			return nil, fmt.Errorf("illegal content type '%T', should be []byte", i)
		}
	}
	return bs, nil
}

func DetermineContentType(value *types.TypedValue) string {
	if value == nil {
		return contentTypeBytes
	}

	ct, ok := value.GetLabel(headerContentType)
	if ok && len(ct) > 0 {
		return ct
	}

	// Otherwise, check for primitive types of the main input
	switch typedvalues.ValueType(value.Type) {
	// TODO task and workflow
	case typedvalues.TypeMap:
		fallthrough
	case typedvalues.TypeList:
		return contentTypeJSON
	case typedvalues.TypeNumber:
		fallthrough
	case typedvalues.TypeExpression:
		fallthrough
	case typedvalues.TypeString:
		return contentTypeText
	default:
		return contentTypeBytes
	}
}

func DetermineContentTypeInputs(inputs map[string]*types.TypedValue) string {
	// Check for forced contentType in inputs
	ctTv, ok := inputs[inputContentType]
	if ok && ctTv != nil {
		contentType, err := typedvalues.FormatString(ctTv)
		if err == nil {
			return contentType
		}
		logrus.Error("Invalid content type in inputs: %+v", ctTv)
	}

	// Otherwise, check for label on body input
	if inputs[types.InputBody] != nil {
		return DetermineContentType(inputs[types.InputBody])
	} else {
		return DetermineContentType(inputs[types.InputMain])
	}
}

// TODO support multi-headers at some point
func FormatHeaders(inputs map[string]*types.TypedValue) http.Header {
	headers := http.Header{}
	rawHeaders, ok := inputs[types.InputHeaders]
	if !ok || rawHeaders == nil {
		return headers
	}

	// TODO handle partial map
	i, err := typedvalues.Format(rawHeaders)
	if err != nil {
		logrus.Errorf("Failed to format headers: %v", err)
	}

	switch i.(type) {
	case map[string]interface{}:
		for k, v := range i.(map[string]interface{}) {
			headers.Add(k, fmt.Sprintf("%v", v))
		}
	default:
		logrus.Warnf("Ignoring invalid type of headers input (expected map[string]interface{}, was %v)",
			reflect.TypeOf(i))
	}
	return headers
}

// Util

func flattenMultimap(mm map[string][]string) map[string]interface{} {
	target := map[string]interface{}{}
	for k, v := range mm {
		target[k] = v[0]
	}
	return target
}

func normalizeContentType(contentType string) string {
	contentType = strings.ToLower(strings.TrimSpace(contentType))
	matchContentType := contentType
	// Heuristics, because everything to do with HTTP is ambiguous...
	if strings.Contains(contentType, "json") { // TODO exclude JSON representation of protobuf objects
		matchContentType = contentTypeJSON
	}
	if strings.HasPrefix(contentType, "text") {
		matchContentType = contentTypeText
	}
	if strings.Contains(contentType, "protobuf") {
		matchContentType = contentTypeProtobuf
	}
	return matchContentType
}
