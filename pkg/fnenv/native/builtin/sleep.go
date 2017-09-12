package builtin

import (
	"time"

	"fmt"

	"strconv"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
)

const (
	SLEEP_INPUT_MS         = types.INPUT_MAIN
	SLEEP_INPUT_MS_DEFAULT = time.Duration(1) * time.Second
)

type FunctionSleep struct{}

func (f *FunctionSleep) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {
	duration := SLEEP_INPUT_MS_DEFAULT
	input, ok := spec.Inputs[SLEEP_INPUT_MS]
	if ok {
		i, err := typedvalues.Format(input)
		if err != nil {
			return nil, err
		}

		switch t := i.(type) {
		case string: // TODO support ints at some point
			ms, err := strconv.Atoi(t)
			if err != nil {
				return nil, err
			}
			duration = time.Duration(ms) * time.Millisecond
		default:
			return nil, fmt.Errorf("invalid input '%v'", input.Type)
		}
	}

	time.Sleep(duration)

	return nil, nil
}
