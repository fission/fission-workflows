// Validate package contains validation functions for the common structures used in the
// workflow engine, such as Workflows, Tasks, WorkflowInvocations, etc.
package validate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/graph"
	"gonum.org/v1/gonum/graph/topo"
)

var (
	ErrObjectEmpty                  = errors.New("no object provided")
	ErrInvalidApiVersion            = errors.New("unknown API version")
	ErrWorkflowWithoutTasks         = errors.New("workflow needs at least one task")
	ErrUndefinedDependency          = errors.New("task contains undefined dependency")
	ErrTaskIdMissing                = errors.New("task misses an id")
	ErrTaskNotUnique                = errors.New("task is not unique")
	ErrWorkflowWithoutStartTasks    = errors.New("workflow does not contain any start tasks (tasks with 0 dependencies)")
	ErrTaskRequiresFnRef            = errors.New("task requires a function name")
	ErrCircularDependency           = errors.New("workflow contains circular dependency")
	ErrInvalidOutputTask            = errors.New("unknown output task")
	ErrNoParentTaskDependency       = errors.New("dynamic task does not contain parent dependency")
	ErrMultipleParentTaskDependency = errors.New("dynamic task contains multiple parent tasks")
	ErrNoWorkflowInvocation         = errors.New("workflow invocation id is required")
	ErrNoFnRef                      = errors.New("function reference is required")
	ErrNoWorkflow                   = errors.New("workflow id is required")
	ErrNoId                         = errors.New("id is required")
	ErrNoStatus                     = errors.New("status is required")
)

type Error struct {
	val  string
	errs []error
}

// Note that this does not take nested Error's into account
func (ie Error) Reasons() []error {
	return ie.errs
}

func (ie Error) Error() string {
	vt := ie.val
	if len(vt) == 0 {
		vt = "Value"
	}
	return fmt.Sprintf("%s is invalid", vt)
}

func (ie *Error) IsEmpty() bool {
	return ie.errs == nil || len(ie.errs) == 0
}

func (ie Error) GetOrNil() error {
	if ie.IsEmpty() {
		return nil
	} else {
		return ie
	}
}

func (ie *Error) append(err error) {
	if err == nil {
		return
	}
	if ie.errs == nil {
		ie.errs = []error{err}
	} else {
		ie.errs = append(ie.errs, err)
	}
}

// WorkflowSpec validates the Workflow Specification.
func WorkflowSpec(spec *types.WorkflowSpec) error {
	errs := Error{val: "WorkflowSpec"}

	if spec == nil {
		errs.append(ErrObjectEmpty)
		return errs.GetOrNil()
	}

	if len(spec.ApiVersion) > 0 && !strings.EqualFold(spec.GetApiVersion(), types.WorkflowApiVersion) {
		errs.append(fmt.Errorf("%v: '%v'", ErrInvalidApiVersion, spec.GetApiVersion()))
	}

	if len(spec.Tasks) == 0 {
		errs.append(ErrWorkflowWithoutTasks)
	}

	_, ok := spec.Tasks[spec.OutputTask]
	if !ok {
		errs.append(ErrInvalidOutputTask)
	}

	refTable := map[string]*types.TaskSpec{}
	for taskId, task := range spec.Tasks {
		if len(taskId) == 0 {
			errs.append(ErrTaskIdMissing)
		}

		errs.append(TaskSpec(task))

		_, ok := refTable[taskId]
		if ok {
			errs.append(fmt.Errorf("%v: '%v'", ErrTaskNotUnique, taskId))
		}
		refTable[taskId] = task
	}

	// Dependency testing
	var startTasks []*types.TaskSpec
	for taskId, task := range spec.GetTasks() {
		if len(task.Requires) == 0 {
			startTasks = append(startTasks, task)
		}

		// Check for undefined dependencies
		for depName := range task.Requires {
			if _, ok := refTable[depName]; !ok {
				errs.append(fmt.Errorf("%v: '%v->%v'", ErrUndefinedDependency, taskId, depName))
			}
		}
	}

	// Check for circular dependencies
	dg := graph.Parse(graph.NewTaskSpecIterator(spec.Tasks))
	if len(topo.DirectedCyclesIn(dg)) > 0 {
		errs.append(ErrCircularDependency)
	}

	// Check if there are starting points
	if len(startTasks) == 0 {
		errs.append(ErrWorkflowWithoutStartTasks)
	}

	return errs.GetOrNil()
}

func TaskSpec(spec *types.TaskSpec) error {
	errs := Error{val: "TaskSpec"}

	if spec == nil {
		errs.append(ErrObjectEmpty)
		return errs.GetOrNil()
	}

	if len(spec.FunctionRef) == 0 {
		errs.append(ErrTaskRequiresFnRef)
	}

	return errs.GetOrNil()
}

func DynamicTaskSpec(task *types.TaskSpec) error {
	err := TaskSpec(task)
	if err != nil {
		return err
	}
	errs := Error{
		val: "TaskSpec.dynamic",
	}

	// Check if there is a parent link
	var parents int
	for _, params := range task.Requires {
		if params.Type == types.TaskDependencyParameters_DYNAMIC_OUTPUT {
			parents += 1
		}
	}
	if parents == 0 {
		errs.append(ErrNoParentTaskDependency)
	} else if parents > 1 {
		errs.append(ErrMultipleParentTaskDependency)
	}
	return errs.GetOrNil()
}

func Task(task *types.Task) error {
	errs := Error{val: "Task"}
	if task == nil {
		errs.append(ErrObjectEmpty)
		return errs.GetOrNil()
	}

	errs.append(TaskSpec(task.Spec))
	errs.append(ObjectMetadata(task.Metadata))

	if task.Status == nil {
		errs.append(ErrNoStatus)
	}

	return errs.GetOrNil()
}

func ObjectMetadata(o *types.ObjectMetadata) error {
	errs := Error{val: "WorkflowInvocationSpec"}

	if o == nil {
		errs.append(ErrObjectEmpty)
		return errs.GetOrNil()
	}
	if len(o.Id) == 0 {
		errs.append(ErrNoId)
	}

	return errs.GetOrNil()
}

func WorkflowInvocationSpec(spec *types.WorkflowInvocationSpec) error {
	errs := Error{val: "WorkflowInvocationSpec"}

	if spec == nil {
		errs.append(ErrObjectEmpty)
		return errs.GetOrNil()
	}

	if len(spec.WorkflowId) == 0 {
		errs.append(ErrNoWorkflow)
	}

	return errs.GetOrNil()
}

func TaskInvocationSpec(spec *types.TaskInvocationSpec) error {
	errs := Error{val: "TaskInvocationSpec"}

	if spec == nil {
		errs.append(ErrObjectEmpty)
		return errs.GetOrNil()
	}

	if len(spec.InvocationId) == 0 {
		errs.append(ErrNoWorkflowInvocation)
	}

	if spec.FnRef == nil {
		errs.append(ErrNoFnRef)
	}
	return errs.GetOrNil()
}

// Format is an utility for printing the error hierarchies of Error
func Format(rawErr error) string {
	switch err := rawErr.(type) {
	case Error:
		result := err.Error() + "\n"
		for _, reason := range err.Reasons() {
			var formatted string
			switch t := reason.(type) {
			case Error:
				reasons := Format(t)
				reasons = strings.Replace(reasons, "\n", "\n  ", -1)
				reasons = reasons[:len(reasons)-2]
				formatted = fmt.Sprintf("- %v\n", reasons)
			default:
				formatted = fmt.Sprintf("- %v\n", reason)
			}
			result += formatted
		}

		return strings.TrimSpace(result)
	default:
		return fmt.Sprintf("%v", err)
	}
}
