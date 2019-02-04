package expr

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/typedvalues/controlflow"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/imdario/mergo"
	"github.com/pkg/errors"
)

var (
	ErrMergeTypeMismatch = errors.New("cannot merge incompatible types")
)

type DeepCopier interface {
	DeepCopy() DeepCopier
}

// Scope is the broadest view of the workflow invocation, which can be queried by the user.
type Scope struct {
	Workflow   *WorkflowScope
	Invocation *InvocationScope
	Tasks      Tasks
}

func (s *Scope) DeepCopy() DeepCopier {
	if s == nil {
		return nil
	}
	return &Scope{
		Workflow:   s.Workflow.DeepCopy().(*WorkflowScope),
		Invocation: s.Invocation.DeepCopy().(*InvocationScope),
		Tasks:      s.Tasks.DeepCopy().(Tasks),
	}
}

type Tasks map[string]*TaskScope

// WorkflowScope provides information about the workflow definition.
type WorkflowScope struct {
	*ObjectMetadata
	UpdatedAt int64  // unix timestamp
	Status    string // workflow status
	Name      string
	Internal  bool
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
	*ObjectMetadata
	Status        string // TaskInvocation status
	UpdatedAt     int64  // unix timestamp
	Inputs        map[string]interface{}
	Requires      map[string]*types.TaskDependencyParameters
	Output        interface{}
	OutputHeaders interface{}
	Function      string
}

func (s Tasks) DeepCopy() DeepCopier {
	if s == nil {
		return nil
	}
	copied := make(Tasks, len(s))
	for k, v := range s {
		copied[k] = v.DeepCopy().(*TaskScope)
	}
	return copied
}

func (s *WorkflowScope) DeepCopy() DeepCopier {
	if s == nil {
		return nil
	}
	return &WorkflowScope{
		ObjectMetadata: s.ObjectMetadata.DeepCopy().(*ObjectMetadata),
		UpdatedAt:      s.UpdatedAt,
		Status:         s.Status,
		Name:           s.Name,
		Internal:       s.Internal,
	}
}

func (s *InvocationScope) DeepCopy() DeepCopier {
	if s == nil {
		return nil
	}
	return &InvocationScope{
		ObjectMetadata: s.ObjectMetadata.DeepCopy().(*ObjectMetadata),
		Inputs:         DeepCopy(s.Inputs).(map[string]interface{}),
	}
}

func (s *ObjectMetadata) DeepCopy() DeepCopier {
	if s == nil {
		return nil
	}
	return &ObjectMetadata{
		Id:        s.Id,
		CreatedAt: s.CreatedAt,
	}
}

func (s *TaskScope) DeepCopy() DeepCopier {
	if s == nil {
		return nil
	}
	var requires map[string]*types.TaskDependencyParameters
	if s.Requires != nil {
		requires = make(map[string]*types.TaskDependencyParameters, len(s.Requires))
		for k, v := range s.Requires {
			requires[k] = DeepCopy(v).(*types.TaskDependencyParameters)
		}
	}

	return &TaskScope{
		ObjectMetadata: s.ObjectMetadata.DeepCopy().(*ObjectMetadata),
		Status:         s.Status,
		UpdatedAt:      s.UpdatedAt,
		Inputs:         DeepCopy(s.Inputs).(map[string]interface{}),
		Requires:       requires,
		Output:         DeepCopy(s.Output),
		OutputHeaders:  DeepCopy(s.OutputHeaders),
		Function:       s.Function,
	}
}

// NewScope creates a new scope given the workflow invocation and its associates workflow definition.
func NewScope(base *Scope, wfi *types.WorkflowInvocation) (*Scope, error) {
	updated := &Scope{}
	if wf := wfi.Workflow(); wf != nil {
		updated.Workflow = formatWorkflow(wf)
	}
	if wfi != nil {
		invocationParams, err := typedvalues.UnwrapMapTypedValue(wfi.Spec.Inputs)
		if err != nil {
			return nil, errors.Wrap(err, "failed to format invocation inputs")
		}
		updated.Invocation = &InvocationScope{
			ObjectMetadata: formatMetadata(wfi.Metadata),
			Inputs:         invocationParams,
		}
	}

	for taskId, task := range types.GetTasks(wfi) {
		if updated.Tasks == nil {
			updated.Tasks = map[string]*TaskScope{}
		}

		// Dep: pipe output of dynamic tasks
		t := controlflow.ResolveTaskOutput(taskId, wfi)
		output, err := typedvalues.Unwrap(t)
		if err != nil {
			panic(err)
		}

		h := controlflow.ResolveTaskOutputHeaders(taskId, wfi)
		outputHeaders, err := typedvalues.Unwrap(h)
		if err != nil {
			panic(err)
		}
		inputs, err := typedvalues.UnwrapMapTypedValue(task.Spec.Inputs)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to format inputs of task %v", taskId)
		}
		updated.Tasks[taskId] = &TaskScope{
			ObjectMetadata: formatMetadata(task.Metadata),
			Status:         task.Status.Status.String(),
			UpdatedAt:      formatTimestamp(task.Status.UpdatedAt),
			Inputs:         inputs,
			Requires:       task.Spec.Requires,
			Output:         output,
			OutputHeaders:  outputHeaders,
			Function:       task.Spec.FunctionRef,
		}
	}

	if base == nil {
		return updated, nil
	}
	if base.Tasks == nil {
		base.Tasks = map[string]*TaskScope{}
	}
	err := mergo.Merge(updated, base)
	if err != nil {
		return nil, err
	}
	return updated, nil
}

func formatWorkflow(wf *types.Workflow) *WorkflowScope {
	return &WorkflowScope{
		ObjectMetadata: formatMetadata(wf.Metadata),
		UpdatedAt:      formatTimestamp(wf.Status.UpdatedAt),
		Status:         wf.Status.Status.String(),
		Name:           wf.GetMetadata().GetName(),
		Internal:       wf.GetSpec().GetInternal(),
	}
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

func DeepCopy(i interface{}) interface{} {
	if i == nil {
		return i
	}
	switch t := i.(type) {
	// TODO support any function as primitive (use reflection API)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, uintptr,
		complex64, complex128, string, bool:
		return t
	case DeepCopier:
		return t.DeepCopy()
	case map[string]interface{}: // TODO support any map
		copied := make(map[string]interface{}, len(t))
		for k, v := range t {
			copied[k] = DeepCopy(v)
		}
		return copied
	case []interface{}: // TODO support any array
		copied := make([]interface{}, len(t))
		for k, v := range t {
			copied[k] = DeepCopy(v)
		}
		return copied
	case proto.Message:
		return proto.Clone(t)
	default:
		panic(fmt.Sprintf("cannot deepcopy unknown type %T", t))
	}
}
