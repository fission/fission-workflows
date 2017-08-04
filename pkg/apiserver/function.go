package apiserver

import (
	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/types"
	"golang.org/x/net/context"
)

type GrpcFunctionApiServer struct {
	api function.Api
}

func NewGrpcFunctionApiServer(api function.Api) FunctionEnvApiServer {
	return &GrpcFunctionApiServer{api}
}

func (gf *GrpcFunctionApiServer) Invoke(ctx context.Context, fn *types.FunctionInvocationSpec) (*types.FunctionInvocation, error) {
	return gf.api.InvokeSync(fn)
}
