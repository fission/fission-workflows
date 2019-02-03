package apiserver

import "google.golang.org/grpc"

type Client struct {
	Admin      AdminAPIClient
	Invocation WorkflowInvocationAPIClient
	Workflow   WorkflowAPIClient
}

func NewClient(conn *grpc.ClientConn) *Client {
	return &Client{
		Admin:      NewAdminAPIClient(conn),
		Invocation: NewWorkflowInvocationAPIClient(conn),
		Workflow:   NewWorkflowAPIClient(conn),
	}
}
