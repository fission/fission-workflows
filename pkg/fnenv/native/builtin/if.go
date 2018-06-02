package builtin

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
)

const (
	If               = "if"
	IfInputCondition = "if"
	IfInputThen      = "then"
	IfInputElse      = "else"
)

/*
FunctionIf is the simplest ways of altering the control flow of a workflow.
It allows you to implement an if-else construct; executing a branch or returning a specific output based on the
result of an execution.

**Specification**

**input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
if              | yes      | bool              | The condition to evaluate.
then            | no       | *                 | Value or action to return if the condition is true.
else            | no       | *                 | Value or action to return if the condition is false.

Unless the content type is specified explicitly, the workflow engine will infer the content-type based on the body.

**output** (*) Either the input of `then` or `else` (or none if not set).

**Example**

The following example shows the dynamic nature of this control flow.
If the if-expression evaluates to true, a static value is outputted.
Otherwise, a function is outputted (which in turn is executed).

```yaml
# ...
ifExample:
  run: if
  inputs:
    if: { param() > 42  }
    then: "foo"
    else:
      run: noop
# ...
```

A complete example of this function can be found in the [maybewhale](../examples/whales/maybewhale.wf.yaml) example.
*/
type FunctionIf struct{}

func (fn *FunctionIf) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {

	// Verify and get condition
	expr, err := ensureInput(spec.GetInputs(), IfInputCondition)
	if err != nil {
		return nil, err
	}

	// Get consequent alternative, if one of those does not exist, that is fine.
	consequent := spec.GetInputs()[IfInputThen]
	alternative := spec.GetInputs()[IfInputElse]

	// Parse condition to a bool
	i, err := typedvalues.Format(expr)
	if err != nil {
		return nil, err
	}
	condition, ok := i.(bool)
	if !ok {
		return nil, fmt.Errorf("condition '%v' needs to be a 'bool', but was '%v'", i, expr.Type)
	}

	// Output consequent or alternative based on condition
	logrus.Infof("If-task has evaluated to '%b''", condition)
	if condition {
		return consequent, nil
	}
	return alternative, nil
}
