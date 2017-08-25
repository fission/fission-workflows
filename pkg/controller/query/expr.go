package query

import (
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
	"github.com/robertkrimen/otto"
)

type ExpressionParser interface {
	Resolve(scope interface{}, expr *types.TypedValue) (*types.TypedValue, error)
}

// TODO measure performance
type JavascriptExpressionParser struct {
	vm     *otto.Otto // TODO limit functionality (might need a fork?)
	parser typedvalues.Parser
}

func NewJavascriptExpressionParser(parser typedvalues.Parser) *JavascriptExpressionParser {
	// TODO inject helper functions
	return &JavascriptExpressionParser{
		vm:     otto.New(),
		parser: parser,
	}
}

func (oe *JavascriptExpressionParser) Resolve(scope interface{}, expr *types.TypedValue) (*types.TypedValue, error) {
	if !typedvalues.IsExpression(expr) {
		return expr, nil
	}

	oe.vm.Set("$", scope)
	jsResult, err := oe.vm.Run(expr.Value)
	if err != nil {
		return nil, err
	}

	i, _ := jsResult.Export() // Err is always nil
	return oe.parser.Parse(i)
}
