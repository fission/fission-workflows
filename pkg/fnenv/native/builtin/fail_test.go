package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestFunctionFail_InvokeEmpty(t *testing.T) {
	fn := &FunctionFail{}
	out, err := fn.Invoke(&types.TaskInvocationSpec{})
	assert.Nil(t, out)
	assert.EqualError(t, err, typedvalues.MustFormat(defaultErrMsg).(string))

}

func TestFunctionFail_InvokeString(t *testing.T) {
	fn := &FunctionFail{}
	errMsg := "custom error message"
	out, err := fn.Invoke(&types.TaskInvocationSpec{
		Inputs: typedvalues.Input(errMsg),
	})
	assert.Nil(t, out)
	assert.EqualError(t, err, errMsg)
}
