package typedvalues

import (
	"errors"

	"github.com/fission/fission-workflows/pkg/types"
)

// TODO move to more appropriate package
func ResolveTaskOutput(taskId string, invoc *types.WorkflowInvocation) *types.TypedValue {
	val, ok := invoc.Status.Tasks[taskId]
	if !ok {
		return nil
	}

	output := val.Status.Output
	if output == nil {
		return nil
	}

	switch ValueType(output.Type) {
	case TypeTask:
		for outputTaskId, outputTask := range invoc.Status.DynamicTasks {
			if dep, ok := outputTask.Spec.Requires[taskId]; ok && dep.Type == types.TaskDependencyParameters_DYNAMIC_OUTPUT {
				return ResolveTaskOutput(outputTaskId, invoc)
			}
		}
		return nil
	case TypeWorkflow:
		for outputTaskId, outputTask := range invoc.Status.DynamicTasks {
			if dep, ok := outputTask.Spec.Requires[taskId]; ok && dep.Type == types.TaskDependencyParameters_DYNAMIC_OUTPUT {
				return ResolveTaskOutput(outputTaskId, invoc)
			}
		}
		return nil
	}
	return output
}

func FormatMap(t *types.TypedValue) (map[string]interface{}, error) {
	i, err := Format(t)
	if err != nil {
		return nil, err
	}
	v, ok := i.(map[string]interface{})
	if !ok {
		return nil, errors.New("invalid type")
	}
	return v, nil
}

func FormatArray(t *types.TypedValue) ([]interface{}, error) {
	i, err := Format(t)
	if err != nil {
		return nil, err
	}
	v, ok := i.([]interface{})
	if !ok {
		return nil, errors.New("invalid type")
	}
	return v, nil
}

type Inputs map[string]*types.TypedValue

func Input(i interface{}) Inputs {
	in := Inputs{}
	in[types.INPUT_MAIN] = MustParse(i)
	return in
}

func verifyTypedValue(v *types.TypedValue, acceptableTypes ...ValueType) error {
	if v == nil {
		return TypedValueErr{
			src: v,
			err: ErrUnsupportedType,
		}
	}
	if !IsType(v, acceptableTypes...) {
		return TypedValueErr{
			src: v,
			err: ErrUnsupportedType,
		}
	}
	return nil
}

func IsType(v *types.TypedValue, ts ...ValueType) bool {
	if v == nil {
		return false
	}
	vt := ValueType(v.Type)
	for _, t := range ts {
		if t == vt {
			return true
		}
	}
	return false
}
