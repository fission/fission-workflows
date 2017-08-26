package query

import (
	"testing"

	"fmt"
	"strings"

	"github.com/fission/fission-workflow/pkg/types/typedvalues"
)

var pf = typedvalues.NewDefaultParserFormatter()

var scope = map[string]interface{}{
	"bit": "bat",
}
var rootscope = map[string]interface{}{
	"foo":          "bar",
	"currentScope": scope,
}

func TestResolveTestRootScopePath(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(pf)

	resolved, err := exprParser.Resolve(rootscope, rootscope, typedvalues.Expr("$.currentScope.bit"))
	if err != nil {
		t.Error(err)
	}

	resolvedString, err := pf.Format(resolved)
	if err != nil {
		t.Error(err)
	}

	expected := scope["bit"]
	if resolvedString != expected {
		t.Errorf("Expected value '%v' does not match '%v'", expected, resolved)
	}
}

func TestResolveTestScopePath(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(pf)

	resolved, err := exprParser.Resolve(rootscope, scope, typedvalues.Expr("task.bit"))
	if err != nil {
		t.Error(err)
	}

	resolvedString, err := pf.Format(resolved)
	if err != nil {
		t.Error(err)
	}

	expected := scope["bit"]
	if resolvedString != expected {
		t.Errorf("Expected value '%v' does not match '%v'", expected, resolved)
	}
}

func TestResolveLiteral(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(pf)

	expected := "foobar"
	resolved, _ := exprParser.Resolve(rootscope, scope, typedvalues.Expr(fmt.Sprintf("'%s'", expected)))

	resolvedString, _ := pf.Format(resolved)
	if resolvedString != expected {
		t.Errorf("Expected value '%v' does not match '%v'", expected, resolved)
	}
}

func TestResolveTransformation(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(pf)

	src := "foobar"
	resolved, _ := exprParser.Resolve(rootscope, scope, typedvalues.Expr(fmt.Sprintf("'%s'.toUpperCase()", src)))
	expected := strings.ToUpper(src)

	resolvedString, _ := pf.Format(resolved)
	if resolvedString != expected {
		t.Errorf("Expected value '%v' does not match '%v'", src, resolved)
	}
}

func TestResolveInjectedFunction(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(pf)

	resolved, err := exprParser.Resolve(rootscope, scope, typedvalues.Expr("uid()"))

	if err != nil {
		t.Error(err)
	}

	resolvedString, _ := pf.Format(resolved)

	if resolvedString == "" {
		t.Error("Uid returned empty string")
	}
}
