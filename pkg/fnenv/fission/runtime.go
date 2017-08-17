package fission

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/fission/fission"
	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/types"
	controller "github.com/fission/fission/controller/client"
	poolmgr "github.com/fission/fission/poolmgr/client"
	"github.com/sirupsen/logrus"
)

type FunctionEnv struct {
	poolmgr    *poolmgr.Client
	controller *controller.Client
}

func NewFunctionEnv(poolmgr *poolmgr.Client, controller *controller.Client) function.Runtime {
	return &FunctionEnv{
		poolmgr:    poolmgr,
		controller: controller,
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

	input := strings.NewReader(spec.Input[types.INPUT_MAIN])
	// TODO map other parameters as well (to params)

	req, err := http.NewRequest("GET", url, input) // TODO allow change of method
	if err != nil {
		panic(fmt.Errorf("Failed to make request for '%s': %v", serviceUrl, err))
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(fmt.Errorf("Error for url '%s': %v", serviceUrl, err))
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	logrus.Infof("[%s][output]: %v", meta.Name, string(body))

	return &types.FunctionInvocationStatus{
		Status: types.FunctionInvocationStatus_SUCCEEDED,
		Output: body,
	}, nil
}
