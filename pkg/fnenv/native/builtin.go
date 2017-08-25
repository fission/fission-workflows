package native

import (
	"errors"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/sirupsen/logrus"
)

// Temporary file containing built-in internal functions
//
// Should be refactored to a extensible system, using go plugins for example.
type FunctionIf struct{}

func (fn *FunctionIf) Invoke(spec *types.FunctionInvocationSpec) (*types.TypedValue, error) {
	return nil, errors.New("Not implemented FunctionIf")
}

type FunctionNoop struct{}

func (fn *FunctionNoop) Invoke(spec *types.FunctionInvocationSpec) (*types.TypedValue, error) {
	output := spec.GetInputs()[types.INPUT_MAIN]
	logrus.WithFields(logrus.Fields{
		"spec":   spec,
		"output": output,
	}).Info("Internal Noop-function invoked.")
	return output, nil
}
