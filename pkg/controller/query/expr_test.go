package query

import (
	"testing"

	"fmt"
	"strings"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/ptypes"
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

	resolved, err := exprParser.Resolve(rootscope, rootscope, nil, typedvalues.Expr("{$.currentScope.bit}"))
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

	resolved, err := exprParser.Resolve(rootscope, scope, nil, typedvalues.Expr("task.bit"))
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
	resolved, _ := exprParser.Resolve(rootscope, scope, "output", typedvalues.Expr(fmt.Sprintf("'%s'", expected)))

	resolvedString, _ := typedvalues.Format(resolved)
	if resolvedString != expected {
		t.Errorf("Expected value '%v' does not match '%v'", expected, resolved)
	}
}

func TestResolveOutput(t *testing.T) {
	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	expected := "expected"
	resolved, _ := exprParser.Resolve(rootscope, scope, map[string]string{
		"acme": expected,
	}, typedvalues.Expr("output.acme"))

	resolvedString, _ := typedvalues.Format(resolved)
	if resolvedString != expected {
		t.Errorf("Expected value '%v' does not match '%v'", expected, resolved)
	}
}

func TestResolveTransformation(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	src := "foobar"
	resolved, _ := exprParser.Resolve(rootscope, scope, nil, typedvalues.Expr(fmt.Sprintf("'%s'.toUpperCase()", src)))
	expected := strings.ToUpper(src)

	resolvedString, _ := typedvalues.Format(resolved)
	if resolvedString != expected {
		t.Errorf("Expected value '%v' does not match '%v'", src, resolved)
	}
}

func TestResolveInjectedFunction(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	resolved, err := exprParser.Resolve(rootscope, scope, nil, typedvalues.Expr("uid()"))

	if err != nil {
		t.Error(err)
	}

	resolvedString, _ := typedvalues.Format(resolved)

	if resolvedString == "" {
		t.Error("Uid returned empty string")
	}
}

func TestScope(t *testing.T) {
	expected := "hello world"
	expectedOutput, _ := typedvalues.Parse(expected)

	actualScope := NewScope(&types.Workflow{
		Metadata: &types.ObjectMetadata{
			Id:        "testWorkflow",
			CreatedAt: ptypes.TimestampNow(),
		},
		Status: &types.WorkflowStatus{
			Status:    types.WorkflowStatus_READY,
			UpdatedAt: ptypes.TimestampNow(),
			ResolvedTasks: map[string]*types.TaskTypeDef{
				"fooTask": {
					Src:      "fissionFunction",
					Runtime:  "fission",
					Resolved: "resolvedFissionFunction",
				},
			},
		},
		Spec: &types.WorkflowSpec{
			ApiVersion: "1",
			OutputTask: "fooTask",
			Tasks: map[string]*types.Task{
				"fooTask": {
					FunctionRef: "fissionFunction",
				},
			},
		},
	}, &types.WorkflowInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        "testWorkflowInvocation",
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec: &types.WorkflowInvocationSpec{
			WorkflowId: "testWorkflow",
		},
		Status: &types.WorkflowInvocationStatus{
			Status: types.WorkflowInvocationStatus_IN_PROGRESS,
			Tasks: map[string]*types.TaskInvocation{
				"fooTask": {
					Spec: &types.TaskInvocationSpec{},
					Status: &types.TaskInvocationStatus{
						Output: expectedOutput,
					},
				},
			},
		},
	})

	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	resolved, _ := exprParser.Resolve(actualScope, actualScope.Tasks["fooTask"], nil,
		typedvalues.Expr("{$.Tasks.fooTask.Output}"))

	resolvedString, _ := typedvalues.Format(resolved)
	if resolvedString != expected {
		t.Errorf("Expected value '%v' does not match '%v'", resolvedString, expectedOutput)
	}
}
