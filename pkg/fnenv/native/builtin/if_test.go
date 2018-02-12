package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
)

func TestFunctionIfConsequentFlow(t *testing.T) {
	expectedTask := &types.Task{
		FunctionRef: "DoThisTask",
	}
	internalFunctionTest(t,
		&FunctionIf{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				IfInputCondition: parseUnsafe(true),
				IfInputThen:      parseUnsafe(expectedTask),
			},
		},
		expectedTask)
}

func TestFunctionIfAlternativeFlow(t *testing.T) {
	task := &types.Task{
		FunctionRef: "DoThisTask",
	}
	alternativeTask := &types.Task{
		FunctionRef: "DoThisOtherTask",
	}
	internalFunctionTest(t,
		&FunctionIf{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				IfInputCondition: parseUnsafe(false),
				IfInputThen:      parseUnsafe(task),
				IfInputElse:      parseUnsafe(alternativeTask),
			},
		},
		alternativeTask)
}

func TestFunctionIfLiteral(t *testing.T) {
	internalFunctionTest(t,
		&FunctionIf{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				IfInputCondition: parseUnsafe(true),
				IfInputThen:      parseUnsafe("foo"),
				IfInputElse:      parseUnsafe("bar"),
			},
		},
		"foo")
}

func TestFunctionIfMissingAlternative(t *testing.T) {
	internalFunctionTest(t,
		&FunctionIf{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				IfInputCondition: parseUnsafe(false),
				IfInputThen:      parseUnsafe("then"),
			},
		},
		nil)
}
