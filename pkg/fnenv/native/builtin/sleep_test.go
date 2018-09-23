package builtin

import (
	"testing"

	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestSleepFunctionString(t *testing.T) {
	start := time.Now()
	internalFunctionTest(t,
		&FunctionSleep{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*typedvalues.TypedValue{
				SleepInput: typedvalues.MustParse("1000ms"),
			},
		},
		nil)
	end := time.Now()
	assert.True(t, (end.UnixNano()-start.UnixNano()) > (time.Duration(900)*time.Millisecond).Nanoseconds())
}

func TestSleepFunctionInt(t *testing.T) {
	start := time.Now()
	internalFunctionTest(t,
		&FunctionSleep{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*typedvalues.TypedValue{
				SleepInput: typedvalues.MustParse(1000),
			},
		},
		nil)
	end := time.Now()
	assert.True(t, (end.UnixNano()-start.UnixNano()) > (time.Duration(900)*time.Millisecond).Nanoseconds())
}
