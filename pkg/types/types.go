package types

import (
	"fmt"
	"strings"
)

// Types other than specified in protobuf
const (
	SUBJECT_INVOCATION = "invocation"
	SUBJECT_WORKFLOW   = "workflows"
	INPUT_MAIN         = "default"

	typedValueShortMaxLen = 32
)

// InvocationEvent
var invocationFinalStates = []WorkflowInvocationStatus_Status{
	WorkflowInvocationStatus_ABORTED,
	WorkflowInvocationStatus_SUCCEEDED,
	WorkflowInvocationStatus_FAILED,
}

var taskFinalStates = []TaskInvocationStatus_Status{
	TaskInvocationStatus_FAILED,
	TaskInvocationStatus_ABORTED,
	TaskInvocationStatus_SKIPPED,
	TaskInvocationStatus_SUCCEEDED,
}

func (wi WorkflowInvocationStatus_Status) Finished() bool {
	for _, event := range invocationFinalStates {
		if event == wi {
			return true
		}
	}
	return false
}

func (wi WorkflowInvocationStatus_Status) Successful() bool {
	return wi == WorkflowInvocationStatus_SUCCEEDED
}

func (ti TaskInvocationStatus_Status) Finished() bool {
	for _, event := range taskFinalStates {
		if event == ti {
			return true
		}
	}
	return false
}

// Prints a short description of the value
func (tv TypedValue) Short() string {
	var val string
	if len(tv.Value) > typedValueShortMaxLen {
		val = fmt.Sprintf("%s[..%d..]", tv.Value[:typedValueShortMaxLen], len(tv.Value)-typedValueShortMaxLen)
	} else {
		val = fmt.Sprintf("%s", tv.Value)
	}

	return fmt.Sprintf("<Type=\"%s\", Val=\"%v\">", tv.Type, strings.Replace(val, "\n", "", -1))
}
