package fission

import (
	"bytes"
	"fmt"
	"net/http"

	"encoding/json"
	"github.com/fission/fission"
	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
	controller "github.com/fission/fission/controller/client"
	poolmgr "github.com/fission/fission/poolmgr/client"
	"github.com/sirupsen/logrus"
	"io/ioutil"
)

type FunctionEnv struct {
	poolmgr    *poolmgr.Client
	controller *controller.Client
	ct         *ContentTypeMapper
}

func NewFunctionEnv(poolmgr *poolmgr.Client, controller *controller.Client, pf typedvalues.ParserFormatter) function.Runtime {
	return &FunctionEnv{
		poolmgr:    poolmgr,
		controller: controller,
		ct:         &ContentTypeMapper{pf},
	}
}

func (fe *FunctionEnv) Invoke(spec *types.FunctionInvocationSpec) (*types.FunctionInvocationStatus, error) {
	meta := &fission.Metadata{
		Name: spec.GetType().GetSrc(),
		Uid:  spec.GetType().GetResolved(),
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

	mainInput := spec.Inputs[types.INPUT_MAIN]
	input := bytes.NewReader(mainInput.Value)
	// TODO map other parameters as well (to params)

	req, err := http.NewRequest("GET", url, input) // TODO allow change of method
	if err != nil {
		panic(fmt.Errorf("Failed to make request for '%s': %v", serviceUrl, err))
	}

	reqContentType := fe.ct.ToContentType(mainInput)
	req.Header.Set("Content-Type", reqContentType)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Errorf("Error for url '%s': %v", serviceUrl, err))
	}

	output := fe.ct.ToTypedValue(resp)
	logrus.Infof("[%s][output]: %v", meta.Name, output)

	return &types.FunctionInvocationStatus{
		Status: types.FunctionInvocationStatus_SUCCEEDED,
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
	contentType, ok := formatMapping[val.Type]
	if !ok {
		contentType = "text/plain"
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
