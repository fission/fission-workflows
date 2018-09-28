package controlflow

import (
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/proto"
)

type FlowType string

var (
	TypeTask         string
	TypeWorkflow     string
	TypeFlow         string
	FlowTypeTask     FlowType
	FlowTypeWorkflow FlowType
	FlowTypeNone     FlowType
	Types            []string
)

// Note: ensure that this file is lexically after the generated Protobuf messages because of package initialization
// order.
func init() {
	TypeTask = proto.MessageName(&types.TaskSpec{})
	TypeWorkflow = proto.MessageName(&types.WorkflowSpec{})
	TypeFlow = proto.MessageName(&Flow{})
	FlowTypeWorkflow = FlowType(TypeWorkflow)
	FlowTypeTask = FlowType(TypeTask)
	FlowTypeNone = FlowType("")
	Types = []string{
		TypeTask,
		TypeWorkflow,
		TypeFlow,
	}
}
