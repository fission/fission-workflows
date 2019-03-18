package types

import (
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/ptypes"
)

func NewWorkflow(id string) *Workflow {
	return &Workflow{
		Metadata: &ObjectMetadata{
			Id:        id,
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec:   NewWorkflowSpec(),
		Status: NewWorkflowStatus(),
	}
}

func NewWorkflowStatus() *WorkflowStatus {
	return &WorkflowStatus{
		UpdatedAt: ptypes.TimestampNow(),
	}
}

func NewWorkflowInvocation(wfID string, invocationID string) *WorkflowInvocation {
	return &WorkflowInvocation{
		Metadata: NewObjectMetadata(invocationID),
		Spec:     NewWorkflowInvocationSpec(wfID),
		Status:   &WorkflowInvocationStatus{},
	}
}

func NewWorkflowInvocationSpec(wfID string) *WorkflowInvocationSpec {
	return &WorkflowInvocationSpec{
		WorkflowId: wfID,
	}
}

func NewTask(id string, fn string) *Task {
	return &Task{
		Metadata: NewObjectMetadata(id),
		Spec:     NewTaskSpec(fn),
		Status:   NewTaskStatus(),
	}
}

func NewTaskSpec(fn string) *TaskSpec {
	return &TaskSpec{
		FunctionRef: fn,
	}
}

func NewTaskStatus() *TaskStatus {
	return &TaskStatus{
		UpdatedAt: ptypes.TimestampNow(),
	}
}

func NewObjectMetadata(id string) *ObjectMetadata {
	return &ObjectMetadata{
		Id:        id,
		CreatedAt: ptypes.TimestampNow(),
	}
}

func NewTaskInvocation(id string) *TaskInvocation {
	return &TaskInvocation{
		Metadata: NewObjectMetadata(id),
		Spec:     &TaskInvocationSpec{},
		Status:   &TaskInvocationStatus{},
	}
}

func SingleInput(key string, t *typedvalues.TypedValue) map[string]*typedvalues.TypedValue {
	return map[string]*typedvalues.TypedValue{
		key: t,
	}
}

func SingleDefaultInput(t *typedvalues.TypedValue) map[string]*typedvalues.TypedValue {
	return SingleInput(InputMain, t)
}

type Requires map[string]*TaskDependencyParameters

func (r Requires) Add(s ...string) Requires {
	for _, v := range s {
		r[v] = nil
	}
	return r
}

func Require(s ...string) Requires {
	return Requires{}.Add(s...)
}

func NewWorkflowSpec() *WorkflowSpec {
	return &WorkflowSpec{
		ApiVersion: WorkflowAPIVersion,
	}
}

type Tasks map[string]*TaskSpec

type NamedTypedValue struct {
	typedvalues.TypedValue
	name string
}

type Inputs map[string]*typedvalues.TypedValue

func NewTaskInvocationSpec(invocationId string, task *Task) *TaskInvocationSpec {
	return &TaskInvocationSpec{
		InvocationId: invocationId,
		Task:         task,
		FnRef:        task.GetStatus().GetFnRef(),
		TaskId:       task.ID(),
	}
}

func Input(val interface{}) map[string]*typedvalues.TypedValue {
	return map[string]*typedvalues.TypedValue{
		InputMain: typedvalues.MustWrap(val),
	}
}
