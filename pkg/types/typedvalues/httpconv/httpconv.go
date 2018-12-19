// package httpconv provides methods for mapping TypedValues to and from HTTP requests and responses.
//
// Any interaction with HTTP requests or responses should be handled by the high-level implementations in this
// package, such as ParseResponse, FormatResponse, ParseRequest, or FormatRequest.
//
// Although you get most reliable results by properly formatting your requests (especially the Content-Type header)
// this package aims to be lenient by trying to infer content-types, and so on.
package httpconv

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/util/mediatype"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

const (
	inputContentType  = "content-type"
	headerContentType = "Content-Type"
)

var DefaultHTTPMapper = &HTTPMapper{
	DefaultHTTPMethod: http.MethodPost,
	ValueTypeResolver: func(tv *typedvalues.TypedValue) *mediatype.MediaType {
		// Check metadata of the value
		if tv == nil {
			return MediaTypeBytes
		}

		if ct, ok := tv.GetMetadataValue(headerContentType); ok {
			mt, err := mediatype.Parse(ct)
			if err == nil {
				return mt
			}
		}

		// Handle special cases
		switch tv.ValueType() {
		case typedvalues.TypeBytes:
			return MediaTypeBytes
		case typedvalues.TypeNil:
			return MediaTypeBytes
		case typedvalues.TypeString:
			return MediaTypeText
		}

		// For all other structured values, use JSON
		return MediaTypeJSON
	},
	MediaTypeResolver: func(mt *mediatype.MediaType) ParserFormatter {
		if mt == nil {
			return bytesMapper
		}

		// Choose the mapper based on some hard-coded heuristics
		switch mt.Identifier() {
		case MediaTypeJSON.String(), "text/json":
			return jsonMapper
		case MediaTypeText.String():
			return textMapper
		case MediaTypeBytes.String():
			return bytesMapper
		case MediaTypeProtobuf.String(), "application/vnd.google.protobuf", "application/x.google.protobuf",
			"application/x.protobuf":
			return protobufMapper
		}

		// text/* -> TextMapper
		if mt.Type == "text" {
			return textMapper
		}

		// application/* -> BytesMapper
		if mt.Type == "application" {
			return bytesMapper
		}

		// All other: use default
		return bytesMapper
	},
}

func ParseRequest(req *http.Request) (map[string]*typedvalues.TypedValue, error) {
	return DefaultHTTPMapper.ParseRequest(req)
}

func ParseResponse(resp *http.Response) (*typedvalues.TypedValue, error) {
	return DefaultHTTPMapper.ParseResponse(resp)
}

func ParseResponseHeaders(resp *http.Response) *typedvalues.TypedValue {
	return DefaultHTTPMapper.ParseResponseHeaders(resp)
}

func FormatRequest(source map[string]*typedvalues.TypedValue, target *http.Request) error {
	return DefaultHTTPMapper.FormatRequest(source, target)
}

func FormatResponse(w http.ResponseWriter, output *typedvalues.TypedValue, outputHeaders *typedvalues.TypedValue, outputErr *types.Error) {
	DefaultHTTPMapper.FormatResponse(w, output, outputHeaders, outputErr)
}

type HTTPMapper struct {
	DefaultHTTPMethod string
	ValueTypeResolver func(tv *typedvalues.TypedValue) *mediatype.MediaType
	DefaultMediaType  *mediatype.MediaType
	MediaTypeResolver func(mediaType *mediatype.MediaType) ParserFormatter
}

func (h *HTTPMapper) ParseResponse(resp *http.Response) (*typedvalues.TypedValue, error) {
	contentType := h.getRequestContentType(resp.Header)
	defer resp.Body.Close()
	return DefaultHTTPMapper.parseBody(resp.Body, contentType)
}

func (h *HTTPMapper) ParseResponseHeaders(resp *http.Response) *typedvalues.TypedValue {
	return DefaultHTTPMapper.parseRespHeaders(resp)
}

// ParseRequest maps a HTTP request to a target map of typedvalues.
func (h *HTTPMapper) ParseRequest(req *http.Request) (map[string]*typedvalues.TypedValue, error) {
	// Determine content-type
	contentType := h.getRequestContentType(req.Header)
	defer req.Body.Close()

	var body *typedvalues.TypedValue
	var err error
	// TODO support multipart
	switch contentType.Identifier() {
	// Special case: application/x-www-form-urlencoded is the only content-type (?) which also stores data in the url
	case "application/x-www-form-urlencoded":
		req.ParseForm()
		mp := map[string]interface{}{}
		// simplify the map, because we don't support multi-value maps yet
		for k, vs := range req.Form {
			if len(vs) > 0 {
				mp[k] = vs[0]
			}
		}
		body, err = typedvalues.Wrap(mp)
		if err != nil {
			return nil, errors.Errorf("failed to parse form request: %v", err)
		}

	// Default case parse body using the Parser interface
	default:
		body, err = h.parseBody(req.Body, contentType)
		if err != nil {
			return nil, errors.Errorf("failed to parse request: %v", err)
		}
	}

	return map[string]*typedvalues.TypedValue{
		// Map body to "body" input
		types.InputBody: body,

		// Deprecated: Map body to "main/default" input
		types.InputMain: body,

		// Map query to "query.x"
		types.InputQuery: h.parseQuery(req),

		// Map headers to "headers.x"
		types.InputHeaders: h.parseReqHeaders(req),

		// Map http method to "method"
		types.InputMethod: h.parseMethod(req),
	}, nil
}

// FormatResponse maps an TypedValue to an HTTP response
func (h *HTTPMapper) FormatResponse(w http.ResponseWriter, output *typedvalues.TypedValue, outputHeaders *typedvalues.TypedValue, outputErr *types.Error) {
	if w == nil {
		panic("cannot format response to nil")
	}

	if outputErr != nil {
		http.Error(w, outputErr.Error(), http.StatusInternalServerError)
		return
	}

	if output == nil {
		w.WriteHeader(http.StatusNoContent)
		output = typedvalues.MustWrap(nil)
		return
	}

	headers := h.formatHeaders(outputHeaders)
	for k, v := range headers {
		if len(v) > 0 {
			w.Header().Set(k, v[0])
		}
	}

	w.WriteHeader(http.StatusOK)
	contentType := h.ValueTypeResolver(output)
	err := h.formatBody(w, output, contentType)
	if err != nil {
		h.FormatResponse(w, nil, nil, &types.Error{
			Message: fmt.Sprintf("Failed to format response body: %v", err),
		})
	}
	return
}

// FormatRequest maps a map of typed values to an HTTP request
func (h *HTTPMapper) FormatRequest(source map[string]*typedvalues.TypedValue, target *http.Request) error {
	if target == nil {
		panic("cannot format request to nil")
	}

	// Map content-type to the request's content-type
	contentType := h.determineContentTypeFromInputs(source)

	// Map 'body' input to the body of the request
	mainInput := getFirstDefined(source, types.InputBody, types.InputMain)
	if mainInput != nil {
		err := h.formatBody(&requestWriter{req: target}, mainInput, contentType)
		if err != nil {
			return err
		}
	}

	// Map method input to HTTP method
	method := h.formatMethod(source)
	target.Method = method

	// Map query input to URL query
	query := h.formatQuery(source)
	if query != nil {
		if target.URL == nil {
			panic("request has no URL")
		}
		target.URL.RawQuery = query.Encode()
	}

	// Map headers input to HTTP headers
	headers := http.Header{}
	rawHeaders, ok := source[types.InputHeaders]
	if ok && rawHeaders != nil {
		headers = h.formatHeaders(rawHeaders)
	}

	if target.Header == nil {
		target.Header = headers
	} else {
		for k, v := range headers {
			if len(v) > 0 {
				target.Header.Set(k, v[0])
			}
		}
	}
	return nil
}

func (h *HTTPMapper) Clone() *HTTPMapper {
	return &HTTPMapper{
		DefaultMediaType:  h.DefaultMediaType.Copy(),
		DefaultHTTPMethod: h.DefaultHTTPMethod,
		ValueTypeResolver: h.ValueTypeResolver,
		MediaTypeResolver: h.MediaTypeResolver,
	}
}

// parseBody maps the body of the HTTP request to a corresponding typedvalue.
func (h *HTTPMapper) parseBody(data io.Reader, contentType *mediatype.MediaType) (*typedvalues.TypedValue, error) {
	if contentType == nil {
		contentType = h.DefaultMediaType
	}

	return h.MediaTypeResolver(contentType).Parse(contentType, data)
}

// parseMethod maps the method param from a request to a TypedValue
func (h *HTTPMapper) parseMethod(r *http.Request) *typedvalues.TypedValue {
	return typedvalues.MustWrap(r.Method)
}

// parseReqHeaders maps the headers from a request to the "headers" key in the target map
func (h *HTTPMapper) parseReqHeaders(r *http.Request) *typedvalues.TypedValue {
	// For now we do not support multi-valued headers
	headers := flattenMultimap(r.Header)
	return typedvalues.MustWrap(headers)
}

// parseRespHeaders maps the headers from a response to the "headers" key in the target map
func (h *HTTPMapper) parseRespHeaders(r *http.Response) *typedvalues.TypedValue {
	// For now we do not support multi-valued headers
	headers := flattenMultimap(r.Header)
	return typedvalues.MustWrap(headers)
}

// parseQuery maps the query params from a request to the "query" key in the target map
func (h *HTTPMapper) parseQuery(r *http.Request) *typedvalues.TypedValue {
	// For now we do not support multi-valued query params
	query := flattenMultimap(r.URL.Query())
	return typedvalues.MustWrap(query)
}

func (h *HTTPMapper) formatMethod(inputs map[string]*typedvalues.TypedValue) string {
	tv, ok := inputs[types.InputMethod]
	if ok && tv != nil {
		contentType, err := typedvalues.UnwrapString(tv)
		if err == nil {
			return contentType
		}
		logrus.Errorf("Invalid method in inputs: %+v", tv)
	}
	return h.DefaultHTTPMethod
}

// FUTURE: support multivalued query params
func (h *HTTPMapper) formatQuery(inputs map[string]*typedvalues.TypedValue) url.Values {
	queryInput := inputs[types.InputQuery]
	if queryInput == nil {
		return nil
	}
	target := url.Values{}

	i, err := typedvalues.Unwrap(queryInput)
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

func (h *HTTPMapper) formatBody(w http.ResponseWriter, body *typedvalues.TypedValue, contentType *mediatype.MediaType) error {
	if contentType == nil {
		contentType = h.ValueTypeResolver(body)
	}

	return h.MediaTypeResolver(contentType).Format(w, body)
}

func (h *HTTPMapper) findAndParseContentType(inputs map[string]*typedvalues.TypedValue) (*mediatype.MediaType, error) {
	// Check the input[content-type]
	s, err := typedvalues.UnwrapString(inputs[inputContentType])
	if err == nil {
		return mediatype.Parse(s)
	}

	// Check the input[headers][content-type
	headers, err := typedvalues.UnwrapMap(inputs[types.InputHeaders])
	if err != nil {
		return nil, err
	}

	ctHeader, ok := headers[headerContentType].(string)
	if !ok {
		return nil, errors.New("cannot find or parse content-type")
	}

	return mediatype.Parse(ctHeader)
}

func (h *HTTPMapper) determineContentTypeFromInputs(inputs map[string]*typedvalues.TypedValue) *mediatype.MediaType {
	if inputs == nil {
		return h.DefaultMediaType
	}

	mt, err := h.findAndParseContentType(inputs)
	if err != nil {
		mt = h.ValueTypeResolver(getFirstDefined(inputs, types.InputBody, types.InputMain))
	}
	return mt
}

func (h *HTTPMapper) getRequestContentType(headers http.Header) *mediatype.MediaType {
	var contentType *mediatype.MediaType
	ct, err := mediatype.Parse(headers.Get(headerContentType))
	if err != nil {
		contentType = h.DefaultMediaType
	} else {
		contentType = ct
	}
	return contentType
}

// FUTURE: support multi-headers at some point
func (h *HTTPMapper) formatHeaders(inputsHeaders *typedvalues.TypedValue) http.Header {
	headers := http.Header{}

	// TODO handle partial map
	i, err := typedvalues.Unwrap(inputsHeaders)
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

//
// Util
//

func flattenMultimap(mm map[string][]string) map[string]interface{} {
	target := map[string]interface{}{}
	for k, v := range mm {
		target[k] = v[0]
	}
	return target
}

// requestWriter is a wrapper over http.Request to ensure that it conforms with the http.ResponseWriter interface
type requestWriter struct {
	req *http.Request
	buf *bytes.Buffer
}

func (rw *requestWriter) Header() http.Header {
	if rw.req.Header == nil {
		rw.req.Header = http.Header{}
	}
	return rw.req.Header
}

func (rw *requestWriter) Write(data []byte) (int, error) {
	if rw.buf == nil {
		rw.buf = &bytes.Buffer{}
		rw.req.Body = ioutil.NopCloser(rw.buf)
	}
	n, err := rw.buf.Write(data)
	if err != nil {
		return n, err
	}
	rw.Header().Set("Content-Length", fmt.Sprintf("%d", rw.buf.Len()))
	return n, nil
}

func (rw *requestWriter) WriteHeader(statusCode int) {
	return // Not relevant for http.Request
}

func getFirstDefined(inputs map[string]*typedvalues.TypedValue, keys ...string) *typedvalues.TypedValue {
	for _, key := range keys {
		if val, ok := inputs[key]; ok {
			return val
		}
	}
	return nil
}
