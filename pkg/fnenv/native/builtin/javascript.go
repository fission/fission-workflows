package builtin

import (
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/robertkrimen/otto"
	"github.com/sirupsen/logrus"
)

const (
	Javascript          = "javascript"
	JavascriptInputExpr = "expr"
	JavascriptInputArgs = "args"
	execTimeout         = time.Duration(100) * time.Millisecond
	errTimeout          = "javascript time out"
)

/*
FunctionJavascript allows you to create a task that evaluates an arbitrary JavaScript expression.
The implementation is similar to the inline evaluation of JavaScript in [expressions](./expressions.md) in inputs.
In that sense this implementations does not offer more functionality than inline expressions.
However, as it allows you to implement the entire task in JavaScript, this function is useful for prototyping and
stubbing particular functions.

**Specification**

**input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
expr            | yes      | string            | The JavaScript expression
args            | no       | *                 | The arguments that need to be present in the expression.

Note: the `expr` is of type `string` - not a `expression` - to prevent the workflow engine from evaluating the
expression prematurely.

**output** (*) The output of the expression.

**Example**

```yaml
# ...
JsExample:
  run: javascript
  inputs:
    expr: "a ^ b"
    args:
      a: 42
      b: 10
# ...
```

A complete example of this function can be found in the [fibonacci](../examples/misc/fibonacci.wf.yaml) example.
*/
type FunctionJavascript struct {
	vm *otto.Otto
}

func NewFunctionJavascript() *FunctionJavascript {
	return &FunctionJavascript{
		vm: otto.New(),
	}
}

func (fn *FunctionJavascript) Invoke(spec *types.TaskInvocationSpec) (*typedvalues.TypedValue, error) {
	exprVal, err := ensureInput(spec.Inputs, JavascriptInputExpr, typedvalues.TypeString)
	argsVal, _ := spec.Inputs[JavascriptInputArgs]
	if err != nil {
		return nil, err
	}
	expr, err := typedvalues.UnwrapString(exprVal)
	if err != nil {
		return nil, err
	}
	args, err := typedvalues.Unwrap(argsVal)
	if err != nil {
		return nil, err
	}
	logrus.WithField("taskID", spec.TaskId).
		Infof("[internal://%s] args: %v | expr: %v", Javascript, args, expr)
	result, err := fn.exec(expr, args)
	if err != nil {
		return nil, err
	}
	logrus.WithField("taskID", spec.TaskId).
		Infof("[internal://%s] %v => %v", Javascript, expr, result)

	return typedvalues.Wrap(result)
}

func (fn *FunctionJavascript) exec(expr string, args interface{}) (interface{}, error) {
	defer func() {
		if caught := recover(); caught != nil {
			if errTimeout != caught {
				panic(caught)
			}
		}
	}()

	scoped := fn.vm.Copy()
	switch t := args.(type) {
	case map[string]interface{}:
		for key, arg := range t {
			err := scoped.Set(key, arg)
			if err != nil {
				return nil, err
			}
		}
	default:
		err := scoped.Set("arg", t)
		if err != nil {
			return nil, err
		}
	}

	go func() {
		<-time.After(execTimeout)
		scoped.Interrupt <- func() {
			panic(errTimeout)
		}
	}()

	jsResult, err := scoped.Run(expr)
	if err != nil {
		return nil, err
	}
	i, _ := jsResult.Export() // Err is always nil
	return i, nil
}
