package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestFunctionWhile_Invoke(t *testing.T) {
	out, err := (&FunctionWhile{}).Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*types.TypedValue{
			WhileInputExpr:  typedvalues.UnsafeParse(false),
			WhileInputLimit: typedvalues.UnsafeParse(10),
			"count":         typedvalues.UnsafeParse(4),
			WhileInputDelay: typedvalues.UnsafeParse("1h"),
			WhileInputAction: typedvalues.UnsafeParse(&types.TaskSpec{
				FunctionRef: Noop,
			}),
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, typedvalues.TypeWorkflow, out.Type)
}

func TestFunctionWhile_InvokeCompletedInitial(t *testing.T) {
	out, err := (&FunctionWhile{}).Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*types.TypedValue{
			WhileInputExpr:  typedvalues.UnsafeParse(true),
			WhileInputLimit: typedvalues.UnsafeParse(10),
			"count":         typedvalues.UnsafeParse(4),
			WhileInputDelay: typedvalues.UnsafeParse("1h"),
			WhileInputAction: typedvalues.UnsafeParse(&types.TaskSpec{
				FunctionRef: Noop,
			}),
		},
	})
	assert.NoError(t, err)
	assert.Nil(t, out)
}

func TestFunctionWhile_InvokeCompleted(t *testing.T) {
	prev := typedvalues.UnsafeParse("prev result")
	out, err := (&FunctionWhile{}).Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*types.TypedValue{
			WhileInputExpr:  typedvalues.UnsafeParse(true),
			WhileInputLimit: typedvalues.UnsafeParse(10),
			"count":         typedvalues.UnsafeParse(4),
			WhileInputDelay: typedvalues.UnsafeParse("1h"),
			WhileInputAction: typedvalues.UnsafeParse(&types.TaskSpec{
				FunctionRef: Noop,
			}),
			"prev": prev,
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, prev, out)
}

func TestFunctionWhile_Invoke_LimitExceeded(t *testing.T) {
	out, err := (&FunctionWhile{}).Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*types.TypedValue{
			WhileInputExpr:  typedvalues.UnsafeParse(false),
			WhileInputLimit: typedvalues.UnsafeParse(10),
			"count":         typedvalues.UnsafeParse(11),
			WhileInputDelay: typedvalues.UnsafeParse("1h"),
			WhileInputAction: typedvalues.UnsafeParse(&types.TaskSpec{
				FunctionRef: Noop,
			}),
		},
	})
	assert.EqualError(t, err, ErrLimitExceeded.Error())
	assert.Nil(t, out)
}
