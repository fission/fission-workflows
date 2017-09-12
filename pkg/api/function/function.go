package function

import (
	"github.com/fission/fission-workflow/pkg/types"
)

type Runtime interface {
	Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocationStatus, error)
}

type AsyncRuntime interface {
	InvokeAsync(spec *types.TaskInvocationSpec) (string, error)

	Cancel(fnInvocationId string) error

	Status(fnInvocationId string) (*types.TaskInvocationStatus, error)
}

type Resolver interface {
	// Resolve an ambiguous function name to a unique identifier of a function
	//
	// If the fnName does not exist an error will be displayed
	Resolve(fnName string) (string, error) // TODO refactor to FunctionRef (env v2)
}
