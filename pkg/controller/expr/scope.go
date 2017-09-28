package expr

import (
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/sirupsen/logrus"
)

// Scope is a custom view of the data, which can be queried by the user.
type Scope struct {
	Workflow   *WorkflowScope
	Invocation *InvocationScope
	Tasks      map[string]*TaskScope
}

type WorkflowScope struct {
	*ObjectMetadata
	UpdatedAt     int64  // unix timestamp
	Status        string // workflow status
	ResolvedTasks map[string]*TaskDefScope
}

type TaskDefScope struct {
	Runtime  string
	Src      string
	Resolved string
}

type InvocationScope struct {
	*ObjectMetadata
	Inputs map[string]interface{}
}

type ObjectMetadata struct {
	Id        string
	CreatedAt int64 // unix timestamp
}

type TaskScope struct {
	*ObjectMetadata
	Status    string // TaskInvocation status
	UpdatedAt int64  // unix timestamp
	Inputs    map[string]interface{}
	Requires  map[string]*types.TaskDependencyParameters
	Name      string
	Output    interface{}
}

func NewScope(wf *types.Workflow, invoc *types.WorkflowInvocation) *Scope {

	tasks := map[string]*TaskScope{}
	for taskId, fn := range invoc.Status.Tasks {

		// Dep: pipe output of dynamic tasks
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
			ObjectMetadata: formatMetadata(fn.Metadata),
			Status:         fn.Status.Status.String(),
			UpdatedAt:      formatTimestamp(fn.Status.UpdatedAt),
			Inputs:         formatTypedValueMap(fn.Spec.Inputs),
			Requires:       taskDef.Requires,
			Name:           taskDef.FunctionRef,
			Output:         output,
		}
	}

	return &Scope{
		Workflow: &WorkflowScope{
			ObjectMetadata: formatMetadata(wf.Metadata),
			UpdatedAt:      formatTimestamp(wf.Status.UpdatedAt),
			Status:         wf.Status.Status.String(),
			ResolvedTasks:  formatResolvedTask(wf.Status.ResolvedTasks),
		},
		Invocation: &InvocationScope{
			ObjectMetadata: formatMetadata(invoc.Metadata),
			Inputs:         formatTypedValueMap(invoc.Spec.Inputs),
		},
		Tasks: tasks,
	}
}

func formatResolvedTask(resolved map[string]*types.TaskTypeDef) map[string]*TaskDefScope {
	results := map[string]*TaskDefScope{}
	for k, v := range resolved {
		results[k] = &TaskDefScope{
			Src:      v.Src,
			Runtime:  v.Runtime,
			Resolved: v.Resolved,
		}
	}
	return results
}

func formatTypedValueMap(values map[string]*types.TypedValue) map[string]interface{} {
	result := map[string]interface{}{}
	for k, v := range values {
		i, err := typedvalues.Format(v)
		if err != nil {
			logrus.Errorf("Failed to format: %s=%v", k, v)
			panic(err)
		}
		result[k] = i
	}
	return result
}

func formatMetadata(meta *types.ObjectMetadata) *ObjectMetadata {
	if meta == nil {
		return nil
	}
	return &ObjectMetadata{
		Id:        meta.Id,
		CreatedAt: formatTimestamp(meta.CreatedAt),
	}
}

func formatTimestamp(pts *timestamp.Timestamp) int64 {
	ts, _ := ptypes.Timestamp(pts)
	return ts.UnixNano()
}
