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
	INPUT_HEADERS      = "headers"
	INPUT_QUERY        = "query"
	INPUT_METHOD       = "method"

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

func (wi WorkflowInvocationStatus) Finished() bool {
	for _, event := range invocationFinalStates {
		if event == wi.Status {
			return true
		}
	}
	return false
}

func (wi WorkflowInvocationStatus) Successful() bool {
	return wi.Status == WorkflowInvocationStatus_SUCCEEDED
}

func (ti TaskInvocationStatus) Finished() bool {
	for _, event := range taskFinalStates {
		if event == ti.Status {
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
