package typedvalues

import (
	"errors"

	"github.com/fission/fission-workflows/pkg/types"
)

func ResolveTaskOutput(taskId string, invoc *types.WorkflowInvocation) *types.TypedValue {
	val, ok := invoc.Status.Tasks[taskId]
	if !ok {
		return nil
	}

	output := val.Status.Output
	if output == nil {
		return nil
	}

	switch output.Type {
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

// MustParse transforms the value into a TypedValue or panics.
func UnsafeParse(i interface{}) *types.TypedValue {
	tv, err := Parse(i)
	if err != nil {
		panic(err)
	}
	return tv
}

// MustParse transforms the TypedValue into a value or panics.
func UnsafeFormat(tv *types.TypedValue) interface{} {
	i, err := Format(tv)
	if err != nil {
		panic(err)
	}
	return i
}

// TODO do not depend on json
func ParseString(s string) *types.TypedValue {
	return UnsafeParse("'" + s + "'")
}

func FormatString(t *types.TypedValue) (string, error) {
	i, err := Format(t)
	if err != nil {
		return "", err
	}
	v, ok := i.(string)
	if !ok {
		return "", errors.New("invalid type")
	}
	return v, nil
}

func FormatBool(t *types.TypedValue) (bool, error) {
	i, err := Format(t)
	if err != nil {
		return false, err
	}
	v, ok := i.(bool)
	if !ok {
		return false, errors.New("invalid type")
	}
	return v, nil
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

func FormatNumber(t *types.TypedValue) (float64, error) {
	i, err := Format(t)
	if err != nil {
		return 0, err
	}
	v, ok := i.(float64)
	if !ok {
		return 0, errors.New("invalid type")
	}
	return v, nil
}

type Inputs map[string]*types.TypedValue

func Input(i interface{}) Inputs {
	in := Inputs{}
	in[types.INPUT_MAIN] = UnsafeParse(i)
	return in
}
