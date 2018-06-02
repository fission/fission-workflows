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
	"uid":    &UidFn{},
	"input":  &InputFn{},
	"output": &OutputFn{},
	"param":  &ParamFn{},
	"task":   &TaskFn{},
}

// UidFn provides a function to generate a unique (string) id
type UidFn struct{}

// Apply generates a unique (string) id
func (qf *UidFn) Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value {
	uid, _ := vm.ToValue(util.UID())
	return uid
}

// InputFn provides a function to get the input of a task for the given key. If no key is provided,
// the default key is used.
type InputFn struct{}

// Apply gets the input of a task for the given key. If no key is provided, the default key is used.
// If no argument is provided at all, the default key of the current task will be used.
func (qf *InputFn) Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value {
	var task, inputKey string
	switch len(call.ArgumentList) {
	case 0:
		task = varCurrentTask
		fallthrough
	case 1:
		inputKey = types.InputMain
		fallthrough
	case 2:
		fallthrough
	default:
		// Set task if argument provided
		if len(call.ArgumentList) > 0 {
			task = fmt.Sprintf("\"%s\"", call.Argument(0).String())
		}
		// Set input key if argument provided
		if len(call.ArgumentList) > 1 {
			inputKey = call.Argument(1).String()
		}
		lookup := fmt.Sprintf("$.Tasks[%s].Inputs[\"%s\"]", task, inputKey)
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
		// Set task if argument provided
		if len(call.ArgumentList) > 0 {
			task = fmt.Sprintf("\"%s\"", call.Argument(0).String())
		}
		lookup := fmt.Sprintf("$.Tasks[%s].Output", task)
		result, err := vm.Eval(lookup)
		if err != nil {
			logrus.Warnf("Failed to lookup output: %s", lookup)
			return otto.UndefinedValue()
		}
		return result
	}
}

// ParmFn provides a function to get the invocation param for the given key. If no key is provided, the default key
// is used.
type ParamFn struct{}

// Apply gets the invocation param for the given key. If no key is provided, the default key is used.
func (qf *ParamFn) Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value {
	var key string
	switch len(call.ArgumentList) {
	case 0:
		key = types.InputMain
		fallthrough
	default:
		// Set key if argument provided
		if len(call.ArgumentList) > 0 {
			key = call.Argument(0).String()
		}
		lookup := fmt.Sprintf("$.Invocation.Inputs[\"%s\"]", key)
		result, err := vm.Eval(lookup)
		if err != nil {
			logrus.Warnf("Failed to lookup param: %s", lookup)
			return otto.UndefinedValue()
		}
		return result
	}
}

// TaskFn provides a function to get a task for the given taskId. If no argument is provided the current task is
// returned.
type TaskFn struct{}

// Apply gets the task for the given taskId. If no argument is provided the current task is returned.
func (qf *TaskFn) Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value {
	var task string
	switch len(call.ArgumentList) {
	case 0:
		task = varCurrentTask
		fallthrough
	default:
		// Set task if argument provided
		if len(call.ArgumentList) > 0 {
			task = fmt.Sprintf("\"%s\"", call.Argument(0).String())
		}
		lookup := fmt.Sprintf("$.Tasks[%s]", task)
		result, err := vm.Eval(lookup)
		if err != nil {
			logrus.Warnf("Failed to lookup param: %s", lookup)
			return otto.UndefinedValue()
		}
		return result
	}
}

func manualEval(vm *otto.Otto, s string) interface{} {
	result, err := vm.Eval(s)
	if err != nil {
		panic(err)
	}
	i, err := result.Export()
	if err != nil {
		panic(err)
	}
	return i
}
