package builtin

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
)

const (
	If               = "if"
	IfInputCondition = "if"
	IfInputThen      = "then"
	IfInputElse      = "else"
)

type FunctionIf struct{}

func (fn *FunctionIf) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {

	// Verify and get condition
	expr, err := ensureInput(spec.GetInputs(), IfInputCondition)
	if err != nil {
		return nil, err
	}

	// Get consequent alternative, if one of those does not exist, that is fine.
	consequent := spec.GetInputs()[IfInputThen]
	alternative := spec.GetInputs()[IfInputElse]

	// Parse condition to a bool
	i, err := typedvalues.Format(expr)
	if err != nil {
		return nil, err
	}
	condition, ok := i.(bool)
	if !ok {
		return nil, fmt.Errorf("condition '%v' needs to be a 'bool', but was '%v'", i, expr.Type)
	}

	// Output consequent or alternative based on condition
	logrus.Infof("If-task has evaluated to '%b''", condition)
	if condition {
		return consequent, nil
	} else {
		return alternative, nil
	}
}
