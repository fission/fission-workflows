package query

import (
	"time"

	"errors"

	"strings"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
	"github.com/sirupsen/logrus"
)

type ExpressionParser interface {
	Resolve(rootScope interface{}, currentScope interface{}, output interface{}, expr *types.TypedValue) (*types.TypedValue, error)
}

var (
	RESOLVING_TIMEOUT = time.Duration(100) * time.Millisecond

	ErrTimeOut = errors.New("expression resolver timed out")
)

type JavascriptExpressionParser struct {
	vm     *otto.Otto
	parser typedvalues.Parser
}

func NewJavascriptExpressionParser(parser typedvalues.Parser) *JavascriptExpressionParser {
	vm := otto.New()

	// Load builtin functions into Otto
	for varName, fn := range BuiltinFunctions {
		err := vm.Set(varName, func(call otto.FunctionCall) otto.Value {
			return fn.Apply(vm, call)
		})
		if err != nil {
			panic(err)
		}
	}
	return &JavascriptExpressionParser{
		vm:     vm,
		parser: parser,
	}
}

func (oe *JavascriptExpressionParser) Resolve(rootScope interface{}, currentScope interface{}, output interface{}, expr *types.TypedValue) (*types.TypedValue, error) {
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

			resolved, err := oe.Resolve(rootScope, currentScope, output, field)
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

	defer func() {
		if caught := recover(); caught != nil {
			if ErrTimeOut != caught {
				panic(caught)
			}
		}
	}()

	scoped := oe.vm.Copy()
	err := scoped.Set("$", rootScope)
	err = scoped.Set("task", currentScope)
	if output != nil {
		err = scoped.Set("output", output)
	}
	if err != nil {
		// Failed to set some variable
		return nil, err
	}

	go func() {
		<-time.After(RESOLVING_TIMEOUT)
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
