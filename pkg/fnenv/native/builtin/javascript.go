package builtin

import (
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/robertkrimen/otto"
)

const (
	Javascript          = "javascript"
	JavascriptInputExpr = "expr"
	JavascriptInputArgs = "args"
	execTimeout         = time.Duration(100) * time.Millisecond
	errTimeout          = "javascript time out"
)

type FunctionJavascript struct {
	vm *otto.Otto
}

func NewFunctionJavascript() *FunctionJavascript {
	return &FunctionJavascript{
		vm: otto.New(),
	}
}

func (fn *FunctionJavascript) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {
	exprVal, err := verifyInput(spec.Inputs, JavascriptInputExpr, "string")
	argsVal, _ := spec.Inputs[JavascriptInputArgs]
	if err != nil {
		return nil, err
	}
	expr, err := typedvalues.FormatString(exprVal)
	if err != nil {
		return nil, err
	}
	args, err := typedvalues.Format(argsVal)
	if err != nil {
		return nil, err
	}

	result, err := fn.exec(expr, args)
	if err != nil {
		return nil, err
	}

	return typedvalues.Parse(result)
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
	return typedvalues.Parse(i)
}
