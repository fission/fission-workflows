package builtin

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/fission/fission-workflows/pkg/fnenv/native"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

var DefaultBuiltinFunctions = map[string]native.InternalFunction{
	If:         &FunctionIf{},
	Noop:       &FunctionNoop{},
	"nop":      &FunctionNoop{}, // nop is an alias for 'noop'
	Compose:    &FunctionCompose{},
	Sleep:      &FunctionSleep{},
	Repeat:     &FunctionRepeat{},
	Javascript: NewFunctionJavascript(),
	Fail:       &FunctionFail{},
	Http:       &FunctionHttp{},
	Foreach:    &FunctionForeach{},
	Switch:     &FunctionSwitch{},
	While:      &FunctionWhile{},
}

// ensureInput verifies that the input for the given key exists and is of one of the provided types.
func ensureInput(inputs map[string]*types.TypedValue, key string, validTypes ...string) (*types.TypedValue, error) {

	i, ok := inputs[key]
	if !ok {
		return nil, fmt.Errorf("input '%s' is not set", key)
	}

	if len(validTypes) > 0 {
		var valid bool
		for _, t := range validTypes {
			if strings.Contains(i.Type, t) {
				valid = true
				break
			}
		}
		if !valid {
			return nil, fmt.Errorf("input '%s' is not a valid type (expected: %s)", key,
				strings.Join(validTypes, "|"))
		}
	}

	return i, nil
}

func internalFunctionTest(t *testing.T, fn native.InternalFunction, input *types.TaskInvocationSpec, expected interface{}) {
	output, err := fn.Invoke(input)
	if err != nil {
		t.Fatal(err)
	}

	outputtedTask, err := typedvalues.Format(output)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(outputtedTask, expected) {
		t.Errorf("Output '%v' does not match expected output '%v'", outputtedTask, expected)
	}
}

// getFirstDefinedTypedValue returns the first input and key of the inputs argument that matches a field in fields.
// For example, given inputs { a : b, c : d }, getFirstDefinedTypedValue(inputs, z, x, c, a) would return (c, d)
func getFirstDefinedTypedValue(inputs map[string]*types.TypedValue, fields ...string) (string, *types.TypedValue) {
	var result *types.TypedValue
	var key string
	for _, key = range fields {
		val, ok := inputs[key]
		if ok {
			result = val
			break
		}
	}
	return key, result
}
