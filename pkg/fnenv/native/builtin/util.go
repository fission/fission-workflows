package builtin

import (
	"fmt"

	"reflect"
	"testing"

	"github.com/fission/fission-workflows/pkg/fnenv/native"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

var DefaultBuiltinFunctions = map[string]native.InternalFunction{
	"if":      &FunctionIf{},
	"noop":    &FunctionNoop{},
	"compose": &FunctionCompose{},
	"sleep":   &FunctionSleep{},
	"scope":   &FunctionScope{},
}

// Utils
func verifyInput(inputs map[string]*types.TypedValue, key string, expType string) (*types.TypedValue, error) {

	i, ok := inputs[key]
	if !ok {
		return nil, fmt.Errorf("input '%s' is not set", key)
	}

	//if !typedvalues.(i.Type, expType) {
	//	return nil, fmt.Errorf("Input '%s' is not of expected type '%s' (was '%s')", key, expType, i.Type)
	//}
	return i, nil
}

func parseUnsafe(i interface{}) *types.TypedValue {
	val, err := typedvalues.Parse(i)
	if err != nil {
		panic(err)
	}
	return val
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
