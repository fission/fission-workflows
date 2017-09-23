package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
)

func TestFunctionComposePassInput(t *testing.T) {
	expected := "ComposeCompose"
	internalFunctionTest(t,
		&FunctionCompose{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				COMPOSE_INPUT: parseUnsafe(expected),
			},
		},
		expected)
}

func TestFunctionComposeEmpty(t *testing.T) {
	internalFunctionTest(t,
		&FunctionCompose{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{},
		},
		nil)
}

func TestFunctionComposeObject(t *testing.T) {
	internalFunctionTest(t,
		&FunctionCompose{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*types.TypedValue{
				"foo":     parseUnsafe(true),
				"bar":     parseUnsafe(false),
				"default": parseUnsafe("hello"),
			},
		},
		map[string]interface{}{
			"foo":     true,
			"bar":     false,
			"default": "hello",
		})
}
