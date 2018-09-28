package controlflow

import (
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/proto"
)

var (
	TypeTask     string
	TypeWorkflow string
	TypeFlow     string
	Types        []string
)

// Note: ensure that this file is lexically after the generated Protobuf messages because of package initialization
// order.
func init() {
	TypeTask = proto.MessageName(&types.TaskSpec{})
	TypeWorkflow = proto.MessageName(&types.WorkflowSpec{})
	TypeFlow = proto.MessageName(&Flow{})
	Types = []string{
		TypeTask,
		TypeWorkflow,
		TypeFlow,
	}
}
