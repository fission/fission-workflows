package builtin

import (
	"testing"

	"github.com/fission/fission-workflow/pkg/types"
)

func TestFunctionIfConsequentFlow(t *testing.T) {
	expectedTask := &types.Task{
		FunctionRef: "DoThisTask",
	}
	internalFunctionTest(t,
		&FunctionIf{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				IF_INPUT_CONDITION:  parseUnsafe(true),
				IF_INPUT_CONSEQUENT: parseUnsafe(expectedTask),
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
				IF_INPUT_CONDITION:   parseUnsafe(false),
				IF_INPUT_CONSEQUENT:  parseUnsafe(task),
				IF_INPUT_ALTERNATIVE: parseUnsafe(alternativeTask),
			},
		},
		alternativeTask)
}
