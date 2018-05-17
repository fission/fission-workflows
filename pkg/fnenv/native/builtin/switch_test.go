package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestFunctionSwitch_Invoke(t *testing.T) {
	val := "case1Val"
	fn := &FunctionSwitch{}
	spec := &types.TaskInvocationSpec{
		Inputs: map[string]*types.TypedValue{
			SwitchInputCondition: typedvalues.ParseString("case1"),
			SwitchInputCases: typedvalues.MustParse([]interface{}{
				switchCase("case1", val),
			}),
			SwitchInputDefaultCase: typedvalues.MustParse("default"),
		},
	}
	out, err := fn.Invoke(spec)
	assert.NoError(t, err)
	assert.Equal(t, "case1Val", typedvalues.MustFormat(out))
}

func TestFunctionSwitch_InvokeDefaultCase(t *testing.T) {
	fn := &FunctionSwitch{}
	spec := &types.TaskInvocationSpec{
		Inputs: map[string]*types.TypedValue{
			SwitchInputCondition: typedvalues.ParseString("case1"),
			SwitchInputCases: typedvalues.MustParse([]interface{}{
				switchCase("case2", "case2"),
			}),
			SwitchInputDefaultCase: typedvalues.MustParse("default"),
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
			SwitchInputCondition: typedvalues.MustParse("case1"),
			"case2":              typedvalues.MustParse("case2"),
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
			"case2": typedvalues.MustParse("case2"),
		},
	}
	out, err := fn.Invoke(spec)
	assert.Error(t, err)
	assert.Nil(t, out)
}
