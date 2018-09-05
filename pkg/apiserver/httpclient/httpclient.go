// Package httpclient is a lightweight implementation of a client for the HTTP gateway.
package httpclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

var (
	ErrResponseError = errors.New("response error")
	ErrRequestCreate = errors.New("failed to create request")
	ErrRequestSend   = errors.New("failed to send request")
	ErrSerialize     = errors.New("failed to serialize input")
	ErrDeserialize   = errors.New("failed to deserialize input")
)

var defaultHTTPClient = http.Client{}
var defaultJSONPBMarshaller = jsonpb.Marshaler{}

func toJSON(dst io.Writer, m proto.Message) error {
	return defaultJSONPBMarshaller.Marshal(dst, m)
}

func fromJSON(src io.Reader, dst proto.Message) error {
	return jsonpb.Unmarshal(src, dst)
}

func callWithJSON(ctx context.Context, method string, url string, in proto.Message, out proto.Message) error {
	buf := bytes.NewBuffer(nil)
	if in != nil {
		err := toJSON(buf, in)
		if err != nil {
			return fmt.Errorf("%v: %v", ErrSerialize, err)
		}
	}
	if logrus.GetLevel() == logrus.DebugLevel {
		logrus.Debugf("--> %s %s", method, url)
		if in != nil {
			buf := bytes.NewBuffer(nil)
			err := toJSON(buf, in)
			if err != nil {
				logrus.Errorf("Failed to jsonify debug data: %v", err)
			}
			data, err := ioutil.ReadAll(buf)
			if err != nil {
				logrus.Errorf("Failed to read debug data: %v", err)
			}
			logrus.Debug("body: '%v'", data)
		}
	}
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrRequestCreate, err)
	}
	req.Header.Set("Content-Type", "application/json")

	// If set, inject opentracing into HTTP request
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		carrier := opentracing.HTTPHeadersCarrier(req.Header)
		opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders, carrier)
	}

	resp, err := defaultHTTPClient.Do(req.WithContext(ctx))
	if err != nil {
		return fmt.Errorf("%v: %v", ErrRequestSend, err)
	}
	if logrus.GetLevel() == logrus.DebugLevel {
		logrus.Debugf("<-- %s - %s", resp.Status, url)
		if resp.Body != nil {
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logrus.Errorf("Failed to read debug body: %v", err)
			}
			logrus.Debugf("Body: '%v'", string(data))
		}
	}
	if resp.StatusCode >= 400 {
		data, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return fmt.Errorf("%v (%s): %s", ErrResponseError, resp.Status, strings.TrimSpace(string(data)))
	}

	if out != nil && resp.ContentLength != 0 {

		err = fromJSON(resp.Body, out)
		if err != nil {
			return fmt.Errorf("%v: %v", ErrDeserialize, err)
		}
		resp.Body.Close()
		if err != nil {
			logrus.Errorf("Failed to close response body: %v", err)
		}
	}
	return nil
}

type baseAPI struct {
	endpoint string
	client   http.Client
}

func (api *baseAPI) formatURL(path string) string {
	return api.endpoint + path
}
