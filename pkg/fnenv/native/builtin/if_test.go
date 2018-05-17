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
				IfInputCondition: typedvalues.MustParse(true),
				IfInputThen:      typedvalues.MustParse(expectedTask),
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
				IfInputCondition: typedvalues.MustParse(false),
				IfInputThen:      typedvalues.MustParse(task),
				IfInputElse:      typedvalues.MustParse(alternativeTask),
			},
		},
		alternativeTask)
}

func TestFunctionIfLiteral(t *testing.T) {
	internalFunctionTest(t,
		&FunctionIf{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				IfInputCondition: typedvalues.MustParse(true),
				IfInputThen:      typedvalues.MustParse("foo"),
				IfInputElse:      typedvalues.MustParse("bar"),
			},
		},
		"foo")
}

func TestFunctionIfMissingAlternative(t *testing.T) {
	internalFunctionTest(t,
		&FunctionIf{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				IfInputCondition: typedvalues.MustParse(false),
				IfInputThen:      typedvalues.MustParse("then"),
			},
		},
		nil)
}
