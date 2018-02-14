// Package fnenv provides interfaces to consistently communicate with 'function runtime environments' (fnenvs).
//
// A fnenv is a component responsible for (part of) the execution of the tasks/functions, it commonly consists
// out of at least the following implemented interfaces:
// - Resolver: resolves function references in workflow definitions to deterministic function IDs of the fnenv.
// - Runtime: executes a function in the fnenv given the task spec and returns the output.
//
// The fnenv package avoids a single, huge interface, which would make new implementations constrained and expensive,
// by splitting up the functionality into small (optional) interfaces. There is no required combination of interfaces
// that a fnenv needs to implement, although a Resolver and Runtime are considered the basic interfaces.
//
// A fnenv could implement additional interfaces which would allow the workflow engine to improve the execution.
// For example, by implementing the Notifier interface, the workflow engine will notify the fnenv ahead of time of the
// incoming function request.
package fnenv

import (
	"time"

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
	InvokeAsync(spec *types.TaskInvocationSpec) (fnInvocationID string, err error)

	// Cancel cancels a function invocation using the function invocation ID.
	Cancel(fnInvocationID string) error

	// Status fetches the status of a invocation.
	//
	// The interface user is responsible for determining whether the status indicates that a invocation has completed.
	Status(fnInvocationID string) (*types.TaskInvocationStatus, error)
}

// Resolver is the component that resolves a reference to a function to a deterministic, runtime-specific function UID.
type Resolver interface {

	// Resolve an ambiguous function name to a unique identifier of a function
	//
	// If the fnName does not exist an error will be displayed
	// TODO refactor to FunctionRef (Fission Env v2)
	Resolve(fnRef string) (string, error)
}

// Notifier allows signalling of an incoming function invocation.
//
// This allows implementations to prepare for those invocations; performing the necessary
// resource provisioning or setup.
type Notifier interface {

	// Notify signals that a function invocation is expected at a specific point in time.
	//
	// expectedAt time should be in the future. Any time in the past is interpreted as
	// a signal that the function invocation will come (almost) immediately. fnId is an optional
	// identifier for the signal, which the implementation can use this to identify signals.
	// By default, if fnId is empty, it is not possible to later update the notification.
	Notify(taskID string, fn types.ResolvedTask, expectedAt time.Time) error
}
