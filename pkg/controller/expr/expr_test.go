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

	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	resolved, err := exprParser.Resolve(rootScope, "", typedvalues.Expr("{$.currentScope.bit}"))
	if err != nil {
		t.Error(err)
	}

	resolvedString, err := typedvalues.Format(resolved)
	if err != nil {
		t.Error(err)
	}

	expected := scope["bit"]
	assert.Equal(t, expected, resolvedString)
}

func TestResolveTestScopePath(t *testing.T) {
	currentTask := "fooTask"
	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	resolved, err := exprParser.Resolve(rootScope, currentTask, typedvalues.Expr(varCurrentTask))
	assert.NoError(t, err)

	resolvedString, err := typedvalues.Format(resolved)
	assert.NoError(t, err)

	assert.Equal(t, currentTask, resolvedString)
}

func TestResolveLiteral(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	expected := "foobar"
	resolved, err := exprParser.Resolve(rootScope, "output", typedvalues.Expr(fmt.Sprintf("'%s'", expected)))
	assert.NoError(t, err)

	resolvedString, _ := typedvalues.Format(resolved)
	assert.Equal(t, expected, resolvedString)
}

func TestResolveTransformation(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	src := "foobar"
	expected := strings.ToUpper(src)
	resolved, err := exprParser.Resolve(rootScope, "", typedvalues.Expr(fmt.Sprintf("'%s'.toUpperCase()", src)))
	assert.NoError(t, err)

	resolvedString, _ := typedvalues.Format(resolved)
	assert.Equal(t, expected, resolvedString)
}

func TestResolveInjectedFunction(t *testing.T) {

	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	resolved, err := exprParser.Resolve(rootScope, "", typedvalues.Expr("uid()"))
	assert.NoError(t, err)

	resolvedString, _ := typedvalues.Format(resolved)

	assert.NotEmpty(t, resolvedString)
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
			Tasks: map[string]*types.TaskStatus{
				"fooTask": {
					FnRef: &types.FnRef{
						Runtime:   "fission",
						RuntimeId: "resolvedFissionFunction",
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

	exprParser := NewJavascriptExpressionParser(typedvalues.DefaultParserFormatter)

	resolved, err := exprParser.Resolve(actualScope, "fooTask", typedvalues.Expr("{$.Tasks.fooTask.Output}"))
	assert.NoError(t, err)

	resolvedString, _ := typedvalues.Format(resolved)
	assert.Equal(t, expected, resolvedString)
}
