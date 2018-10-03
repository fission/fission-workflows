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
			Inputs: map[string]*typedvalues.TypedValue{
				IfInputCondition: typedvalues.MustWrap(true),
				IfInputThen:      typedvalues.MustWrap(expectedTask),
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
			Inputs: map[string]*typedvalues.TypedValue{
				IfInputCondition: typedvalues.MustWrap(false),
				IfInputThen:      typedvalues.MustWrap(task),
				IfInputElse:      typedvalues.MustWrap(alternativeTask),
			},
		},
		alternativeTask)
}

func TestFunctionIfLiteral(t *testing.T) {
	internalFunctionTest(t,
		&FunctionIf{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*typedvalues.TypedValue{
				IfInputCondition: typedvalues.MustWrap(true),
				IfInputThen:      typedvalues.MustWrap("foo"),
				IfInputElse:      typedvalues.MustWrap("bar"),
			},
		},
		"foo")
}

func TestFunctionIfMissingAlternative(t *testing.T) {
	internalFunctionTest(t,
		&FunctionIf{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*typedvalues.TypedValue{
				IfInputCondition: typedvalues.MustWrap(false),
				IfInputThen:      typedvalues.MustWrap("then"),
			},
		},
		nil)
}
