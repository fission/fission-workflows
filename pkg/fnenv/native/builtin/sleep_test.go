package builtin

import (
	"testing"

	"time"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestSleepFunction(t *testing.T) {
	start := time.Now()
	internalFunctionTest(t,
		&FunctionSleep{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				SLEEP_INPUT_MS: parseUnsafe("1000"),
			},
		},
		nil)
	end := time.Now()
	assert.True(t, (end.UnixNano()-start.UnixNano()) > (time.Duration(900)*time.Millisecond).Nanoseconds())
}
