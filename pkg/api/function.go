package api

import (
	"github.com/fission/fission-workflow/pkg/types"
)

type FunctionRuntimeEnv interface {
	Invoke(spec *types.FunctionInvocationSpec) (string, error)

	InvokeSync(spec *types.FunctionInvocationSpec) (*types.FunctionInvocationStatus, error)

	Cancel(fnInvocationId string) error

	Status(fnInvocationId string) (*types.FunctionInvocationStatus, error)
}

type FunctionRegistry interface {
	// Resolve an ambiguous function name to a unique identifier of a function
	//
	// If the fnName does not exist an error will be displayed
	Resolve(fnName string) (string, error) // TODO refactor to FunctionRef (env v2)
}
