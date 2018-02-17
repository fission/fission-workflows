package expr

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/robertkrimen/otto"
	"github.com/sirupsen/logrus"
)

// Built-in functions for the expression parser
var BuiltinFunctions = map[string]Function{
	"uid":     &UidFn{},
	"input":   &InputFn{},
	"output":  &OutputFn{},
	"param":   &ParamFn{},
	"task":    &TaskFn{},
	"current": &CurrentFn{},
}

// UidFn provides a function to generate a unique (string) id
type UidFn struct{}

// Apply generates a unique (string) id
func (qf *UidFn) Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value {
	uid, _ := vm.ToValue(util.Uid())
	return uid
}

// InputFn provides a function to get the input of a task for the given key. If no key is provided,
// the default key is used.
type InputFn struct{}

// Apply gets the input of a task for the given key. If no key is provided, the default key is used.
func (qf *InputFn) Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value {
	var task, inputKey string
	switch len(call.ArgumentList) {
	case 0:
		return otto.UndefinedValue()
	case 1:
		inputKey = types.INPUT_MAIN
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

// OutputFn provides a function to Get the output of a task.
type OutputFn struct{}

// Apply gets the output of a task. If no argument is provided the output of the current task is returned.
func (qf *OutputFn) Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value {
	var task string
	switch len(call.ArgumentList) {
	case 0:
		task = varCurrentTask
		fallthrough
	default:
		if len(task) == 0 {
			task = fmt.Sprintf("\"%s\"", call.Argument(0).String())
		}
		lookup := fmt.Sprintf("$.Tasks.%s.Output", task)
		result, err := vm.Eval(lookup)
		if err != nil {
			logrus.Warnf("Failed to lookup output: %s", lookup)
			return otto.UndefinedValue()
		}
		return result
	}
}

// ParmFn provides a function to get the invocation param for the given key
type ParamFn struct{}

// Apply gets the invocation param for the given key
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

// TaskFn provides a function to get a task for the given taskId.
type TaskFn struct{}

// Apply gets the task for the given taskId.
func (qf *TaskFn) Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value {
	switch len(call.ArgumentList) {
	case 0:
		return otto.UndefinedValue()
	default:
		param := call.Argument(0).String()
		lookup := fmt.Sprintf("$.Tasks.%s", param)
		result, err := vm.Eval(lookup)
		if err != nil {
			logrus.Warnf("Failed to lookup param: %s", lookup)
			return otto.UndefinedValue()
		}
		return result
	}
}

type CurrentFn struct{}

func (ct *CurrentFn) Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value {
	lookup := fmt.Sprintf("$.Tasks[%s]", varCurrentTask)
	result, err := vm.Eval(lookup)
	if err != nil {
		logrus.Warnf("Failed to lookup param: %s", lookup)
		return otto.UndefinedValue()
	}
	return result
}
