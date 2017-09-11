package query

import (
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
)

// Scope is a custom view of the data, which can be queried by the user.
type Scope struct {
	Workflow   *WorkflowScope
	Invocation *InvocationScope
	Tasks      map[string]*TaskScope
}

type WorkflowScope struct {
	*types.ObjectMetadata
	*types.WorkflowStatus
}

type InvocationScope struct {
	*types.ObjectMetadata
	Inputs map[string]interface{}
}

type TaskScope struct {
	*types.ObjectMetadata
	*types.FunctionInvocationStatus
	Inputs       map[string]interface{}
	Dependencies map[string]*types.TaskDependencyParameters
	Name         string
	Output       interface{}
}

func NewScope(wf *types.Workflow, invoc *types.WorkflowInvocation) *Scope {

	tasks := map[string]*TaskScope{}
	for taskId, fn := range invoc.Status.Tasks {

		// Dep: pipe output of control flow structs
		t := typedvalues.ResolveTaskOutput(taskId, invoc)
		output, err := typedvalues.Format(t)
		if err != nil {
			panic(err)
		}

		taskDef, ok := invoc.Status.DynamicTasks[taskId]
		if !ok {
			taskDef = wf.Spec.Tasks[taskId]
		}

		tasks[taskId] = &TaskScope{
			ObjectMetadata:           fn.Metadata,
			FunctionInvocationStatus: fn.Status,
			Inputs:       formatTypedValueMap(fn.Spec.Inputs),
			Dependencies: taskDef.Dependencies,
			Name:         taskDef.Name,
			Output:       output,
		}
	}

	return &Scope{
		Workflow: &WorkflowScope{
			ObjectMetadata: wf.Metadata,
			WorkflowStatus: wf.Status,
		},
		Invocation: &InvocationScope{
			ObjectMetadata: invoc.Metadata,
			Inputs:         formatTypedValueMap(invoc.Spec.Inputs),
		},
		Tasks: tasks,
	}
}

func formatTypedValueMap(values map[string]*types.TypedValue) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range values {
		i, err := typedvalues.Format(v)
		if err != nil {
			panic(err)
		}
		result[k] = i
	}
	return result
}
