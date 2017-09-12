package types

// Types other than specified in protobuf
const (
	SUBJECT_INVOCATION = "invocation"
	SUBJECT_WORKFLOW   = "workflows"
	INPUT_MAIN         = "default"
)

// InvocationEvent
var invocationFinalStates = []WorkflowInvocationStatus_Status{
	WorkflowInvocationStatus_ABORTED,
	WorkflowInvocationStatus_SUCCEEDED,
	WorkflowInvocationStatus_FAILED,
}

var functionFinalStates = []FunctionInvocationStatus_Status{
	FunctionInvocationStatus_FAILED,
	FunctionInvocationStatus_ABORTED,
	FunctionInvocationStatus_SKIPPED,
	FunctionInvocationStatus_SUCCEEDED,
}

func (wi WorkflowInvocationStatus_Status) Finished() bool {
	for _, event := range invocationFinalStates {
		if event == wi {
			return true
		}
	}
	return false
}

// True if workflow was successfully completed
func (wi WorkflowInvocationStatus_Status) Successful() bool {
	return wi == WorkflowInvocationStatus_SUCCEEDED
}

func (fi FunctionInvocationStatus_Status) Finished() bool {
	for _, event := range functionFinalStates {
		if event == fi {
			return true
		}
	}
	return false
}

func (t Task) Requires(taskId string) bool {
	_, ok := t.Dependencies[taskId]
	return ok
}
