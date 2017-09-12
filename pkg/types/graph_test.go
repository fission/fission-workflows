package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateWorkflowWithDynamicTasks(t *testing.T) {
	dynamicTask := &Task{
		Id:          "injected",
		FunctionRef: "InjectedFunction",
		Requires: map[string]*TaskDependencyParameters{
			"foo": {
				Type: TaskDependencyParameters_DYNAMIC_OUTPUT,
			},
		},
	}

	workflow := &Workflow{
		Spec: &WorkflowSpec{
			Tasks: map[string]*Task{
				"foo": {
					FunctionRef: "123Function",
				},
				"bar": {
					FunctionRef: "123Function",
					Requires: map[string]*TaskDependencyParameters{
						"foo": {},
					},
				},
				"bar2": {
					FunctionRef: "123Function",
					Requires: map[string]*TaskDependencyParameters{
						"foo_injected": {},
					},
				},
			},
		},
	}
	invocation := &WorkflowInvocation{
		Status: &WorkflowInvocationStatus{
			DynamicTasks: map[string]*Task{
				"foo_injected": dynamicTask,
			},
		},
	}
	cwf := CalculateTaskDependencyGraph(workflow, invocation)

	assert.Equal(t, 4, len(cwf))
	assert.Equal(t, 2, len(cwf["bar"].Requires))
	assert.Equal(t, 1, len(cwf["foo_injected"].Requires))
	assert.Equal(t, 1, len(cwf["bar2"].Requires))
	assert.Equal(t, 0, len(cwf["foo"].Requires))
}
