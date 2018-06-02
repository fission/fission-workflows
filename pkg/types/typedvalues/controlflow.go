package typedvalues

import (
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

const (
	TypeTask     ValueType = "task"
	TypeWorkflow ValueType = "workflow"
)

type ControlFlowParserFormatter struct {
}

func (pf *ControlFlowParserFormatter) Accepts() []ValueType {
	return []ValueType{
		TypeTask,
		TypeWorkflow,
	}
}

func (pf *ControlFlowParserFormatter) Parse(ctx Parser, i interface{}) (*types.TypedValue, error) {
	switch cf := i.(type) {
	case *types.TaskSpec:
		return ParseTask(cf), nil
	case *types.WorkflowSpec:
		return ParseWorkflow(cf), nil
	default:
		return nil, TypedValueErr{
			src: i,
			err: ErrUnsupportedType,
		}
	}
}

func (pf *ControlFlowParserFormatter) Format(ctx Formatter, v *types.TypedValue) (interface{}, error) {
	switch ValueType(v.Type) {
	case TypeTask:
		return FormatTask(v)
	case TypeWorkflow:
		return FormatWorkflow(v)
	default:
		return nil, TypedValueErr{
			src: v,
			err: ErrUnsupportedType,
		}
	}
}

func ParseTask(task *types.TaskSpec) *types.TypedValue {
	data, err := proto.Marshal(task)
	if err != nil {
		panic(err)
	}
	return &types.TypedValue{
		Type:  string(TypeTask),
		Value: data,
	}
}

func ParseWorkflow(wf *types.WorkflowSpec) *types.TypedValue {
	data, err := proto.Marshal(wf)
	if err != nil {
		panic(err)
	}
	return &types.TypedValue{
		Type:  string(TypeWorkflow),
		Value: data,
	}
}

func FormatTask(v *types.TypedValue) (*types.TaskSpec, error) {
	t := &types.TaskSpec{}
	err := proto.Unmarshal(v.Value, t)
	if err != nil {
		return nil, TypedValueErr{
			src: v,
			err: errors.Wrap(err, "failed to format task"),
		}
	}
	return t, nil
}

func FormatWorkflow(v *types.TypedValue) (*types.WorkflowSpec, error) {
	t := &types.WorkflowSpec{}
	err := proto.Unmarshal(v.Value, t)
	if err != nil {
		return nil, TypedValueErr{
			src: v,
			err: errors.Wrap(err, "failed to format workflow"),
		}
	}
	return t, nil
}

func IsControlFlow(v ValueType) bool {
	return v == TypeTask || v == TypeWorkflow
}

func FormatControlFlow(v *types.TypedValue) (*Flow, error) {
	switch ValueType(v.Type) {
	case TypeTask:
		t, err := FormatTask(v)
		if err != nil {
			return nil, err
		}
		return FlowTask(t), nil
	case TypeWorkflow:
		wf, err := FormatWorkflow(v)
		if err != nil {
			return nil, err
		}
		return FlowWorkflow(wf), nil
	default:
		return nil, ErrUnsupportedType
	}
}

func ParseControlFlow(i interface{}) (*types.TypedValue, error) {
	switch t := i.(type) {
	case *types.TaskSpec:
		return ParseTask(t), nil
	case *types.WorkflowSpec:
		return ParseWorkflow(t), nil
	case *Flow:
		return ParseControlFlow(t.Proto())
	default:
		return nil, ErrUnsupportedType
	}
}

// Flow is a generic data type to provide a common API to working with dynamic tasks and workflows
type Flow struct {
	task *types.TaskSpec
	wf   *types.WorkflowSpec
}

func (f *Flow) Input(key string, i types.TypedValue) {
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
		return FlowTask(proto.Clone(f.task).(*types.TaskSpec))
	}
	if f.wf != nil {
		return FlowWorkflow(proto.Clone(f.wf).(*types.WorkflowSpec))
	}
	return nil
}

func (f *Flow) Task() *types.TaskSpec {
	if f == nil {
		return nil
	}
	return f.task
}

func (f *Flow) Workflow() *types.WorkflowSpec {
	if f == nil {
		return nil
	}
	return f.wf
}

func (f *Flow) ApplyTask(fn func(t *types.TaskSpec)) {
	if f != nil && f.task != nil {
		fn(f.task)
	}
}

func (f *Flow) ApplyWorkflow(fn func(t *types.WorkflowSpec)) {
	if f != nil && f.wf != nil {
		fn(f.wf)
	}
}

func FlowTask(task *types.TaskSpec) *Flow {
	return &Flow{task: task}
}

func FlowWorkflow(wf *types.WorkflowSpec) *Flow {
	return &Flow{wf: wf}
}

func FlowInterface(i interface{}) (*Flow, error) {
	switch t := i.(type) {
	case *types.WorkflowSpec:
		return FlowWorkflow(t), nil
	case *types.TaskSpec:
		return FlowTask(t), nil
	default:
		return nil, ErrUnsupportedType
	}
}
