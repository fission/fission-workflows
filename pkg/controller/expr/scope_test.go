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
			Workflow: &types.Workflow{
				Metadata: &types.ObjectMetadata{
					Id:        "testWorkflow",
					CreatedAt: ptypes.TimestampNow(),
				},
				Status: &types.WorkflowStatus{
					Status:    types.WorkflowStatus_PENDING,
					UpdatedAt: ptypes.TimestampNow(),
					Tasks: map[string]*types.Task{
						"fooTask": {
							Status: &types.TaskStatus{
								FnRef: &types.FnRef{
									Runtime: "fission",
									ID:      "resolvedFissionFunction42",
								},
							},
						},
					},
				},
				Spec: &types.WorkflowSpec{
					ApiVersion: "1",
					OutputTask: "fooTask",
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
			Workflow: &types.Workflow{
				Metadata: &types.ObjectMetadata{
					Id:        "testWorkflow",
					CreatedAt: ptypes.TimestampNow(),
				},
				Status: &types.WorkflowStatus{
					Status:    types.WorkflowStatus_PENDING,
					UpdatedAt: ptypes.TimestampNow(),
					Tasks: map[string]*types.Task{
						"fooTask": {
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
					ApiVersion: "1",
					OutputTask: "fooTask",
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

	// Test with overriding workflow
	//scope3, err := NewScope(scope1, &types.Workflow{
	//	Status: &types.WorkflowStatus{
	//		Status: types.WorkflowStatus_READY,
	//	},
	//}, nil)
	//assert.NoError(t, err)
	//assert.NotEqual(t, scope2, scope3)
	//assert.Equal(t, scope2.Invocation, scope3.Invocation)
	//assert.Equal(t, scope2.Tasks, scope3.Tasks)

	// Test with overriding invocation
	scope4, err := NewScope(scope1, &types.WorkflowInvocation{
		Spec: &types.WorkflowInvocationSpec{
			Inputs: typedvalues.MustWrapMapTypedValue(map[string]interface{}{
				"foo": "bar",
			}),
		},
	})
	assert.NoError(t, err)
	assert.NotEqual(t, scope2, scope4)
	assert.Equal(t, scope2.Workflow, scope4.Workflow)
}
