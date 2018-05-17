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

// NamedInput provides the value along with the associated key.
type NamedInput struct {
	Key string
	Val *types.TypedValue
}

type namedInputSlice []NamedInput

func (n namedInputSlice) Len() int      { return len(n) }
func (n namedInputSlice) Swap(i, j int) { n[i], n[j] = n[j], n[i] }
func (n namedInputSlice) Less(i, j int) bool {
	return priority(n[i].Val) < priority(n[j].Val)
}

func sortNamedInputSlices(inputs []NamedInput) {
	sort.Sort(sort.Reverse(namedInputSlice(inputs)))
}

func priority(t *types.TypedValue) int {
	var p int
	if ps, ok := t.GetLabels()["priority"]; ok {
		i, err := strconv.Atoi(ps)
		if err != nil {
			logrus.Warnf("Ignoring invalid priority: %v", ps)
		} else {
			p = i
		}
	}
	return p
}

func toNamedInputs(inputs map[string]*types.TypedValue) []NamedInput {
	out := make([]NamedInput, len(inputs))
	var i int
	for k, v := range inputs {
		out[i] = NamedInput{
			Val: v,
			Key: k,
		}
		i++
	}
	return out
}

// Prioritize sorts the inputs based on the priority label (descending order)
func Prioritize(inputs map[string]*types.TypedValue) []NamedInput {
	namedInputs := toNamedInputs(inputs)
	sortNamedInputSlices(namedInputs)
	return namedInputs
}
