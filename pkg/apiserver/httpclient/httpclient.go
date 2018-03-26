// Package httpclient is a lightweight implementation of the HTTP gateway.
package httpclient

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

var (
	ErrResponseError = errors.New("response error")
	ErrRequestCreate = errors.New("failed to create request")
	ErrRequestSend   = errors.New("failed to send request")
	ErrSerialize     = errors.New("failed to serialize input")
	ErrDeserialize   = errors.New("failed to deserialize input")
)

var defaultHttpClient = http.Client{}
var defaultJsonPbMarshaller = jsonpb.Marshaler{}

func toJson(dst io.Writer, m proto.Message) error {
	return defaultJsonPbMarshaller.Marshal(dst, m)
}

func fromJson(src io.Reader, dst proto.Message) error {
	return jsonpb.Unmarshal(src, dst)
}

func call(method string, url string, in proto.Message, out proto.Message) error {
	buf := bytes.NewBuffer(nil)
	if in != nil {
		err := toJson(buf, in)
		if err != nil {
			return fmt.Errorf("%v: %v", ErrSerialize, err)
		}
	}
	if logrus.GetLevel() == logrus.DebugLevel {
		logrus.Debugf("--> %s - %s", method, url)
		if in != nil {
			buf := bytes.NewBuffer(nil)
			err := toJson(buf, in)
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
	resp, err := defaultHttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("%v: %v", ErrRequestSend, err)
	}
	if logrus.GetLevel() == logrus.DebugLevel {
		logrus.Debugf("<-- %s - %s", resp.Status, url)
		if resp.Body != nil {
			data, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				logrus.Errorf("Failed to read debug body", err)
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

		err = fromJson(resp.Body, out)
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

type BaseApi struct {
	endpoint string
	client   http.Client
}

// TODO remove hard-coded proxy
func (api *BaseApi) formatUrl(path string) string {
	return api.endpoint + "/proxy/workflows-apiserver" + path
}
