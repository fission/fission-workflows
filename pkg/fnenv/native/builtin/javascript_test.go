package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestFunctionJavascript_InvokeMap(t *testing.T) {
	spec := &types.TaskInvocationSpec{
		Inputs: map[string]*typedvalues.TypedValue{
			JavascriptInputArgs: typedvalues.MustWrap(map[string]interface{}{
				"left":  2,
				"right": 5,
			}),
			JavascriptInputExpr: typedvalues.MustWrap("left * right"),
		},
	}

	js := NewFunctionJavascript()
	tv, err := js.Invoke(spec)
	assert.NoError(t, err)
	assert.Equal(t, 10, int(typedvalues.MustUnwrap(tv).(float64)))
}

func TestFunctionJavascript_Invoke(t *testing.T) {
	spec := &types.TaskInvocationSpec{
		Inputs: map[string]*typedvalues.TypedValue{
			JavascriptInputArgs: typedvalues.MustWrap(10),
			JavascriptInputExpr: typedvalues.MustWrap("arg * 2"),
		},
	}

	js := NewFunctionJavascript()
	tv, err := js.Invoke(spec)
	assert.NoError(t, err)
	assert.Equal(t, 20, int(typedvalues.MustUnwrap(tv).(float64)))
}
