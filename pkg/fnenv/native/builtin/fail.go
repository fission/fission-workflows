package builtin

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
)

const (
	Fail         = "fail"
	FailInputMsg = types.INPUT_MAIN
)

var defaultErrMsg = typedvalues.UnsafeParse("fail function triggered")

type FunctionFail struct{}

func (fn *FunctionFail) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {
	var output *types.TypedValue
	switch len(spec.GetInputs()) {
	case 0:
		output = defaultErrMsg
	default:
		defaultInput, ok := spec.GetInputs()[FailInputMsg]
		if ok {
			output = defaultInput
			break
		}
	}
	logrus.WithFields(logrus.Fields{
		"spec":   spec,
		"output": output,
	}).Info("Internal Fail-function invoked.")

	msg, err := typedvalues.Format(output)
	if err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("%v", msg)
}
