package builtin

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

func TestFunctionComposePassInput(t *testing.T) {
	expected := "ComposeCompose"
	internalFunctionTest(t,
		&FunctionCompose{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*typedvalues.TypedValue{
				ComposeInput: typedvalues.MustWrap(expected),
			},
		},
		expected)
}

func TestFunctionComposeEmpty(t *testing.T) {
	internalFunctionTest(t,
		&FunctionCompose{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*typedvalues.TypedValue{},
		},
		nil)
}

func TestFunctionComposeObject(t *testing.T) {
	internalFunctionTest(t,
		&FunctionCompose{},
		&types.TaskInvocationSpec{
			Inputs: map[string]*typedvalues.TypedValue{
				"foo":     typedvalues.MustWrap(true),
				"bar":     typedvalues.MustWrap(false),
				"default": typedvalues.MustWrap("hello"),
			},
		},
		map[string]interface{}{
			"foo":     true,
			"bar":     false,
			"default": "hello",
		})
}
