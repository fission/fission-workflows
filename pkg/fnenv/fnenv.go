// Package fnenv provides interfaces to consistently communicate with 'function runtime environments' (fnenvs).
//
// A fnenv is a component responsible for (part of) the execution of the tasks/functions, it commonly consists
// out of at least the following implemented interfaces:
// - resolver: resolves function references in workflow definitions to deterministic function IDs of the fnenv.
// - Runtime: executes a function in the fnenv given the task spec and returns the output.
//
// The fnenv package avoids a single, huge interface, which would make new implementations constrained and expensive,
// by splitting up the functionality into small (optional) interfaces. There is no required combination of interfaces
// that a fnenv needs to implement, although a resolver and Runtime are considered the basic interfaces.
//
// A fnenv could implement additional interfaces which would allow the workflow engine to improve the execution.
// For example, by implementing the Preparer interface, the workflow engine will notify the fnenv ahead of time of the
// incoming function request.
package fnenv

import (
	"context"
	"errors"
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ErrInvalidRuntime = errors.New("invalid runtime")

	FnActive = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "workflows",
		Subsystem: "fnenv",
		Name:      "functions_active",
		Help:      "Number of Fission function executions that are currently active",
	}, []string{"fnenv"})

	FnCount = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "fnenv",
		Subsystem: "fission",
		Name:      "functions_execution_total",
		Help:      "Total number of Fission function executions",
	}, []string{"fnenv"})

	FnExecTime = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "fnenv",
		Subsystem: "fission",
		Name:      "function_execution_time_milliseconds",
		Help:      "Execution time summary of the Fission functions",
	}, []string{"fnenv"})
)

func init() {
	prometheus.MustRegister(FnActive, FnCount, FnExecTime)
}

// Runtime is the minimal interface that a function runtime environment needs to conform with to handle tasks.
type Runtime interface {
	// Invoke executes the task in a blocking way.
	//
	// spec contains the complete configuration needed for the execution.
	// It returns the TaskInvocationStatus with a completed (FINISHED, FAILED, ABORTED) status.
	// An error is returned only when error occurs outside of the runtime's control.
	Invoke(spec *types.TaskInvocationSpec, opts ...InvokeOption) (*types.TaskInvocationStatus, error)
}

// AsyncRuntime is a more extended interface that a runtime can optionally support. It allows for asynchronous
// invocations, allowing with progress tracking and invocation cancellations.
type AsyncRuntime interface {
	// InvokeAsync invokes a function in the runtime based on the spec and returns an identifier to allow the caller
	// to reference the invocation.
	InvokeAsync(spec *types.TaskInvocationSpec, opts ...InvokeOption) (asyncID string, err error)

	// Cancel cancels a function invocation using the function invocation id.
	Cancel(asyncID string) error

	// Status fetches the status of a invocation.
	//
	// The interface user is responsible for determining whether the status indicates that a invocation has completed.
	Status(asyncID string) (*types.TaskInvocationStatus, error)
}

// Preparer allows signalling of a future function invocation.
//
// This allows implementations to prepare for those invocations; performing the necessary
// resource provisioning or setup.
type Preparer interface {
	// Prepare signals that a function invocation is expected at a specific point in time.
	//
	// expectedAt time should be in the future. Any time in the past is interpreted as
	// a signal that the function invocation will come (almost) immediately. fnId is an optional
	// identifier for the signal, which the implementation can use this to identify signals.
	// By default, if fnId is empty, it is not possible to later update the notification.
	Prepare(fn types.FnRef, expectedAt time.Time) error
}

// Resolver resolves a reference to a function to a deterministic, unique function id.
type Resolver interface {
	// ResolveTask resolved an ambiguous target function name to a unique identifier of a function
	//
	// If the targetFn does not exist an error will be displayed
	Resolve(targetFn string) (types.FnRef, error)
}

// RuntimeResolver is the runtime environment component that resolves a reference to a function to a deterministic,
// runtime-specific function UID.
type RuntimeResolver interface {
	// ResolveTask resolved an ambiguous target function name to a unique identifier of a function within the scope
	// of a runtime.
	Resolve(ref types.FnRef) (string, error)
}

type InvokeConfig struct {
	Ctx           context.Context
	AwaitWorkflow time.Duration
}

type InvokeOption func(config *InvokeConfig)

func ParseInvokeOptions(opts []InvokeOption) *InvokeConfig {
	// Default
	cfg := &InvokeConfig{
		Ctx: context.Background(),
	}
	// Parse options
	for _, opt := range opts {
		opt(cfg)
	}
	return cfg
}

func AwaitWorkflow(timeout time.Duration) InvokeOption {
	return func(config *InvokeConfig) {
		config.AwaitWorkflow = timeout
	}
}

func WithContext(ctx context.Context) InvokeOption {
	return func(config *InvokeConfig) {
		config.Ctx = ctx
	}
}
