package builtin

import (
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

const (
	Switch                 = "switch"
	SwitchInputCondition   = "switch"
	SwitchInputDefaultCase = "default"
)

type FunctionSwitch struct{}

func (fn *FunctionSwitch) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {
	switchVal, err := fn.getSwitch(spec.Inputs)
	if err != nil {
		return nil, err
	}

	if v, ok := spec.Inputs[switchVal]; ok {
		return v, nil
	}
	return spec.Inputs[SwitchInputDefaultCase], nil
}

func (fn *FunctionSwitch) getSwitch(inputs map[string]*types.TypedValue) (string, error) {
	tv, err := ensureInput(inputs, SwitchInputCondition)
	if err != nil {
		return "", err
	}
	return typedvalues.FormatString(tv)
}
