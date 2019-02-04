// Package controlflow adds support for workflows and tasks (together "flows") to TypedValues.
//
// With the workflow engine supporting dynamic tasks (tasks outputting other tasks or workflows) this package offers a
// useful abstraction of this mechanism in the form  of a Flow. A flow is either a Workflow or Task,
// and is (like a task or workflow) wrappable into a TypedValue.
package controlflow

import (
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

var (
	ErrEmptyFlow = errors.New("flow is empty")
	ErrNotAFlow  = errors.New("value is not a flow")
)

func IsControlFlow(tv *typedvalues.TypedValue) bool {
	if tv != nil {
		vt := tv.ValueType()
		for _, valueType := range Types {
			if valueType == vt {
				return true
			}
		}
	}
	return false
}

func UnwrapControlFlow(tv *typedvalues.TypedValue) (*Flow, error) {
	i, err := typedvalues.Unwrap(tv)
	if err != nil {
		return nil, err
	}
	return FlowInterface(i)
}

func UnwrapTask(tv *typedvalues.TypedValue) (*types.TaskSpec, error) {
	flow, err := UnwrapControlFlow(tv)
	if err != nil {
		return nil, err
	}

	task := flow.GetTask()
	if task == nil {
		return nil, errors.Wrapf(typedvalues.ErrIllegalTypeAssertion, "failed to unwrap %s to task", tv.ValueType())
	}
	return task, nil
}

func UnwrapWorkflow(tv *typedvalues.TypedValue) (*types.WorkflowSpec, error) {
	flow, err := UnwrapControlFlow(tv)
	if err != nil {
		return nil, err
	}

	wf := flow.GetWorkflow()
	if wf == nil {
		return nil, errors.Wrapf(typedvalues.ErrIllegalTypeAssertion, "failed to unwrap %s to workflow", tv.ValueType())
	}
	return wf, nil
}

func (m *Flow) Type() FlowType {
	if m == nil {
		return FlowTypeNone
	}
	if m.Task != nil {
		return FlowTypeTask
	}
	if m.Workflow != nil {
		return FlowTypeWorkflow
	}
	return FlowTypeNone
}

func (m *Flow) Input(key string, i typedvalues.TypedValue) {
	if m == nil {
		return
	}
	if m.Task != nil {
		m.Task.Input(key, &i)
	}
	if m.Workflow != nil {
		// TODO support parameters in workflow spec
	}
}

func (m *Flow) Proto() proto.Message {
	if m == nil {
		return nil
	}
	if m.Task != nil {
		return m.Task
	}
	return m.Workflow
}

func (m *Flow) Clone() *Flow {
	if m == nil {
		return nil
	}
	if m.Task != nil {
		return FlowTask(proto.Clone(m.Task).(*types.TaskSpec))
	}
	if m.Workflow != nil {
		return FlowWorkflow(proto.Clone(m.Workflow).(*types.WorkflowSpec))
	}
	return nil
}

func (m *Flow) ApplyTask(fn func(t *types.TaskSpec)) {
	if m != nil && m.Task != nil {
		fn(m.Task)
	}
}

func (m *Flow) ApplyWorkflow(fn func(t *types.WorkflowSpec)) {
	if m != nil && m.Workflow != nil {
		fn(m.Workflow)
	}
}

func (m *Flow) IsEmpty() bool {
	return m.Workflow == nil && m.Task == nil
}

func FlowTask(task *types.TaskSpec) *Flow {
	return &Flow{Task: task}
}

func FlowWorkflow(workflow *types.WorkflowSpec) *Flow {
	return &Flow{Workflow: workflow}
}

func FlowInterface(i interface{}) (*Flow, error) {
	if i == nil {
		return nil, ErrEmptyFlow
	}
	switch t := i.(type) {
	case *types.WorkflowSpec:
		return FlowWorkflow(t), nil
	case *types.TaskSpec:
		return FlowTask(t), nil
	case *Flow:
		return t, nil
	default:
		return nil, ErrNotAFlow
	}
}

// TODO move to more appropriate package
func ResolveTaskOutput(taskID string, invocation *types.WorkflowInvocation) *typedvalues.TypedValue {
	val, ok := invocation.Status.Tasks[taskID]
	if !ok {
		return nil
	}

	output := val.Status.Output
	if IsControlFlow(output) {
		for outputTaskID, outputTask := range invocation.Status.DynamicTasks {
			if dep, ok := outputTask.Spec.Requires[taskID]; ok &&
				dep.Type == types.TaskDependencyParameters_DYNAMIC_OUTPUT {
				return ResolveTaskOutput(outputTaskID, invocation)
			}
		}
	}
	return output
}

func ResolveTaskOutputHeaders(taskID string, invocation *types.WorkflowInvocation) *typedvalues.TypedValue {
	val, ok := invocation.Status.Tasks[taskID]
	if !ok {
		return nil
	}
	return val.Status.OutputHeaders
}
