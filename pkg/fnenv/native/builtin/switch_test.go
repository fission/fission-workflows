package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestFunctionSwitch_Invoke(t *testing.T) {
	fn := &FunctionSwitch{}
	spec := &types.TaskInvocationSpec{
		Inputs: map[string]*types.TypedValue{
			SwitchInputCondition:   typedvalues.UnsafeParse("case1"),
			"case1":                typedvalues.UnsafeParse("case1"),
			SwitchInputDefaultCase: typedvalues.UnsafeParse("default"),
		},
	}
	out, err := fn.Invoke(spec)
	assert.NoError(t, err)
	assert.Equal(t, spec.Inputs["case1"], out)
}

func TestFunctionSwitch_InvokeDefaultCase(t *testing.T) {
	fn := &FunctionSwitch{}
	spec := &types.TaskInvocationSpec{
		Inputs: map[string]*types.TypedValue{
			SwitchInputCondition:   typedvalues.UnsafeParse("case1"),
			"case2":                typedvalues.UnsafeParse("case2"),
			SwitchInputDefaultCase: typedvalues.UnsafeParse("default"),
		},
	}
	out, err := fn.Invoke(spec)
	assert.NoError(t, err)
	assert.Equal(t, spec.Inputs[SwitchInputDefaultCase], out)
}

func TestFunctionSwitch_InvokeNoCase(t *testing.T) {
	fn := &FunctionSwitch{}
	spec := &types.TaskInvocationSpec{
		Inputs: map[string]*types.TypedValue{
			SwitchInputCondition: typedvalues.UnsafeParse("case1"),
			"case2":              typedvalues.UnsafeParse("case2"),
		},
	}
	out, err := fn.Invoke(spec)
	assert.NoError(t, err)
	assert.Nil(t, out)
}

func TestFunctionSwitch_InvokeNoSwitch(t *testing.T) {
	fn := &FunctionSwitch{}
	spec := &types.TaskInvocationSpec{
		Inputs: map[string]*types.TypedValue{
			"case2": typedvalues.UnsafeParse("case2"),
		},
	}
	out, err := fn.Invoke(spec)
	assert.Error(t, err)
	assert.Nil(t, out)
}
