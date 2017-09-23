package builtin

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

const (
	IF_INPUT_CONDITION   = "if"
	IF_INPUT_CONSEQUENT  = "then"
	IF_INPUT_ALTERNATIVE = "else"
)

type FunctionIf struct{}

func (fn *FunctionIf) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {

	// Verify and get condition
	expr, err := verifyInput(spec.GetInputs(), IF_INPUT_CONDITION, typedvalues.FormatType(typedvalues.FORMAT_JSON, typedvalues.TYPE_BOOL))
	if err != nil {
		return nil, err
	}

	// Get consequent alternative, if one of those does not exist, that is fine.
	consequent := spec.GetInputs()[IF_INPUT_CONSEQUENT]
	alternative := spec.GetInputs()[IF_INPUT_ALTERNATIVE]

	// Parse condition to a bool
	i, err := typedvalues.Format(expr)
	if err != nil {
		return nil, err
	}
	condition, ok := i.(bool)
	if !ok {
		return nil, fmt.Errorf("condition needs to be a bool, but was '%v'", i)
	}

	// Output consequent or alternative based on condition
	if condition {
		return consequent, nil
	} else {
		return alternative, nil
	}
}
