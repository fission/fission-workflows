package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestFunctionForeach_Invoke(t *testing.T) {
	foreachElements := []interface{}{1, 2, 3, 4, "foo"}
	out, err := (&FunctionForeach{}).Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*types.TypedValue{
			ForeachInputForeach: typedvalues.MustParse(foreachElements),
			ForeachInputDo: typedvalues.MustParse(&types.TaskSpec{
				FunctionRef: Noop,
			}),
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, typedvalues.TypeWorkflow, typedvalues.ValueType(out.Type))

	wf, err := typedvalues.FormatWorkflow(out)
	assert.NoError(t, err)
	assert.Equal(t, len(foreachElements)+1, len(wf.Tasks)) // + 1 for the noop-task in the foreach loop.
	assert.NotNil(t, wf.Tasks["do_0"])
	assert.Equal(t, foreachElements[0], int(typedvalues.MustFormat(wf.Tasks["do_0"].Inputs["element"]).(float64)))
}
