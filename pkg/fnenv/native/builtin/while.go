package builtin

import (
	"errors"
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

const (
	While            = "while"
	WhileInputExpr   = "expr"
	WhileInputLimit  = "limit"
	WhileInputDelay  = "delay"
	WhileInputAction = "do"

	WhileDefaultDelay = time.Duration(0)
)

var (
	ErrLimitExceeded = errors.New("while limit exceeded")
)

/*
FunctionWhile consists of a control flow construct that will execute a specific task as long as the condition has
not been met.
The results of the executed action can be accessed using the task ID "action".

**Specification**

**input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
expr            | yes      | bool              | The condition which determines whether to continue or halt the loop.
do              | yes      | task/workflow     | The action to execute on each iteration.
limit           | no       | number            | The max number of iterations of the loop. (default: unlimited)

Notes:
- we currently cannot reevaluate the expr.
There needs to be support for looking up the source of an expression.
Maybe we can add the original expression to the labels.
- we might want to have a `prev` value here to reference the output of the previous iteration.


**output** (*) Either the value of the matching case, the default, or nothing (in case the default is not specified).

**Example**

```yaml
# ...
SwitchExample:
  run: while
  inputs:
    expr: "{ 42 > 0 }"
    limit: 10
    do:
      run: noop
# ...
```

A complete example of this function can be found in the [whilewhale](../examples/whales/whilewhale.wf.yaml) example.
*/
type FunctionWhile struct{}

func (fn *FunctionWhile) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {
	// Expr
	exprTv, err := ensureInput(spec.Inputs, WhileInputExpr)
	if err != nil {
		return nil, err
	}
	expr, err := typedvalues.FormatBool(exprTv)
	if err != nil {
		return nil, err
	}

	// Limit TODO support setting of the limit
	limitTv, err := ensureInput(spec.Inputs, WhileInputLimit)
	if err != nil {
		return nil, err
	}
	l, err := typedvalues.FormatNumber(limitTv)
	if err != nil {
		return nil, err
	}
	limit := int64(l)

	// Counter
	var count int64
	if countTv, ok := spec.Inputs["_count"]; ok {
		n, err := typedvalues.FormatNumber(countTv)
		if err != nil {
			return nil, err
		}
		count = int64(n)
	}

	// Delay
	delay := WhileDefaultDelay
	delayTv, ok := spec.Inputs[WhileInputDelay]
	if ok {
		s, err := typedvalues.FormatString(delayTv)
		if err != nil {
			return nil, err
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return nil, err
		}
		delay = d
	}

	// Action
	action, err := ensureInput(spec.Inputs, WhileInputAction, typedvalues.TypeWorkflow, typedvalues.TypeTask)
	if err != nil {
		return nil, err
	}

	// Logic
	if expr {
		// TODO support referencing of output in output value, to avoid needing to include 'prev' every time.
		if prev, ok := spec.Inputs["_prev"]; ok {
			return prev, nil
		}
		return nil, nil
	}

	if count > limit {
		return nil, ErrLimitExceeded
	}

	wf := &types.WorkflowSpec{
		OutputTask: "condition",
		Tasks: map[string]*types.TaskSpec{
			"wait": {
				FunctionRef: Sleep,
				Inputs: map[string]*types.TypedValue{
					SleepInput: typedvalues.MustParse(delay.String()),
				},
			},
			"action": {
				FunctionRef: Noop,
				Inputs:      typedvalues.Input(action),
				Requires:    types.Require("wait"),
			},
			"condition": {
				FunctionRef: While,
				Inputs: map[string]*types.TypedValue{
					WhileInputExpr:   exprTv,
					WhileInputDelay:  delayTv,
					WhileInputLimit:  limitTv,
					WhileInputAction: action,
					"_count":         typedvalues.MustParse(count + 1),
					"_prev":          typedvalues.MustParse("{output('action')}"),
				},
				Requires: types.Require("action"),
			},
		},
	}

	return typedvalues.Parse(wf)
}
