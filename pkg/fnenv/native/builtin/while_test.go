package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestFunctionWhile_Invoke(t *testing.T) {
	out, err := (&FunctionWhile{}).Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*typedvalues.TypedValue{
			WhileInputExpr:  typedvalues.MustParse(true).SetLabel("src", "{}"),
			WhileInputLimit: typedvalues.MustParse(10),
			"_count":        typedvalues.MustParse(4),
			WhileInputDelay: typedvalues.MustParse("1h"),
			WhileInputAction: typedvalues.MustParse(&types.TaskSpec{
				FunctionRef: Noop,
			}),
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, typedvalues.TypeWorkflow, out.Type)
}

func TestFunctionWhile_InvokeCompletedInitial(t *testing.T) {
	out, err := (&FunctionWhile{}).Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*typedvalues.TypedValue{
			WhileInputExpr:  typedvalues.MustParse(false).SetLabel("src", "{}"),
			WhileInputLimit: typedvalues.MustParse(10),
			"_count":        typedvalues.MustParse(4),
			WhileInputDelay: typedvalues.MustParse("1h"),
			WhileInputAction: typedvalues.MustParse(&types.TaskSpec{
				FunctionRef: Noop,
			}),
		},
	})
	assert.NoError(t, err)
	assert.Nil(t, out)
}

func TestFunctionWhile_InvokeCompleted(t *testing.T) {
	prev := typedvalues.MustParse("prev result")
	out, err := (&FunctionWhile{}).Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*typedvalues.TypedValue{
			WhileInputExpr:  typedvalues.MustParse(false).SetLabel("src", "{}"),
			WhileInputLimit: typedvalues.MustParse(10),
			"_count":        typedvalues.MustParse(4),
			WhileInputDelay: typedvalues.MustParse("1h"),
			WhileInputAction: typedvalues.MustParse(&types.TaskSpec{
				FunctionRef: Noop,
			}),
			"_prev": prev,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, prev, out)
}

func TestFunctionWhile_Invoke_LimitExceeded(t *testing.T) {
	out, err := (&FunctionWhile{}).Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*typedvalues.TypedValue{
			WhileInputExpr:  typedvalues.MustParse(true).SetLabel("src", "{}"),
			WhileInputLimit: typedvalues.MustParse(10),
			"_count":        typedvalues.MustParse(11),
			WhileInputDelay: typedvalues.MustParse("1h"),
			WhileInputAction: typedvalues.MustParse(&types.TaskSpec{
				FunctionRef: Noop,
			}),
		},
	})
	assert.EqualError(t, err, ErrLimitExceeded.Error())
	assert.Nil(t, out)
}
