package query

import (
	"time"

	"errors"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
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
