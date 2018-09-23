package types

import (
	"errors"

	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	proto "github.com/golang/protobuf/proto"
)

type FlowType string

const (
	FlowTypeWorkflow FlowType = "workflow"
	FlowTypeTask     FlowType = "task"
	FlowTypeNone     FlowType = ""
)

// Flow is a generic data type to provide a common API to working with dynamic tasks and workflows
// If a flow contains both a task and a workflow, behavior is non-standard,
// but should in principle default to the task.
type Flow struct {
	task *TaskSpec
	wf   *WorkflowSpec
}

func (f *Flow) Type() FlowType {
	if f.task != nil {
		return FlowTypeTask
	}
	if f.wf != nil {
		return FlowTypeWorkflow
	}
	return FlowTypeNone
}

func (f *Flow) Input(key string, i typedvalues.TypedValue) {
	if f == nil {
		return
	}
	if f.task != nil {
		f.task.Input(key, &i)
	}
	if f.wf != nil {
		// TODO support parameters in workflow spec
	}
}

func (f *Flow) Proto() proto.Message {
	if f == nil {
		return nil
	}
	if f.task != nil {
		return f.task
	}
	return f.wf
}

func (f *Flow) Clone() *Flow {
	if f == nil {
		return nil
	}
	if f.task != nil {
		return FlowTask(proto.Clone(f.task).(*TaskSpec))
	}
	if f.wf != nil {
		return FlowWorkflow(proto.Clone(f.wf).(*WorkflowSpec))
	}
	return nil
}

func (f *Flow) Task() *TaskSpec {
	if f == nil {
		return nil
	}
	return f.task
}

func (f *Flow) Workflow() *WorkflowSpec {
	if f == nil {
		return nil
	}
	return f.wf
}

func (f *Flow) ApplyTask(fn func(t *TaskSpec)) {
	if f != nil && f.task != nil {
		fn(f.task)
	}
}

func (f *Flow) ApplyWorkflow(fn func(t *WorkflowSpec)) {
	if f != nil && f.wf != nil {
		fn(f.wf)
	}
}

func (f *Flow) IsEmpty() bool {
	return f.wf == nil && f.task == nil
}

func FlowTask(task *TaskSpec) *Flow {
	return &Flow{task: task}
}

func FlowWorkflow(wf *WorkflowSpec) *Flow {
	return &Flow{wf: wf}
}

func FlowInterface(i interface{}) (*Flow, error) {
	switch t := i.(type) {
	case *WorkflowSpec:
		return FlowWorkflow(t), nil
	case *TaskSpec:
		return FlowTask(t), nil
	default:
		return nil, errors.New("todo")
	}
}
