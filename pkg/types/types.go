package types

// Types other than specified in protobuf
const (
	SUBJECT_INVOCATION = "invocation"
	SUBJECT_WORKFLOW   = "workflows"
	INPUT_MAIN         = "*"
)

// InvocationEvent
var finalStates = []WorkflowInvocationStatus_Status{
	WorkflowInvocationStatus_ABORTED,
	WorkflowInvocationStatus_SUCCEEDED,
	WorkflowInvocationStatus_FAILED,
}

func (wi WorkflowInvocationStatus_Status) Finished() bool {
	for _, event := range finalStates {
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
