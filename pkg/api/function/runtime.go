package function

import (
	"github.com/fission/fission-workflows/pkg/types"
)

// Runtime is the minimal interface that a function runtime environment needs to conform with to handle tasks.
type Runtime interface {

	// Invoke executes the task in a blocking way.
	//
	// spec contains the complete configuration needed for the execution.
	// It returns the TaskInvocationStatus with a completed (FINISHED, FAILED, ABORTED) status.
	// An error is returned only when error occurs outside of the runtime's control.
	Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocationStatus, error)
}

// AsyncRuntime is a more extended interface that a runtime can optionally support. It allows for asynchronous
// invocations, allowing with progress tracking and invocation cancellations.
type AsyncRuntime interface {

	// InvokeAsync invokes a function in the runtime based on the spec and returns an identifier to allow the caller
	// to reference the invocation.
	InvokeAsync(spec *types.TaskInvocationSpec) (string, error)

	// Cancel cancels a function invocation using the function invocation ID.
	Cancel(fnInvocationId string) error

	// Status fetches the status of a invocation.
	//
	// The interface user is responsible for determining wether the status indicates that a invocation has completed.
	Status(fnInvocationId string) (*types.TaskInvocationStatus, error)
}

// Resolver is the component that resolves a functionRef to a runtime-specific function UID.
type Resolver interface {
	// Resolve an ambiguous function name to a unique identifier of a function
	//
	// If the fnName does not exist an error will be displayed
	Resolve(fnName string) (string, error) // TODO refactor to FunctionRef (Fission Env v2)
}
