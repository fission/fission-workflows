package events

import (
	"fmt"
)

// Parse attempts to convert a string-based flag to the appropriate InvocationEvent.
func ParseInvocation(event string) (Invocation, error) {
	val, ok := Invocation_value[event]
	if !ok {
		return 0, fmt.Errorf("Unknown InvocationEvent '%s'", event)
	}
	return Invocation(val), nil
}

func ParseWorkflow(flag string) (Workflow, error) {
	val, ok := Workflow_value[flag]
	if !ok {
		return 0, fmt.Errorf("Unknown WorkflowEvent '%s'", flag)
	}
	return Workflow(val), nil
}
