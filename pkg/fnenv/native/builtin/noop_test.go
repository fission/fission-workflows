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
			Inputs: map[string]*types.TypedValue{
				NoopInput: typedvalues.UnsafeParse(expected),
			},
		},
		expected)
}

func TestFunctionNoopEmpty(t *testing.T) {
	internalFunctionTest(t,
		&FunctionNoop{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{},
		},
		nil)
}

func TestFunctionNoopObject(t *testing.T) {
	internalFunctionTest(t,
		&FunctionNoop{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				"foo":            typedvalues.UnsafeParse(true),
				"bar":            typedvalues.UnsafeParse(false),
				types.INPUT_MAIN: typedvalues.UnsafeParse("hello"),
			},
		},
		"hello")
}
