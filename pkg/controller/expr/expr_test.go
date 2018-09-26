package expr

import (
	"fmt"
	"strings"
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

var scope = map[string]interface{}{
	"bit": "bat",
}
var rootScope = map[string]interface{}{
	"foo":          "bar",
	"currentScope": scope,
}

func TestResolveTestRootScopePath(t *testing.T) {

	exprParser := NewJavascriptExpressionParser()

	resolved, err := exprParser.Resolve(rootScope, "", mustParseExpr("{$.currentScope.bit}"))
	if err != nil {
		t.Error(err)
	}

	resolvedString, err := typedvalues.Unwrap(resolved)
	if err != nil {
		t.Error(err)
	}

	expected := scope["bit"]
	assert.Equal(t, expected, resolvedString)
}

func TestResolveTestScopePath(t *testing.T) {
	currentTask := "fooTask"
	exprParser := NewJavascriptExpressionParser()

	resolved, err := exprParser.Resolve(rootScope, currentTask, mustParseExpr("{"+varCurrentTask+"}"))
	assert.NoError(t, err)

	resolvedString, err := typedvalues.Unwrap(resolved)
	assert.NoError(t, err)

	assert.Equal(t, currentTask, resolvedString)
}

func TestResolveLiteral(t *testing.T) {

	exprParser := NewJavascriptExpressionParser()

	expected := "foobar"
	resolved, err := exprParser.Resolve(rootScope, "output", mustParseExpr(fmt.Sprintf("{'%s'}", expected)))
	assert.NoError(t, err)

	resolvedString, _ := typedvalues.Unwrap(resolved)
	assert.Equal(t, expected, resolvedString)
}

func TestResolveTransformation(t *testing.T) {

	exprParser := NewJavascriptExpressionParser()

	src := "foobar"
	expected := strings.ToUpper(src)
	resolved, err := exprParser.Resolve(rootScope, "", mustParseExpr(fmt.Sprintf("{'%s'.toUpperCase()}", src)))
	assert.NoError(t, err)

	resolvedString, _ := typedvalues.Unwrap(resolved)
	assert.Equal(t, expected, resolvedString)
}

func TestResolveInjectedFunction(t *testing.T) {

	exprParser := NewJavascriptExpressionParser()

	resolved, err := exprParser.Resolve(rootScope, "", mustParseExpr("{uid()}"))
	assert.NoError(t, err)

	resolvedString, _ := typedvalues.Unwrap(resolved)

	assert.NotEmpty(t, resolvedString)
}

func TestScope(t *testing.T) {
	expected := "hello world"
	expectedOutput, _ := typedvalues.Wrap(expected)

	actualScope, _ := NewScope(&types.Workflow{
		Metadata: &types.ObjectMetadata{
			Id:        "testWorkflow",
			CreatedAt: ptypes.TimestampNow(),
		},
		Status: &types.WorkflowStatus{
			Status:    types.WorkflowStatus_READY,
			UpdatedAt: ptypes.TimestampNow(),
			Tasks: map[string]*types.TaskStatus{
				"fooTask": {
					FnRef: &types.FnRef{
						Runtime: "fission",
						ID:      "resolvedFissionFunction",
					},
				},
			},
		},
		Spec: &types.WorkflowSpec{
			ApiVersion: "1",
			OutputTask: "fooTask",
			Tasks: map[string]*types.TaskSpec{
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

	exprParser := NewJavascriptExpressionParser()

	resolved, err := exprParser.Resolve(actualScope, "fooTask", mustParseExpr("{$.Tasks.fooTask.Output}"))
	assert.NoError(t, err)

	resolvedString, _ := typedvalues.Unwrap(resolved)
	assert.Equal(t, expected, resolvedString)
}

func mustParseExpr(s string) *typedvalues.TypedValue {
	tv := typedvalues.MustWrap(s)
	if tv.ValueType() != typedvalues.TypeExpression {
		panic(fmt.Sprintf("Should be %v, but was '%v'", typedvalues.TypeExpression, tv.ValueType()))
	}

	return tv
}
