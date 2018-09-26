package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/typedvalues/controlflow"
	"github.com/stretchr/testify/assert"
)

func TestFunctionWhile_Invoke(t *testing.T) {
	out, err := (&FunctionWhile{}).Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*typedvalues.TypedValue{
			WhileInputExpr:  typedvalues.MustWrap(true).SetMetadata("src", "{}"),
			WhileInputLimit: typedvalues.MustWrap(10),
			"_count":        typedvalues.MustWrap(4),
			WhileInputDelay: typedvalues.MustWrap("1h"),
			WhileInputAction: typedvalues.MustWrap(&types.TaskSpec{
				FunctionRef: Noop,
			}),
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, controlflow.TypeWorkflow, out.ValueType())
}

func TestFunctionWhile_InvokeCompletedInitial(t *testing.T) {
	out, err := (&FunctionWhile{}).Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*typedvalues.TypedValue{
			WhileInputExpr:  typedvalues.MustWrap(false).SetMetadata("src", "{}"),
			WhileInputLimit: typedvalues.MustWrap(10),
			"_count":        typedvalues.MustWrap(4),
			WhileInputDelay: typedvalues.MustWrap("1h"),
			WhileInputAction: typedvalues.MustWrap(&types.TaskSpec{
				FunctionRef: Noop,
			}),
		},
	})
	assert.NoError(t, err)
	assert.Nil(t, out)
}

func TestFunctionWhile_InvokeCompleted(t *testing.T) {
	prev := typedvalues.MustWrap("prev result")
	out, err := (&FunctionWhile{}).Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*typedvalues.TypedValue{
			WhileInputExpr:  typedvalues.MustWrap(false).SetMetadata("src", "{}"),
			WhileInputLimit: typedvalues.MustWrap(10),
			"_count":        typedvalues.MustWrap(4),
			WhileInputDelay: typedvalues.MustWrap("1h"),
			WhileInputAction: typedvalues.MustWrap(&types.TaskSpec{
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
			WhileInputExpr:  typedvalues.MustWrap(true).SetMetadata("src", "{}"),
			WhileInputLimit: typedvalues.MustWrap(10),
			"_count":        typedvalues.MustWrap(11),
			WhileInputDelay: typedvalues.MustWrap("1h"),
			WhileInputAction: typedvalues.MustWrap(&types.TaskSpec{
				FunctionRef: Noop,
			}),
		},
	})
	assert.EqualError(t, err, ErrLimitExceeded.Error())
	assert.Nil(t, out)
}
