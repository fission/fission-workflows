package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

func TestFunctionNoopPassInput(t *testing.T) {
	expected := "noopnoop"
	internalFunctionTest(t,
		&FunctionNoop{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*typedvalues.TypedValue{
				NoopInput: typedvalues.MustWrap(expected),
			},
		},
		expected)
}

func TestFunctionNoopEmpty(t *testing.T) {
	internalFunctionTest(t,
		&FunctionNoop{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*typedvalues.TypedValue{},
		},
		nil)
}

func TestFunctionNoopObject(t *testing.T) {
	internalFunctionTest(t,
		&FunctionNoop{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*typedvalues.TypedValue{
				"foo":           typedvalues.MustWrap(true),
				"bar":           typedvalues.MustWrap(false),
				types.InputMain: typedvalues.MustWrap("hello"),
			},
		},
		"hello")
}
