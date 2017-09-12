package query

import (
	"testing"

	"fmt"
	"strings"

	"github.com/fission/fission-workflow/pkg/types/typedvalues"
)

var scope = map[string]interface{}{
	"bit": "bat",
}
var rootscope = map[string]interface{}{
	"foo":          "bar",
	"currentScope": scope,
}

func TestResolveTestRootScopePath(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	resolved, err := exprParser.Resolve(rootscope, rootscope, typedvalues.Expr("$.currentScope.bit"))
	if err != nil {
		t.Error(err)
	}

	resolvedString, err := typedvalues.Format(resolved)
	if err != nil {
		t.Error(err)
	}

	expected := scope["bit"]
	if resolvedString != expected {
		t.Errorf("Expected value '%v' does not match '%v'", expected, resolved)
	}
}

func TestResolveTestScopePath(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	resolved, err := exprParser.Resolve(rootscope, scope, typedvalues.Expr("task.bit"))
	if err != nil {
		t.Error(err)
	}

	resolvedString, err := typedvalues.Format(resolved)
	if err != nil {
		t.Error(err)
	}

	expected := scope["bit"]
	if resolvedString != expected {
		t.Errorf("Expected value '%v' does not match '%v'", expected, resolved)
	}
}

func TestResolveLiteral(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	expected := "foobar"
	resolved, _ := exprParser.Resolve(rootscope, scope, typedvalues.Expr(fmt.Sprintf("'%s'", expected)))

	resolvedString, _ := typedvalues.Format(resolved)
	if resolvedString != expected {
		t.Errorf("Expected value '%v' does not match '%v'", expected, resolved)
	}
}

func TestResolveTransformation(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	src := "foobar"
	resolved, _ := exprParser.Resolve(rootscope, scope, typedvalues.Expr(fmt.Sprintf("'%s'.toUpperCase()", src)))
	expected := strings.ToUpper(src)

	resolvedString, _ := typedvalues.Format(resolved)
	if resolvedString != expected {
		t.Errorf("Expected value '%v' does not match '%v'", src, resolved)
	}
}

func TestResolveInjectedFunction(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	resolved, err := exprParser.Resolve(rootscope, scope, typedvalues.Expr("uid()"))

	if err != nil {
		t.Error(err)
	}

	resolvedString, _ := typedvalues.Format(resolved)

	if resolvedString == "" {
		t.Error("Uid returned empty string")
	}
}
