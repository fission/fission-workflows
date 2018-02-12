package builtin

import (
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

const (
	SleepInput        = types.INPUT_MAIN
	SleepInputDefault = time.Duration(1) * time.Second
)

type FunctionSleep struct{}

func (f *FunctionSleep) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {
	duration := SleepInputDefault
	input, ok := spec.Inputs[SleepInput]
	if ok {
		i, err := typedvalues.Format(input)
		if err != nil {
			return nil, err
		}

		switch t := i.(type) {
		case string:
			d, err := time.ParseDuration(t)
			if err != nil {
				return nil, err
			}
			duration = d
		case float64:
			duration = time.Duration(t) * time.Millisecond
		default:
			return nil, fmt.Errorf("invalid input '%v'", input.Type)
		}
	}

	time.Sleep(duration)

	return nil, nil
}
