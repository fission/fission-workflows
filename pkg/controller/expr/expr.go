package expr

import (
	"errors"
	"strings"
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
	"github.com/sirupsen/logrus"
)

const (
	varScope         = "$"
	varCurrentTask   = "taskId"
	ResolvingTimeout = time.Duration(100) * time.Millisecond
)

var (
	ErrTimeOut = errors.New("expression resolver timed out")
)

// resolver resolves an expression within a given context/scope.
type Resolver interface {
	Resolve(rootScope interface{}, currentTask string, expr *types.TypedValue) (*types.TypedValue, error)
}

// Function is an interface for providing functions that are able to be injected into the Otto runtime.
type Function interface {
	Apply(vm *otto.Otto, call otto.FunctionCall) otto.Value
}

type JavascriptExpressionParser struct {
	vm     *otto.Otto
	parser typedvalues.Parser
}

func NewJavascriptExpressionParser(parser typedvalues.Parser) *JavascriptExpressionParser {
	vm := otto.New()

	// Load expression functions into Otto
	return &JavascriptExpressionParser{
		vm:     vm,
		parser: parser,
	}
}

func (oe *JavascriptExpressionParser) Resolve(rootScope interface{}, currentTask string,
	expr *types.TypedValue) (*types.TypedValue, error) {

	defer func() {
		if caught := recover(); caught != nil {
			if ErrTimeOut != caught {
				panic(caught)
			}
		}
	}()

	// TODO fix and add array
	// Handle composites
	if strings.HasSuffix(expr.Type, "object") {
		logrus.WithField("expr", expr).Info("Resolving object...")
		i, err := typedvalues.Format(expr)
		if err != nil {
			return nil, err
		}

		result := map[string]interface{}{}
		obj := i.(map[string]interface{})
		for k, v := range obj {
			field, err := typedvalues.Parse(v)
			if err != nil {
				return nil, err
			}

			resolved, err := oe.Resolve(rootScope, currentTask, field)
			if err != nil {
				return nil, err
			}

			actualVal, err := typedvalues.Format(resolved)
			if err != nil {
				return nil, err
			}
			result[k] = actualVal
		}
		return typedvalues.Parse(result)
	}

	if !typedvalues.IsExpression(expr) {
		return expr, nil
	}

	scoped := oe.vm.Copy()
	injectFunctions(scoped, BuiltinFunctions)
	err := scoped.Set(varScope, rootScope)
	err = scoped.Set(varCurrentTask, currentTask)
	if err != nil {
		// Failed to set some variable
		return nil, err
	}

	go func() {
		<-time.After(ResolvingTimeout)
		scoped.Interrupt <- func() {
			panic(ErrTimeOut)
		}
	}()

	jsResult, err := scoped.Run(expr.Value)
	if err != nil {
		return nil, err
	}

	i, _ := jsResult.Export() // Err is always nil
	return oe.parser.Parse(i)
}

func injectFunctions(vm *otto.Otto, fns map[string]Function) {
	for varName := range fns {
		func(fnName string) {
			err := vm.Set(fnName, func(call otto.FunctionCall) otto.Value {
				return fns[fnName].Apply(vm, call)
			})
			if err != nil {
				panic(err)
			}
		}(varName)
	}
}
