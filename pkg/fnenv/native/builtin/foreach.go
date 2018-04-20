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

/*
FunctionForeach is a control flow construct to execute a certain task for each element in the provided input.
The tasks are executed in parallel.
Currently, foreach does not gather or store the outputs of the tasks in any way.

**Specification**

**input**       | required | types         | description
----------------|----------|---------------|--------------------------------------------------------
foreach/default | yes      | list          | The list of elements that foreach should be looped over.
do              | yes      | task/workflow | The action to perform for every element.
sequential      | no       | bool          | Whether to execute the tasks sequentially (default: false).

The element is made available to the action using the field `element`.

**output** None

**Example**

```
foo:
  run: foreach
  inputs:
    for:
    - a
    - b
    - c
    do:
      run: noop
      inputs: "{ task().element }"
```

A complete example of this function can be found in the [foreachwhale](../examples/whales/foreachwhale.wf.yaml) example.
*/
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
