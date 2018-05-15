package typedvalues

import (
	"errors"
	"sort"
	"strconv"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/sirupsen/logrus"
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

func Input(i interface{}) types.Inputs {
	in := types.Inputs{}
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

type namedInput struct {
	Key string
	Val *types.TypedValue
}

// PrioritizeInputs sorts the inputs based on the priority label (descending order)
func PrioritizeInputs(inputs map[string]*types.TypedValue) []namedInput {
	var priorities []int
	priorityBuckets := map[int][]namedInput{}
	for k, v := range inputs {
		var p int
		if ps, ok := v.Labels["priority"]; ok {
			i, err := strconv.Atoi(ps)
			if err != nil {
				logrus.Warnf("Ignoring invalid priority: %v", ps)
			} else {
				p = i
			}
		}
		b, ok := priorityBuckets[p]
		if !ok {
			b = []namedInput{}
			priorities = append(priorities, p)
		}
		b = append(b, namedInput{
			Val: v,
			Key: k,
		})
		priorityBuckets[p] = b
	}

	// Sort priorities in descending order
	sort.Ints(priorities)
	var result []namedInput
	for i := len(priorities) - 1; i >= 0; i-- {
		bucket := priorityBuckets[i]
		for _, input := range bucket {
			result = append(result, input)
		}
	}

	return result
}
