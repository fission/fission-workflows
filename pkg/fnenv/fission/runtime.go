package fission

import (
	"bytes"
	"fmt"
	"net/http"

	"encoding/json"
	"io/ioutil"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
	poolmgr "github.com/fission/fission/poolmgr/client"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/1.5/pkg/api"

	"strings"

	k8stypes "k8s.io/client-go/1.5/pkg/types"
)

// FunctionEnv adapts the Fission platform to the function execution runtime.
type FunctionEnv struct {
	poolmgr *poolmgr.Client
	ct      *ContentTypeMapper
}

func NewFunctionEnv(poolmgr *poolmgr.Client) *FunctionEnv {
	return &FunctionEnv{
		poolmgr: poolmgr,
		ct:      &ContentTypeMapper{typedvalues.DefaultParserFormatter},
	}
}

func (fe *FunctionEnv) Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocationStatus, error) {
	meta := &api.ObjectMeta{
		Name:      spec.GetType().GetSrc(),
		UID:       k8stypes.UID(spec.GetType().GetResolved()),
		Namespace: api.NamespaceDefault,
	}
	logrus.WithFields(logrus.Fields{
		"metadata": meta,
	}).Debug("Invoking Fission function.")
	serviceUrl, err := fe.poolmgr.GetServiceForFunction(meta)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"err":  err,
			"meta": meta,
		}).Error("Fission function failed!")
		return nil, err
	}

	url := fmt.Sprintf("http://%s", serviceUrl)

	// Map input parameters to actual Fission function parameters

	var input []byte
	mainInput, ok := spec.Inputs[types.INPUT_MAIN]
	if ok {
		input = mainInput.Value
	}
	// TODO map other parameters as well (to params)

	req, err := http.NewRequest("GET", url, bytes.NewReader(input)) // TODO allow change of method
	if err != nil {
		panic(fmt.Errorf("failed to make request for '%s': %v", serviceUrl, err))
	}

	reqContentType := fe.ct.ToContentType(mainInput)
	req.Header.Set("Content-Type", reqContentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Errorf("error for url '%s': %v", serviceUrl, err))
	}

	output := fe.ct.ToTypedValue(resp)
	logrus.Infof("[%s][output]: %v", meta.Name, output)

	return &types.TaskInvocationStatus{
		Status: types.TaskInvocationStatus_SUCCEEDED,
		Output: output,
	}, nil
}

type ContentTypeMapper struct {
	parserFormatter typedvalues.ParserFormatter
}

var formatMapping = map[string]string{
	typedvalues.FORMAT_JSON: "application/json",
}

func (ct *ContentTypeMapper) ToContentType(val *types.TypedValue) string {
	contentType := "text/plain"
	if val == nil {
		return contentType
	}

	// Temporary solution
	if strings.HasPrefix(val.Type, "json") {
		contentType = "application/json"
	}
	return contentType
}

func (ct *ContentTypeMapper) ToTypedValue(resp *http.Response) *types.TypedValue {
	contentType := resp.Header.Get("Content-Type")
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var i interface{} = body
	switch contentType {
	case "application/json":
		fallthrough
	case "text/json":
		json.Unmarshal(body, &i)
	}

	tv, err := ct.parserFormatter.Parse(i)
	if err != nil {
		panic(err)
	}
	return tv
}
