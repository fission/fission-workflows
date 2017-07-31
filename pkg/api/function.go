package api

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/fission/fission"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission/poolmgr/client"
	"github.com/sirupsen/logrus"
)

// Towards environment (Fission)
// Should not be exposed externally, used by controller & co
type FunctionEnvApi interface {
	// Request function invocation (Async)
	Invoke(fn *types.FunctionInvocationSpec) (string, error)

	InvokeSync(fn *types.FunctionInvocationSpec) (*types.FunctionInvocation, error)
	// Cancel function invocation
	//Cancel(id string) error

	// Request status update of function
	//Status()
}

// Towards engine
type FunctionHandlerApi interface {
	// Handler to update the status of the function invocation
	UpdateStatus()
}

type FissionFunctionEnvApi struct {
	poolmgr *client.Client
}

func NewFissionFunctionApi(fission *client.Client) FunctionEnvApi {
	return &FissionFunctionEnvApi{fission}
}

func (fi *FissionFunctionEnvApi) InvokeSync(fn *types.FunctionInvocationSpec) (*types.FunctionInvocation, error) {
	meta := &fission.Metadata{
		Name: fn.GetName(),
		Uid:  fn.GetId(),
	}
	logrus.WithFields(logrus.Fields{
		"metadata": meta,
	}).Debug("Invoking Fission function.")
	serviceUrl, err := fi.poolmgr.GetServiceForFunction(meta)
	if err != nil {
		return nil, err
	}
	//serviceUrl := fmt.Sprintf("http://192.168.99.100:31314/fission-function/%s", fn.GetName())

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

	return &types.FunctionInvocation{}, nil
}

func (fi *FissionFunctionEnvApi) Invoke(fn *types.FunctionInvocationSpec) (string, error) {
	panic("implement me")
}
