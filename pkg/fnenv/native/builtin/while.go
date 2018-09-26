package builtin

import (
	"errors"
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/typedvalues/controlflow"
	"github.com/sirupsen/logrus"
)

const (
	While            = "while"
	WhileInputExpr   = "expr"
	WhileInputLimit  = "limit"
	WhileInputDelay  = "delay"
	WhileInputAction = "do"

	WhileDefaultDelay = time.Duration(100) * time.Millisecond
)

var (
	ErrLimitExceeded = errors.New("while limit exceeded")
)

/*
FunctionWhile consists of a control flow construct that will execute a specific task as long as the condition has
not been met.
The results of the executed action can be accessed using the task id "action".

**Specification**

**input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
expr            | yes      | bool              | The condition which determines whether to continue or halt the loop.
do              | yes      | task/workflow     | The action to execute on each iteration.
limit           | yes      | number            | The max number of iterations of the loop.

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

func (fn *FunctionWhile) Invoke(spec *types.TaskInvocationSpec) (*typedvalues.TypedValue, error) {
	// Expr
	exprTv, err := ensureInput(spec.Inputs, WhileInputExpr, typedvalues.TypeBool)
	if err != nil {
		return nil, err
	}
	expr, err := typedvalues.UnwrapBool(exprTv)
	if err != nil {
		return nil, fmt.Errorf("failed to format while condition to a boolean: %v", err)
	}
	exprSrc, ok := exprTv.GetMetadataValue("src")
	if !ok {
		return nil, fmt.Errorf("could not get source of '%v'", expr)
	}
	exprSrcTv, err := typedvalues.Wrap(exprSrc)
	if err != nil {
		return nil, err
	}

	// Limit
	limitTv, err := ensureInput(spec.Inputs, WhileInputLimit, typedvalues.TypeNumber...)
	if err != nil {
		return nil, err
	}
	limit, err := typedvalues.UnwrapInt64(limitTv)
	if err != nil {
		return nil, fmt.Errorf("failed to format limit to a number: %v", err)
	}
	// Counter
	var count int64
	if countTv, ok := spec.Inputs["_count"]; ok {
		count, err = typedvalues.UnwrapInt64(countTv)
		if err != nil {
			return nil, fmt.Errorf("failed to format _count to a number: %v", err)
		}
	}

	// Delay
	delay := WhileDefaultDelay
	delayTv, ok := spec.Inputs[WhileInputDelay]
	if ok {
		s, err := typedvalues.UnwrapString(delayTv)
		if err != nil {
			return nil, fmt.Errorf("failed to parse delay (%v) to string: %v", delayTv, err)
		}
		d, err := time.ParseDuration(s)
		if err != nil {
			return nil, fmt.Errorf("failed to parse duration: %v", err)
		}
		delay = d
	}
	delayTv = typedvalues.MustWrap(delay.String())

	// Action
	action, err := ensureInput(spec.Inputs, WhileInputAction)
	if err != nil {
		return nil, err
	}

	// Logic: escape while loop when expression is no longer true.
	if !expr {
		// TODO support referencing of output in output value, to avoid needing to include 'prev' every time.
		if prev, ok := spec.Inputs["_prev"]; ok {
			return prev, nil
		}
		return nil, nil
	}

	if count >= limit {
		return nil, ErrLimitExceeded
	}

	// Create the while-specific inputs
	prevTv := typedvalues.MustWrap("{output('action')}")
	prevTv.SetMetadata(typedvalues.MetadataPriority, "100")
	countTv := typedvalues.MustWrap(count + 1)
	countTv.SetMetadata(typedvalues.MetadataPriority, "100")
	logrus.Infof("[while] count: %v (limit %v)", count, limit)

	// If the action is a control flow construct add the while-specific inputs
	if controlflow.IsControlFlow(action) {
		cf, err := controlflow.UnwrapControlFlow(action)
		if err != nil {
			return nil, fmt.Errorf("failed to format workflow action: %v", err)
		}
		if count > 0 {
			cf.Input("_prev", *prevTv)
		}
		cf.Input("_count", *countTv)
		action, err = typedvalues.Wrap(cf.Proto())
		if err != nil {
			return nil, fmt.Errorf("failed to format task action: %v", err)
		}
	}

	wf := &types.WorkflowSpec{
		OutputTask: "condition",
		Tasks: map[string]*types.TaskSpec{
			"wait": {
				FunctionRef: Sleep,
				Inputs: map[string]*typedvalues.TypedValue{
					SleepInput: delayTv,
				},
			},
			"action": {
				FunctionRef: Noop,
				Inputs: map[string]*typedvalues.TypedValue{
					NoopInput: action,
				},
				Requires: types.Require("wait"),
			},
			"condition": {
				FunctionRef: While,
				Inputs: map[string]*typedvalues.TypedValue{
					WhileInputExpr:   exprSrcTv,
					WhileInputDelay:  delayTv,
					WhileInputLimit:  limitTv,
					WhileInputAction: action,
					"_count":         countTv,
					"_prev":          prevTv,
				},
				Requires: types.Require("action"),
			},
		},
	}
	wfTv, err := typedvalues.Wrap(wf)
	if err != nil {
		return nil, fmt.Errorf("failed to create while workflow: %v", err)
	}
	return wfTv, nil
}
