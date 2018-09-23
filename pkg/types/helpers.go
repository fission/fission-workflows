package types

import (
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/ptypes"
)

// GetTasks gets all tasks in a workflow. This includes the dynamic tasks added during
// the invocation.
func GetTasks(wf *Workflow, wfi *WorkflowInvocation) map[string]*Task {
	tasks := map[string]*Task{}
	for _, task := range wf.Tasks() {
		tasks[task.ID()] = task
	}
	if wfi != nil {
		for id := range wfi.Status.DynamicTasks {
			task, _ := GetTask(wf, wfi, id)
			tasks[task.ID()] = task
		}
	}
	return tasks
}

func GetTaskContainers(wf *Workflow, wfi *WorkflowInvocation) map[string]*TaskInstance {
	tasks := map[string]*TaskInstance{}
	for _, task := range wf.Tasks() {
		id := task.ID()
		i, ok := wfi.TaskInvocation(id)
		if !ok {
			i = NewTaskInvocation(id)
		}
		tasks[task.ID()] = &TaskInstance{
			Task:       task,
			Invocation: i,
		}
	}
	if wfi != nil {
		for id := range wfi.Status.DynamicTasks {
			i, ok := wfi.TaskInvocation(id)
			if !ok {
				i = NewTaskInvocation(id)
			}
			task, _ := GetTask(wf, wfi, id)
			tasks[task.ID()] = &TaskInstance{
				Task:       task,
				Invocation: i,
			}
		}
	}
	return tasks
}

// GetTask gets the task associated with the id. Both static and dynamic tasks are searched.
func GetTask(wf *Workflow, wfi *WorkflowInvocation, id string) (*Task, bool) {
	task := wfi.Status.DynamicTasks[id]
	if task != nil {
		return task, true
	}

	spec, ok := GetTaskSpec(wf, wfi, id)
	if !ok {
		return nil, false
	}

	// Find TaskStatus
	status, ok := wf.Status.Tasks[id]
	if !ok {
		status = &TaskStatus{
			UpdatedAt: ptypes.TimestampNow(),
		}
	}

	return &Task{
		Metadata: &ObjectMetadata{
			Id: id,
			// TODO createdAt is not true for dynamic tasks
			CreatedAt: wf.Metadata.CreatedAt,
		},
		Spec:   spec,
		Status: status,
	}, true
}

func GetTaskSpec(wf *Workflow, wfi *WorkflowInvocation, id string) (*TaskSpec, bool) {
	// Find TaskSpec and overlay if needed
	spec, ok := wf.Spec.Tasks[id]
	var dtask *Task
	var dok bool
	if wfi != nil {
		dtask, dok = wfi.Status.DynamicTasks[id]
	}
	if !ok && !dok {
		return nil, false
	}
	if dok {
		if ok {
			spec = spec.Overlay(dtask.Spec)
		} else {
			spec = dtask.Spec
		}
	}
	return spec, true
}

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

type WorkflowInstance struct {
	Workflow *Workflow

	// Invocation is nil if not yet invoked
	Invocation *WorkflowInvocation
}

type TaskInstance struct {
	Task *Task
	// Invocation is nil if not yet invoked
	Invocation *TaskInvocation
}

type NamedTypedValue struct {
	typedvalues.TypedValue
	name string
}

type Inputs map[string]*typedvalues.TypedValue

func NewTaskInvocationSpec(invocationId string, taskId string, fnRef FnRef) *TaskInvocationSpec {
	return &TaskInvocationSpec{
		FnRef:        &fnRef,
		TaskId:       taskId,
		InvocationId: invocationId,
	}
}
