package builtin

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

const (
	RepeatInputTimes = "times"
	RepeatInputDo    = "do"
)

type FunctionRepeat struct{}

func (fn *FunctionRepeat) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {

	n, ok := spec.GetInputs()[RepeatInputTimes]
	if !ok {
		return nil, fmt.Errorf("repeat needs '%s'", RepeatInputTimes)
	}

	do, ok := spec.GetInputs()[RepeatInputDo]
	if !ok {
		return nil, fmt.Errorf("repeat needs '%s'", RepeatInputDo)
	}

	// Parse condition to a int
	i, err := typedvalues.Format(n)
	if err != nil {
		return nil, err
	}
	times, ok := i.(int64)
	if !ok {
		return nil, fmt.Errorf("condition '%v' needs to be a 'bool', but was '%v'", i, n.Type)
	}

	if times > 0 {
		return do, nil // TODO inline workflow { do, FunctionRepeat(n - 1) }
	} else {
		return nil, nil
	}
}
