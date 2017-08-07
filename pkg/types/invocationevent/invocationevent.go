package invocationevent

import (
	"fmt"
	"github.com/fission/fission-workflow/pkg/types"
)

// Parse attempts to convert a string-based flag to the appropriate InvocationEvent.
func Parse(flag string) (types.InvocationEvent, error) {
	val, ok := types.InvocationEvent_value[flag]
	if !ok {
		return 0, fmt.Errorf("Unknown InvocationEvent '%s'", flag)
	}
	return types.InvocationEvent(val), nil
}
