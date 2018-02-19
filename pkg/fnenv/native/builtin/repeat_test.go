package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestFunctionRepeat_Invoke(t *testing.T) {
	taskToRepeat := &types.TaskSpec{
		FunctionRef: Noop,
		Inputs: map[string]*types.TypedValue{
			types.INPUT_MAIN: typedvalues.UnsafeParse("foo"),
		},
	}

	repeatFn := &FunctionRepeat{}
	spec := &types.TaskInvocationSpec{
		Inputs: map[string]*types.TypedValue{
			RepeatInputDo:    typedvalues.UnsafeParse(taskToRepeat),
			RepeatInputTimes: typedvalues.UnsafeParse(10),
		},
	}
	result, err := repeatFn.Invoke(spec)
	assert.NoError(t, err)
	wf, err := typedvalues.FormatWorkflow(result)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(wf.Tasks))
}
