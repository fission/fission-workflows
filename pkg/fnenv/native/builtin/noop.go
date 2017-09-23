package builtin

import (
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
)

const (
	NOOP_INPUT = types.INPUT_MAIN
)

type FunctionNoop struct{}

func (fn *FunctionNoop) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {

	var output *types.TypedValue
	switch len(spec.GetInputs()) {
	case 0:
		output = nil
	case 1:
		defaultInput, ok := spec.GetInputs()[NOOP_INPUT]
		if ok {
			output = defaultInput
			break
		}
		fallthrough
	default:
		results := map[string]interface{}{}
		for k, v := range spec.GetInputs() {
			i, err := typedvalues.Format(v)
			if err != nil {
				return nil, err
			}
			results[k] = i
		}
		p, err := typedvalues.Parse(results)
		if err != nil {
			return nil, err
		}
		output = p
	}
	logrus.WithFields(logrus.Fields{
		"spec":   spec,
		"output": output,
	}).Info("Internal Noop-function invoked.")
	return output, nil
}
