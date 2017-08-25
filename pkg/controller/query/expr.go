package query

import (
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
	"github.com/fission/fission-workflow/pkg/util"
	"github.com/robertkrimen/otto"
)

type ExpressionParser interface {
	Resolve(rootScope interface{}, scope interface{}, expr *types.TypedValue) (*types.TypedValue, error)
}

// TODO measure performance
type JavascriptExpressionParser struct {
	vm     *otto.Otto // TODO limit functionality (might need a fork?)
	parser typedvalues.Parser
}

/*
Helper functions
task().
dependency(id).

guid
 */
func NewJavascriptExpressionParser(parser typedvalues.Parser) *JavascriptExpressionParser {
	// TODO inject helper functions
	vm := otto.New()
	err := vm.Set("uid", func(call otto.FunctionCall) otto.Value {
		uid, _ := vm.ToValue(util.Uid())
		return uid
	})
	if err != nil {
		panic(err)
	}
	return &JavascriptExpressionParser{
		vm:     vm,
		parser: parser,
	}
}

func (oe *JavascriptExpressionParser) Resolve(rootScope interface{}, scope interface{}, expr *types.TypedValue) (*types.TypedValue, error) {
	if !typedvalues.IsExpression(expr) {
		return expr, nil
	}

	oe.vm.Set("$", rootScope)
	oe.vm.Set("@", scope)

	jsResult, err := oe.vm.Run(expr.Value)
	if err != nil {
		return nil, err
	}

	i, _ := jsResult.Export() // Err is always nil
	return oe.parser.Parse(i)
}
