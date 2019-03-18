package events

var InvocationTerminalEvents = []string{
	EventInvocationCompleted,
	EventInvocationCanceled,
	EventInvocationFailed,
}

var WorkflowTerminalEvents = []string{
	EventWorkflowDeleted,
}
