package types

import (
	"time"

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

func NewTaskInvocationSpec(invocation *WorkflowInvocation, task *Task, startAt time.Time) *TaskInvocationSpec {
	// Decide on the deadline of the task invocation.
	// If there is no timeout specified for the task, use the deadline of the overall workflow invocation.
	// If there is a timeout specified for the task, calculate the deadline using the startAt + timout time, but do not
	// exceed the invocation deadline.
	deadline := invocation.GetSpec().GetDeadline()
	if task.GetSpec().GetTimeout() != nil {
		deadlineTime, err := ptypes.Timestamp(deadline)
		maxRuntime, err := ptypes.Duration(task.GetSpec().GetTimeout())
		if err == nil {
			taskMaxRuntime := startAt.Add(maxRuntime)
			if taskMaxRuntime.Before(deadlineTime) {
				deadline, _ = ptypes.TimestampProto(taskMaxRuntime)
			}
		}
	}

	return &TaskInvocationSpec{
		InvocationId: invocation.ID(),
		Task:         task,
		FnRef:        task.GetStatus().GetFnRef(),
		TaskId:       task.ID(),
		Deadline:     deadline,
		Inputs:       task.GetSpec().GetInputs(),
	}
}

func Input(val interface{}) map[string]*typedvalues.TypedValue {
	return map[string]*typedvalues.TypedValue{
		InputMain: typedvalues.MustWrap(val),
	}
}
