package expr

import (
	"fmt"
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

func makeTestScope() *Scope {
	scope, _ := NewScope(nil, &types.WorkflowInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        "testWorkflowInvocation",
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec: &types.WorkflowInvocationSpec{
			Inputs: map[string]*typedvalues.TypedValue{
				types.InputMain: typedvalues.MustWrap("body"),
				"headers":       typedvalues.MustWrap("http-headers"),
			},
			Workflow: &types.Workflow{
				Metadata: &types.ObjectMetadata{
					Id:        "testWorkflow",
					CreatedAt: ptypes.TimestampNow(),
				},
				Status: &types.WorkflowStatus{
					Status:    types.WorkflowStatus_READY,
					UpdatedAt: ptypes.TimestampNow(),
					Tasks: map[string]*types.Task{
						"TaskA": {
							Metadata: types.NewObjectMetadata("TaskA"),
							Spec: &types.TaskSpec{
								Inputs: map[string]*typedvalues.TypedValue{
									types.InputMain: typedvalues.MustWrap("input-default"),
									"otherInput":    typedvalues.MustWrap("input-otherInput"),
								},
							},
							Status: &types.TaskStatus{
								FnRef: &types.FnRef{
									Runtime: "fission",
									ID:      "resolvedFissionFunction",
								},
							},
						},
					},
				},
				Spec: &types.WorkflowSpec{
					OutputTask: "TaskA",
				},
			},
		},
		Status: &types.WorkflowInvocationStatus{
			Status: types.WorkflowInvocationStatus_IN_PROGRESS,
			Tasks: map[string]*types.TaskInvocation{
				"TaskA": {
					Spec: &types.TaskInvocationSpec{},
					Status: &types.TaskInvocationStatus{
						Output: typedvalues.MustWrap("some output"),
						OutputHeaders: typedvalues.MustWrap(map[string]interface{}{
							"some-key": "some-value",
						}),
					},
				},
			},
		},
	})
	return scope
}

func TestOutputFn_Apply_OneArgument(t *testing.T) {
	parser := NewJavascriptExpressionParser()

	testScope := makeTestScope()
	result, err := parser.Resolve(testScope, "", mustParseExpr("{ output('TaskA') }"))
	assert.NoError(t, err)

	i := typedvalues.MustUnwrap(result)
	fmt.Println(testScope)

	assert.Equal(t, testScope.Tasks["TaskA"].Output, i)
}

func TestOutputFn_Apply_NoArgument(t *testing.T) {
	parser := NewJavascriptExpressionParser()

	testScope := makeTestScope()
	result, err := parser.Resolve(testScope, "TaskA", mustParseExpr("{ output() }"))
	assert.NoError(t, err)

	i := typedvalues.MustUnwrap(result)

	assert.Equal(t, testScope.Tasks["TaskA"].Output, i)
}

func TestInputFn_Apply_NoArgument(t *testing.T) {
	parser := NewJavascriptExpressionParser()

	testScope := makeTestScope()
	result, err := parser.Resolve(testScope, "TaskA", mustParseExpr("{ input() }"))
	assert.NoError(t, err)

	i := typedvalues.MustUnwrap(result)

	assert.Equal(t, "input-default", i)
}

func TestInputFn_Apply_OneArgument(t *testing.T) {
	parser := NewJavascriptExpressionParser()

	testScope := makeTestScope()
	result, err := parser.Resolve(testScope, "", mustParseExpr("{ input('TaskA') }"))
	assert.NoError(t, err)

	i := typedvalues.MustUnwrap(result)

	assert.Equal(t, "input-default", i)
}

func TestInputFn_Apply_TwoArguments(t *testing.T) {
	parser := NewJavascriptExpressionParser()

	testScope := makeTestScope()
	result, err := parser.Resolve(testScope, "", mustParseExpr("{ input('TaskA', 'otherInput') }"))
	assert.NoError(t, err)

	i := typedvalues.MustUnwrap(result)

	assert.Equal(t, "input-otherInput", i)
}

func TestParamFn_Apply_NoArgument(t *testing.T) {
	parser := NewJavascriptExpressionParser()
	testScope := makeTestScope()
	result, err := parser.Resolve(testScope, "", mustParseExpr("{ param() }"))
	assert.NoError(t, err)
	assert.Equal(t, "body", typedvalues.MustUnwrap(result))
}

func TestParamFn_Apply_OneArgument(t *testing.T) {
	parser := NewJavascriptExpressionParser()
	testScope := makeTestScope()
	result, err := parser.Resolve(testScope, "", mustParseExpr("{ param('headers') }"))
	assert.NoError(t, err)
	assert.Equal(t, "http-headers", typedvalues.MustUnwrap(result))
}

func TestUidFn_Apply(t *testing.T) {
	parser := NewJavascriptExpressionParser()
	testScope := makeTestScope()
	result, err := parser.Resolve(testScope, "", mustParseExpr("{ uid() }"))
	assert.NoError(t, err)
	assert.NotEmpty(t, typedvalues.MustUnwrap(result))
}

func TestTaskFn_Apply_OneArgument(t *testing.T) {
	parser := NewJavascriptExpressionParser()

	testScope := makeTestScope()
	result, err := parser.Resolve(testScope, "", mustParseExpr("{ task('TaskA') }"))
	assert.NoError(t, err)
	i := typedvalues.MustUnwrap(result)

	assert.Equal(t, util.MustConvertStructsToMap(testScope.Tasks["TaskA"]), i)
}

func TestTaskFn_Apply_NoArgument(t *testing.T) {
	parser := NewJavascriptExpressionParser()

	testScope := makeTestScope()
	result, err := parser.Resolve(testScope, "TaskA", mustParseExpr("{ task() }"))
	assert.NoError(t, err)

	i := typedvalues.MustUnwrap(result)

	assert.Equal(t, util.MustConvertStructsToMap(testScope.Tasks["TaskA"]), i)
}

func TestOutputHeadersFn_Apply_OneArgument(t *testing.T) {
	parser := NewJavascriptExpressionParser()

	testScope := makeTestScope()
	result, err := parser.Resolve(testScope, "", mustParseExpr("{ outputHeaders('TaskA') }"))
	assert.NoError(t, err)

	i := typedvalues.MustUnwrap(result)

	assert.Equal(t, testScope.Tasks["TaskA"].OutputHeaders, i)
}

func TestOutputHeadersFn_Apply_NoArgument(t *testing.T) {
	parser := NewJavascriptExpressionParser()

	testScope := makeTestScope()
	result, err := parser.Resolve(testScope, "TaskA", mustParseExpr("{ outputHeaders() }"))
	assert.NoError(t, err)

	i := typedvalues.MustUnwrap(result)

	assert.Equal(t, testScope.Tasks["TaskA"].OutputHeaders, i)
}
