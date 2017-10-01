package builtin

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

const (
	REPEAT_INPUT_TIMES = "times"
	REPEAT_INPUT_DO    = "do"
)

type FunctionRepeat struct{}

func (fn *FunctionRepeat) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {

	n, ok := spec.GetInputs()[REPEAT_INPUT_TIMES]
	if !ok {
		return nil, fmt.Errorf("repeat needs '%s'", REPEAT_INPUT_TIMES)
	}

	do, ok := spec.GetInputs()[REPEAT_INPUT_DO]
	if !ok {
		return nil, fmt.Errorf("repeat needs '%s'", REPEAT_INPUT_DO)
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
