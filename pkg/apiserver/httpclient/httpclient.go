// Package httpclient is a lightweight implementation of the HTTP gateway.
package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
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
			return err
		}
	}
	req, err := http.NewRequest(method, url, buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := defaultHttpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode >= 400 {
		data, _ := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		return fmt.Errorf("request failed - %s: %s", resp.Status, strings.TrimSpace(string(data)))
	}

	if out != nil {
		err = fromJson(resp.Body, out)
		if err != nil {
			return err
		}
		err = resp.Body.Close()
		if err != nil {
			return err
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
