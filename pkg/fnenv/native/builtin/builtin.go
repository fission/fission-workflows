package builtin

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/fission/fission-workflows/pkg/fnenv/native"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/golang/protobuf/proto"
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
	Http:       NewFunctionHTTP(),
	Foreach:    &FunctionForeach{},
	Switch:     &FunctionSwitch{},
	While:      &FunctionWhile{},
}

// ensureInput verifies that the input for the given key exists and is of one of the provided types.
func ensureInput(inputs map[string]*typedvalues.TypedValue, key string, validTypes ...string) (*typedvalues.TypedValue, error) {

	tv, ok := inputs[key]
	if !ok {
		return nil, fmt.Errorf("input '%s' is not set", key)
	}

	if len(validTypes) == 0 {
		return tv, nil
	}
	var found bool
	for _, validType := range validTypes {
		if validType == tv.ValueType() {
			found = true
			break
		}
	}

	if !found {
		return nil, fmt.Errorf("input '%s' is not a validType type (expected: %v, was: %T)", key, validTypes, tv.ValueType())
	}
	return tv, nil
}

func internalFunctionTest(t *testing.T, fn native.InternalFunction, input *types.TaskInvocationSpec, expected interface{}) {
	output, err := fn.Invoke(input)
	if err != nil {
		t.Fatal(err)
	}

	outputtedTask, err := typedvalues.Unwrap(output)
	if err != nil {
		t.Fatal(err)
	}

	if _, ok := outputtedTask.(proto.Message); ok {
		util.AssertProtoEqual(t, outputtedTask.(proto.Message), expected.(proto.Message))
	} else if !reflect.DeepEqual(outputtedTask, expected) {
		t.Errorf("Output '%v' (%T) does not match expected output '%v' (%T)", outputtedTask, outputtedTask, expected, expected)
	}
}

// getFirstDefinedTypedValue returns the first input and key of the inputs argument that matches a field in fields.
// For example, given inputs { a : b, c : d }, getFirstDefinedTypedValue(inputs, z, x, c, a) would return (c, d)
func getFirstDefinedTypedValue(inputs map[string]*typedvalues.TypedValue, fields ...string) (string, *typedvalues.TypedValue) {
	var result *typedvalues.TypedValue
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
