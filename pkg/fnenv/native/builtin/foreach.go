package builtin

import (
	"errors"
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

const (
	Foreach                = "foreach"
	ForeachInputForeach    = "foreach"
	ForeachInputDo         = "do"
	ForeachInputCollect    = "collect"
	ForeachInputSequential = "sequential"
)

/*
FunctionForeach is a control flow construct to execute a certain task for each item in the provided input.
The tasks are executed in parallel.
Note, currently the task in the 'do' does not have access to state in the current workflow.

**Specification**

**input**                | required | types         | description
-------------------------|----------|---------------|--------------------------------------------------------
foreach                  | yes      | list          | The list of elements that foreach should be looped over.
do                       | yes      | task/workflow | The action to perform for every element.
sequential               | no       | bool          | Whether to execute the tasks sequentially (default: false).
collect                  | no       | bool          | Collect the outputs of the tasks into an array (default: true).

The element is made available to the action using the field `_item`.

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
      inputs: "{ task().Inputs._item }"
```

A complete example of this function can be found in the [foreachwhale](../examples/whales/foreachwhale.wf.yaml) example.
*/
type FunctionForeach struct{}

func (fn *FunctionForeach) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {
	// Verify and parse foreach
	headerTv, err := ensureInput(spec.GetInputs(), ForeachInputForeach)
	if err != nil {
		return nil, err
	}
	i, err := typedvalues.Format(headerTv)
	if err != nil {
		return nil, err
	}
	foreach, ok := i.([]interface{})
	if !ok {
		return nil, fmt.Errorf("condition '%v' needs to be a 'array', but was '%v'", i, headerTv.Type)
	}

	// Parse task
	taskTv, err := ensureInput(spec.GetInputs(), ForeachInputDo, typedvalues.TypeTask)
	if err != nil {
		return nil, err
	}
	flow, err := typedvalues.FormatControlFlow(taskTv)
	if err != nil {
		return nil, err
	}
	if flow.Workflow() != nil {
		return nil, errors.New("foreach does not support workflow inputs (yet)")
	}

	// Parse collect
	collect := true
	collectTv, ok := spec.Inputs[ForeachInputCollect]
	if ok {
		b, err := typedvalues.FormatBool(collectTv)
		if err != nil {
			return nil, fmt.Errorf("collect could not be parsed into a boolean: %v", err)
		}
		collect = b
	}

	// Parse sequential
	var seq bool
	seqTv, ok := spec.Inputs[ForeachInputSequential]
	if ok {
		b, err := typedvalues.FormatBool(seqTv)
		if err != nil {
			return nil, fmt.Errorf("sequential could not be parsed into a boolean: %v", err)
		}
		seq = b
	}

	// Create the workflows
	wf := &types.WorkflowSpec{
		OutputTask: "collector",
		Tasks:      types.Tasks{},
	}

	// Create the tasks for each element
	var tasks []string // Needed to preserve order of the input array
	for k, item := range foreach {
		f := flow.Clone()
		itemTv := typedvalues.MustParse(item)
		itemTv.SetLabel("priority", "1000") // Ensure that item is resolved before other parameters
		f.Input("_item", *itemTv)

		// TODO support workflows
		t := f.Task()
		name := fmt.Sprintf("do_%d", k)
		wf.AddTask(name, t)
		tasks = append(tasks, name)

		if seq && k != 0 {
			t.Require(tasks[k-1])
		}
	}

	// Add collector task
	ct := &types.TaskSpec{
		FunctionRef: "compose",
		Inputs:      types.Inputs{},
		Requires:    types.Require(tasks...),
	}
	var output []interface{}
	for _, k := range tasks {
		if collect {
			output = append(output, fmt.Sprintf("{output('%s')}", k))
		}
	}
	ct.Input(ComposeInput, typedvalues.MustParse(output))
	wf.AddTask("collector", ct)

	return typedvalues.Parse(wf)
}
