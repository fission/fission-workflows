package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateWorkflowWithDynamicTasks(t *testing.T) {
	dynamicTask := &Task{
		Id:   "injected",
		Name: "InjectedFunction",
		Dependencies: map[string]*TaskDependencyParameters{
			"foo": {
				Type: TaskDependencyParameters_FUNKTOR_OUTPUT,
			},
		},
	}

	workflow := &Workflow{
		Spec: &WorkflowSpec{
			Tasks: map[string]*Task{
				"foo": {
					Name: "123Function",
				},
				"bar": {
					Name: "123Function",
					Dependencies: map[string]*TaskDependencyParameters{
						"foo": {},
					},
				},
				"bar2": {
					Name: "123Function",
					Dependencies: map[string]*TaskDependencyParameters{
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
	assert.Equal(t, 2, len(cwf["bar"].Dependencies))
	assert.Equal(t, 1, len(cwf["foo_injected"].Dependencies))
	assert.Equal(t, 1, len(cwf["bar2"].Dependencies))
	assert.Equal(t, 0, len(cwf["foo"].Dependencies))
}
