package expr

import (
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/sirupsen/logrus"
)

// Scope is the broadest view of the workflow invocation, which can be queried by the user.
type Scope struct {
	Workflow   *WorkflowScope
	Invocation *InvocationScope
	Tasks      map[string]*TaskScope
}

// WorkflowScope provides information about the workflow definition.
type WorkflowScope struct {
	*ObjectMetadata
	UpdatedAt int64  // unix timestamp
	Status    string // workflow status
}

// InvocationScope object provides information about the current invocation.
type InvocationScope struct {
	*ObjectMetadata
	Inputs map[string]interface{}
}

// ObjectMetadata contains identity and meta-data about an object.
type ObjectMetadata struct {
	Id        string
	CreatedAt int64 // unix timestamp
}

// TaskScope holds information about a specific task execution within the current workflow invocation.
type TaskScope struct {
	*types.ObjectMetadata
	Status    string // TaskInvocation status
	UpdatedAt int64  // unix timestamp
	Inputs    map[string]interface{}
	Requires  map[string]*types.TaskDependencyParameters
	Output    interface{}
	Function  *types.FnRef
}

// NewScope creates a new scope given the workflow invocation and its associates workflow definition.
func NewScope(wf *types.Workflow, wfi *types.WorkflowInvocation) *Scope {

	tasks := map[string]*TaskScope{}
	for taskId, fn := range wfi.Status.Tasks {

		// Dep: pipe output of dynamic tasks
		t := typedvalues.ResolveTaskOutput(taskId, wfi)
		output, err := typedvalues.Format(t)
		if err != nil {
			panic(err)
		}

		task, _ := types.GetTask(wf, wfi, taskId)

		tasks[taskId] = &TaskScope{
			ObjectMetadata: fn.Metadata,
			Status:         fn.Status.Status.String(),
			UpdatedAt:      formatTimestamp(fn.Status.UpdatedAt),
			Inputs:         formatTypedValueMap(fn.Spec.Inputs),
			Requires:       task.Spec.Requires,
			Output:         output,
		}
	}

	return &Scope{
		Workflow: &WorkflowScope{
			ObjectMetadata: formatMetadata(wf.Metadata),
			UpdatedAt:      formatTimestamp(wf.Status.UpdatedAt),
			Status:         wf.Status.Status.String(),
		},
		Invocation: &InvocationScope{
			ObjectMetadata: formatMetadata(wfi.Metadata),
			Inputs:         formatTypedValueMap(wfi.Spec.Inputs),
		},
		Tasks: tasks,
	}
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
