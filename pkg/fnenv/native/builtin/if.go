package builtin

import (
	"fmt"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
)

const (
	IF_INPUT_CONDITION   = "condition"
	IF_INPUT_ALTERNATIVE = "alternative"
	IF_INPUT_CONSEQUENT  = "consequent"
)

// Temporary file containing built-in internal functions
//
// Should be refactored to a extensible system, using go plugins for example.
type FunctionIf struct{}

func (fn *FunctionIf) Invoke(spec *types.FunctionInvocationSpec) (*types.TypedValue, error) {

	expr, err := verifyInput(spec.GetInputs(), IF_INPUT_CONDITION, typedvalues.FormatType(typedvalues.FORMAT_JSON, typedvalues.TYPE_BOOL)) // TODO Is already resolved?
	if err != nil {
		return nil, err
	}

	// TODO allow both flow or data
	consequent, err := verifyInput(spec.GetInputs(), IF_INPUT_CONSEQUENT, typedvalues.TYPE_FLOW)
	if err != nil {
		return nil, err
	}

	// TODO check type
	// TODO allow both flow or data
	alternative := spec.GetInputs()[IF_INPUT_ALTERNATIVE]

	i, err := typedvalues.Format(expr)
	if err != nil {
		return nil, err
	}

	condition, ok := i.(bool)
	if !ok {
		return nil, fmt.Errorf("Condition needs to be a bool, but was '%v'", i)
	}

	if condition {
		return consequent, nil
	} else {
		return alternative, nil
	}
}
