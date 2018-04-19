package builtin

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

const (
	Foreach            = "foreach"
	ForeachInputHeader = "foreach"
	ForeachInputDo     = "do"
)

// FunctionForeach is a control flow construct to execute a certain task for each element in the provided input.
// The tasks are executed in parallel. Foreach does not gather or store the outputs of the tasks in any way.
//
// For example:
// foo:
//   run: foreach
//   inputs:
//     for: [],
//     do:
//       run: noop
//       inputs: "{$.???}"
// TODO add option to force serial execution
type FunctionForeach struct{}

func (fn *FunctionForeach) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {

	// Verify and get loop header
	headerTv, err := ensureInput(spec.GetInputs(), ForeachInputHeader)
	if err != nil {
		return nil, err
	}

	// Parse header to an array
	i, err := typedvalues.Format(headerTv)
	if err != nil {
		return nil, err
	}
	header, ok := i.([]interface{}) // TODO do we actually support headers?
	if !ok {
		return nil, fmt.Errorf("condition '%v' needs to be a 'array', but was '%v'", i, headerTv.Type)
	}

	// Task
	// TODO also support workflows
	taskTv, err := ensureInput(spec.GetInputs(), ForeachInputDo, typedvalues.TypeTask)
	if err != nil {
		return nil, err
	}
	task, err := typedvalues.FormatTask(taskTv)
	if err != nil {
		return nil, err
	}

	wf := &types.WorkflowSpec{
		OutputTask: "noopTask",
		Tasks: map[string]*types.TaskSpec{
			"noopTask": {
				FunctionRef: Noop,
			},
		},
	}

	for k, item := range header {
		t := &types.TaskSpec{
			FunctionRef: task.FunctionRef,
			Inputs:      map[string]*types.TypedValue{},
		}
		for inputKey, inputVal := range task.Inputs {
			t.Input(inputKey, inputVal)
		}
		t.Input("_item", typedvalues.MustParse(item))

		wf.AddTask(fmt.Sprintf("do_%d", k), t)
	}

	return typedvalues.Parse(wf)
}
