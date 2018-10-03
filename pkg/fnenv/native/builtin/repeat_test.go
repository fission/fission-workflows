package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/typedvalues/controlflow"
	"github.com/stretchr/testify/assert"
)

func TestFunctionRepeat_Invoke(t *testing.T) {
	taskToRepeat := &types.TaskSpec{
		FunctionRef: Noop,
		Inputs:      types.SingleDefaultInput(typedvalues.MustWrap("foo")),
	}

	repeatFn := &FunctionRepeat{}
	spec := &types.TaskInvocationSpec{
		Inputs: map[string]*typedvalues.TypedValue{
			RepeatInputDo:    typedvalues.MustWrap(taskToRepeat),
			RepeatInputTimes: typedvalues.MustWrap(10),
		},
	}
	result, err := repeatFn.Invoke(spec)
	assert.NoError(t, err)
	wf, err := controlflow.UnwrapWorkflow(result)
	assert.NoError(t, err)
	assert.Equal(t, 10, len(wf.Tasks))
}
