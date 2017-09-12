package events

import (
	"errors"
)

var (
	ErrUnkownEvent = errors.New("unknown event")
)

// Resolve attempts to convert a string-based flag to the appropriate InvocationEvent.
func ParseInvocation(event string) (Invocation, error) {
	val, ok := Invocation_value[event]
	if !ok {
		return -1, ErrUnkownEvent
	}
	return Invocation(val), nil
}

func ParseWorkflow(flag string) (Workflow, error) {
	val, ok := Workflow_value[flag]
	if !ok {
		return -1, ErrUnkownEvent
	}
	return Workflow(val), nil
}

func ParseFunction(event string) (Function, error) {
	val, ok := Function_value[event]
	if !ok {
		return -1, ErrUnkownEvent
	}
	return Function(val), nil
}
