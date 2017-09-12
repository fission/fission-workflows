package builtin

import (
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/sirupsen/logrus"
)

const (
	NOOP_INPUT = types.INPUT_MAIN
)

type FunctionNoop struct{}

func (fn *FunctionNoop) Invoke(spec *types.FunctionInvocationSpec) (*types.TypedValue, error) {
	output := spec.GetInputs()[NOOP_INPUT]
	logrus.WithFields(logrus.Fields{
		"spec":   spec,
		"output": output,
	}).Info("Internal Noop-function invoked.")
	return output, nil
}
