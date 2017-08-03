package function

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/fission/fission"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission/poolmgr/client"
	"github.com/sirupsen/logrus"
)

type Api interface {
	// Request function invocation (Async)
	//Invoke(fn *types.FunctionInvocationSpec) (string, error)
	//
	InvokeSync(fn *types.FunctionInvocationSpec) (*types.FunctionInvocation, error)
	// Cancel function invocation
	//Cancel(id string) error

	// Request status update of function
	//Status()
}

// TODO doesn't belong in the API
// Responsible for executing functions
type FissionFunctionApi struct {
	poolmgr *client.Client
}

func NewFissionFunctionApi(fission *client.Client) Api {
	return &FissionFunctionApi{fission}
}

func (fi *FissionFunctionApi) InvokeSync(spec *types.FunctionInvocationSpec) (*types.FunctionInvocation, error) {
	meta := &fission.Metadata{
		Name: spec.GetFunctionName(),
		Uid:  spec.GetFunctionId(),
	}
	logrus.WithFields(logrus.Fields{
		"metadata": meta,
	}).Debug("Invoking Fission function.")
	serviceUrl, err := fi.poolmgr.GetServiceForFunction(meta)
	if err != nil {
		return nil, err
	}

	resp, err := http.Get(serviceUrl) // TODO allow specifying of http method
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Resp: (%s)", body)

	return &types.FunctionInvocation{
		Spec: spec,
		Status: &types.FunctionInvocationStatus{
			Status: types.FunctionInvocationStatus_SUCCEEDED,
		},
	}, nil
}
