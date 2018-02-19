package typedvalues

import (
	"errors"
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/gogo/protobuf/proto"
)

const (
	TypeTask     = "task"
	TypeWorkflow = "workflow"
)

var (
	ErrInvalidType = errors.New("invalid type")
)

type ControlFlowParserFormatter struct {
}

func (cf *ControlFlowParserFormatter) Parse(i interface{}, allowedTypes ...string) (*types.TypedValue, error) {
	switch cf := i.(type) {
	case *types.Task:
		return ParseTask(cf), nil
	case *types.WorkflowSpec:
		return ParseWorkflow(cf), nil
	default:
		return nil, fmt.Errorf("%s: %s", ErrInvalidType, cf)
	}
}

func (cf *ControlFlowParserFormatter) Format(v *types.TypedValue) (interface{}, error) {
	switch v.Type {
	case TypeTask:
		return FormatTask(v)
	case TypeWorkflow:
		return FormatWorkflow(v)
	default:
		return nil, fmt.Errorf("%s: %s", ErrInvalidType, v)
	}
}

func ParseTask(task *types.Task) *types.TypedValue {
	data, err := proto.Marshal(task)
	if err != nil {
		panic(err)
	}
	return &types.TypedValue{
		Type:  FormatType(TypeTask),
		Value: data,
	}
}

func ParseWorkflow(task *types.WorkflowSpec) *types.TypedValue {
	data, err := proto.Marshal(task)
	if err != nil {
		panic(err)
	}
	return &types.TypedValue{
		Type:  FormatType(TypeWorkflow),
		Value: data,
	}
}

func FormatTask(v *types.TypedValue) (*types.Task, error) {
	t := &types.Task{}
	err := proto.Unmarshal(v.Value, t)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func FormatWorkflow(v *types.TypedValue) (*types.WorkflowSpec, error) {
	t := &types.WorkflowSpec{}
	err := proto.Unmarshal(v.Value, t)
	if err != nil {
		return nil, err
	}
	return t, nil
}
