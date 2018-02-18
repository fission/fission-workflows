package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

func TestFunctionIfConsequentFlow(t *testing.T) {
	expectedTask := &types.TaskSpec{
		FunctionRef: "DoThisTask",
	}
	internalFunctionTest(t,
		&FunctionIf{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				IfInputCondition: typedvalues.UnsafeParse(true),
				IfInputThen:      typedvalues.UnsafeParse(expectedTask),
			},
		},
		expectedTask)
}

func TestFunctionIfAlternativeFlow(t *testing.T) {
	task := &types.TaskSpec{
		FunctionRef: "DoThisTask",
	}
	alternativeTask := &types.TaskSpec{
		FunctionRef: "DoThisOtherTask",
	}
	internalFunctionTest(t,
		&FunctionIf{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				IfInputCondition: typedvalues.UnsafeParse(false),
				IfInputThen:      typedvalues.UnsafeParse(task),
				IfInputElse:      typedvalues.UnsafeParse(alternativeTask),
			},
		},
		alternativeTask)
}

func TestFunctionIfLiteral(t *testing.T) {
	internalFunctionTest(t,
		&FunctionIf{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				IfInputCondition: typedvalues.UnsafeParse(true),
				IfInputThen:      typedvalues.UnsafeParse("foo"),
				IfInputElse:      typedvalues.UnsafeParse("bar"),
			},
		},
		"foo")
}

func TestFunctionIfMissingAlternative(t *testing.T) {
	internalFunctionTest(t,
		&FunctionIf{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				IfInputCondition: typedvalues.UnsafeParse(false),
				IfInputThen:      typedvalues.UnsafeParse("then"),
			},
		},
		nil)
}
