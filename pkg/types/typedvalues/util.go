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
			if dep, ok := outputTask.Requires[taskId]; ok && dep.Type == types.
				TaskDependencyParameters_DYNAMIC_OUTPUT {
				return ResolveTaskOutput(outputTaskId, invoc)
			}
		}
		return nil
	case TypeWorkflow:
		for outputTaskId, outputTask := range invoc.Status.DynamicTasks {
			if dep, ok := outputTask.Requires[taskId]; ok && dep.Type == types.TaskDependencyParameters_DYNAMIC_OUTPUT {
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
	s, ok := i.(string)
	if !ok {
		return "", errors.New("invalid type")
	}
	return s, nil
}
