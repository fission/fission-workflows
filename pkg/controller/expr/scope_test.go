package expr

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/ptypes"
	"github.com/stretchr/testify/assert"
)

func TestScopeExpr(t *testing.T) {
	expected := "hello world"
	expectedOutput, _ := typedvalues.Wrap(expected)

	actualScope, err := NewScope(nil, &types.WorkflowInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        "testWorkflowInvocation",
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec: &types.WorkflowInvocationSpec{
			WorkflowId: "testWorkflow",
			Workflow: &types.Workflow{
				Metadata: &types.ObjectMetadata{
					Id:        "testWorkflow",
					CreatedAt: ptypes.TimestampNow(),
				},
				Status: &types.WorkflowStatus{
					Status:    types.WorkflowStatus_PENDING,
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
			},
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
	assert.NoError(t, err)

	exprParser := NewJavascriptExpressionParser()

	resolved, err := exprParser.Resolve(actualScope, "fooTask", mustParseExpr("{$.Tasks.fooTask.Output}"))
	assert.NoError(t, err)

	resolvedString, _ := typedvalues.Unwrap(resolved)
	assert.Equal(t, expected, resolvedString)
}

func TestScopeOverride(t *testing.T) {
	expected := "hello world"
	expectedOutput, _ := typedvalues.Wrap(expected)

	scope1, err := NewScope(nil, &types.WorkflowInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        "testWorkflowInvocation",
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec: &types.WorkflowInvocationSpec{
			WorkflowId: "testWorkflow",
			Workflow: &types.Workflow{
				Metadata: &types.ObjectMetadata{
					Id:        "testWorkflow",
					CreatedAt: ptypes.TimestampNow(),
				},
				Status: &types.WorkflowStatus{
					Status:    types.WorkflowStatus_PENDING,
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
			},
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
	assert.NoError(t, err)

	// Test overriding with nil values
	scope2, err := NewScope(scope1, nil)
	assert.NoError(t, err)
	assert.Equal(t, scope1, scope2)
	assert.False(t, scope1 == scope2)

	// Test with overriding invocation
	scope4, err := NewScope(scope1, &types.WorkflowInvocation{
		Spec: &types.WorkflowInvocationSpec{
			Inputs: typedvalues.MustWrapMapTypedValue(map[string]interface{}{
				"foo": "bar",
			}),
			Workflow: &types.Workflow{
				Status: &types.WorkflowStatus{
					Status: types.WorkflowStatus_READY,
				},
			},
		},
	})
	assert.NoError(t, err)
	assert.NotEqual(t, scope2, scope4)
	assert.NotEqual(t, scope2.Workflow, scope4.Workflow)
}
