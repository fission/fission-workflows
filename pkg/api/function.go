package api

// Towards environment (Fission)
// Should not be exposed externally, used by controller & co
type FunctionApi interface {
	// Request function invocation
	Invoke()

	// Cancel function invocation
	Cancel()

	// Request status update of function
	Status()
}

// Towards engine
type FunctionHandlerApi interface {
	// Handler to update the status of the function invocation
	UpdateStatus()
}

type FunctionInvocation interface {
	Id() FunctionInvocationMetadata
	Spec() FunctionInvocationSpec
	Status() FunctionInvocationStatus
}

type FunctionInvocationSpec interface {
	// Strategy
}

type FunctionInvocationStatus interface {
}

type FunctionInvocationMetadata interface {
}
