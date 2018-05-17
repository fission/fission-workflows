package types

import (
	"github.com/golang/protobuf/ptypes"
)

// GetTasks gets all tasks in a workflow. This includes the dynamic tasks added during
// the invocation.
func GetTasks(wf *Workflow, wfi *WorkflowInvocation) map[string]*Task {
	tasks := map[string]*Task{}
	for _, task := range wf.Tasks() {
		tasks[task.Id()] = task
	}
	if wfi != nil {
		for id := range wfi.Status.DynamicTasks {
			task, _ := GetTask(wf, wfi, id)
			tasks[task.Id()] = task
		}
	}
	return tasks
}

func GetTaskContainers(wf *Workflow, wfi *WorkflowInvocation) map[string]*TaskInstance {
	tasks := map[string]*TaskInstance{}
	for _, task := range wf.Tasks() {
		id := task.Id()
		i, ok := wfi.TaskInvocation(id)
		if !ok {
			i = NewTaskInvocation(id)
		}
		tasks[task.Id()] = &TaskInstance{
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
			tasks[task.Id()] = &TaskInstance{
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

func NewWorkflowInvocation(wfId string, invocId string) *WorkflowInvocation {
	return &WorkflowInvocation{
		Metadata: NewObjectMetadata(invocId),
		Spec:     NewWorkflowInvocationSpec(wfId),
		Status:   &WorkflowInvocationStatus{},
	}
}

func NewWorkflowInvocationSpec(wfId string) *WorkflowInvocationSpec {
	return &WorkflowInvocationSpec{
		WorkflowId: wfId,
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

func SingleInput(key string, t *TypedValue) map[string]*TypedValue {
	return map[string]*TypedValue{
		key: t,
	}
}

func SingleDefaultInput(t *TypedValue) map[string]*TypedValue {
	return SingleInput(INPUT_MAIN, t)
}

type Requires map[string]*TaskDependencyParameters

func (r Requires) Add(s ...string) Requires {
	if s != nil {
		for _, v := range s {
			r[v] = nil
		}
	}
	return r
}

func Require(s ...string) Requires {
	return Requires{}.Add(s...)
}

func NewWorkflowSpec() *WorkflowSpec {
	return &WorkflowSpec{
		ApiVersion: WorkflowApiVersion,
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
	TypedValue
	name string
}

type Inputs map[string]*TypedValue
