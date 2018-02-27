package events

import (
	"errors"
)

var (
	ErrUnknownEvent = errors.New("unknown event")
)

// ResolveTask attempts to convert a string-based flag to the appropriate InvocationEvent.
func ParseInvocation(event string) (Invocation, error) {
	val, ok := Invocation_value[event]
	if !ok {
		return -1, ErrUnknownEvent
	}
	return Invocation(val), nil
}

func ParseWorkflow(flag string) (Workflow, error) {
	val, ok := Workflow_value[flag]
	if !ok {
		return -1, ErrUnknownEvent
	}
	return Workflow(val), nil
}

func ParseTask(event string) (Task, error) {
	val, ok := Task_value[event]
	if !ok {
		return -1, ErrUnknownEvent
	}
	return Task(val), nil
}
