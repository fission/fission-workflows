package types

import "fmt"

func ParseInvocationEvent(flag string) (InvocationEvent, error) {
	val, ok := InvocationEvent_value[flag]
	if !ok {
		return 0, fmt.Errorf("Unknown InvocationEvent '%s'", flag)
	}
	return InvocationEvent(val), nil
}

func ParseWorkflowEvent(flag string) (WorkflowEvent, error) {
	val, ok := WorkflowEvent_value[flag]
	if !ok {
		return 0, fmt.Errorf("Unknown WorkflowEvent '%s'", flag)
	}
	return WorkflowEvent(val), nil
}
