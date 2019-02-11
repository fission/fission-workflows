package types

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalculateWorkflowWithDynamicTasks(t *testing.T) {
	workflow := NewWorkflow("wf-1")
	workflow.Spec.Tasks = map[string]*TaskSpec{
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
	}
	invocation := NewWorkflowInvocation("wf-1", "wfi-45")
	invocation.Status.DynamicTasks = map[string]*Task{
		"some_dynamic_task": {
			Metadata: NewObjectMetadata("some_dynamic_task"),
			Spec: &TaskSpec{
				FunctionRef: "InjectedFunction",
				Requires: map[string]*TaskDependencyParameters{
					"foo": {},
				},
			},
		},
		"bar": {
			Metadata: NewObjectMetadata("bar"),
			Spec: &TaskSpec{
				Requires: map[string]*TaskDependencyParameters{
					"foo":               {},
					"some_dynamic_task": {},
				},
			},
		},
		"bar2": {
			Metadata: NewObjectMetadata("bar2"),
			Spec: &TaskSpec{
				Requires: map[string]*TaskDependencyParameters{
					"some_dynamic_task": {},
				},
				Await: 42,
			},
		},
	}
	invocation.Spec.Workflow = workflow
	cwf := invocation.Tasks()
	for k := range cwf {
		fmt.Println(k)
	}

	assert.Equal(t, 4, len(cwf))
	assert.Equal(t, 2, len(cwf["bar"].Spec.Requires))
	assert.Equal(t, 1, len(cwf["some_dynamic_task"].Spec.Requires))
	assert.Equal(t, 1, len(cwf["bar2"].Spec.Requires))
	assert.Equal(t, 0, len(cwf["foo"].Spec.Requires))
	assert.Equal(t, int32(42), cwf["bar2"].Spec.Await)
}
