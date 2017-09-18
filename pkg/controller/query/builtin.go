package query

import (
	"fmt"

	"github.com/fission/fission-workflow/pkg/util"
	"github.com/robertkrimen/otto"
	"github.com/sirupsen/logrus"
)

//
// Built-in functions in the query parser
//

type QueryFn interface {
	Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value
}

var BuiltinFunctions = map[string]QueryFn{
	"uid":   &UidFn{},
	"input": &InputFn{},
	"output": &OutputFn{},
	"param": &ParamFn{},
}

type UidFn struct{}

func (qf *UidFn) Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value {
	uid, _ := vm.ToValue(util.Uid())
	return uid
}

type InputFn struct{}

func (qf *InputFn) Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value {
	var task, inputKey string
	switch len(call.ArgumentList) {
	case 0:
		return otto.UndefinedValue()
	case 1:
		inputKey = "default"
		fallthrough
	case 2:
		fallthrough
	default:
		task = call.Argument(0).String()
		if len(call.ArgumentList) > 1 {
			inputKey = call.Argument(1).String()
		}
		lookup := fmt.Sprintf("$.Tasks.%s.Inputs.%s", task, inputKey)
		result, err := vm.Eval(lookup)
		if err != nil {
			logrus.Warnf("Failed to lookup input: %s", lookup)
			return otto.UndefinedValue()
		}
		return result
	}
}

type OutputFn struct{}

func (qf *OutputFn) Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value {
	switch len(call.ArgumentList) {
	case 0:
		return otto.UndefinedValue()
	default:
		task := call.Argument(0).String()
		lookup := fmt.Sprintf("$.Tasks.%s.Output", task)
		result, err := vm.Eval(lookup)
		if err != nil {
			logrus.Warnf("Failed to lookup output: %s", lookup)
			return otto.UndefinedValue()
		}
		return result
	}
}

type ParamFn struct{}

func (qf *ParamFn) Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value {
	switch len(call.ArgumentList) {
	case 0:
		return otto.UndefinedValue()
	default:
		param := call.Argument(0).String()
		lookup := fmt.Sprintf("$.Invocation.Inputs.%s", param)
		result, err := vm.Eval(lookup)
		if err != nil {
			logrus.Warnf("Failed to lookup param: %s", lookup)
			return otto.UndefinedValue()
		}
		return result
	}
}



