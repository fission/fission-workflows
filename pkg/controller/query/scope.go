package query

import (
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
)

// The scope is a custom view of the data that can be queried by the user
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

		out, err := typedvalues.NewDefaultParserFormatter().Format(fn.Status.Output)
		if err != nil {
			panic(err)
		}
		tasks[taskId] = &TaskScope{
			ObjectMetadata:           fn.Metadata,
			FunctionInvocationStatus: fn.Status,
			Inputs:                   formatTypedValueMap(fn.Spec.Inputs),
			Dependencies:             wf.Spec.Tasks[taskId].Dependencies,
			Name:                     wf.Spec.Tasks[taskId].Name,
			Output:                   out,
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
		i, err := typedvalues.NewDefaultParserFormatter().Format(v)
		if err != nil {
			panic(err)
		}
		result[k] = i
	}
	return result
}
