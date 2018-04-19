package typedvalues

import (
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/gogo/protobuf/proto"
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

func ParseWorkflow(task *types.WorkflowSpec) *types.TypedValue {
	data, err := proto.Marshal(task)
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
			err: err,
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
			err: err,
		}
	}
	return t, nil
}

func IsControlFlow(v ValueType) bool {
	return v == TypeTask || v == TypeWorkflow
}
