package scheduler

import (
	"testing"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestCalculateWorkflowDynamicTaskInjection(t *testing.T) {
	dynamicTask, _ := typedvalues.Parse(&types.Task{
		Id:           "injected",
		Name:         "InjectedFunction",
		Dependencies: map[string]*types.TaskDependencyParameters{},
	})

	req := &ScheduleRequest{
		Workflow: &types.Workflow{
			Spec: &types.WorkflowSpec{
				Tasks: map[string]*types.Task{
					"foo": {
						Name: "123Function",
					},
					"bar": {
						Name: "123Function",
						Dependencies: map[string]*types.TaskDependencyParameters{
							"foo": {},
						},
					},
				},
			},
		},
		Invocation: &types.WorkflowInvocation{
			Status: &types.WorkflowInvocationStatus{
				Tasks: map[string]*types.FunctionInvocation{
					"foo": {
						Status: &types.FunctionInvocationStatus{
							Output: dynamicTask,
						},
					},
				},
			},
		},
	}
	cwf := calculateCurrentWorkflow(req)

	assert.Equal(t, len(cwf), 3)
	assert.Equal(t, len(cwf["bar"].Dependencies), 2)
	assert.Equal(t, len(cwf["foo_injected"].Dependencies), 1)
	assert.Equal(t, len(cwf["foo"].Dependencies), 0)
}
