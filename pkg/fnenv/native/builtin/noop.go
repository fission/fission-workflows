package builtin

import (
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/sirupsen/logrus"
)

const (
	Noop      = "noop"
	NoopInput = types.INPUT_MAIN
)

type FunctionNoop struct{}

func (fn *FunctionNoop) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {

	var output *types.TypedValue
	switch len(spec.GetInputs()) {
	case 0:
		output = nil
	default:
		defaultInput, ok := spec.GetInputs()[NoopInput]
		if ok {
			output = defaultInput
			break
		}
	}
	logrus.WithFields(logrus.Fields{
		"spec":   spec,
		"output": output,
	}).Info("Internal Fail-function invoked.")
	return output, nil
}
