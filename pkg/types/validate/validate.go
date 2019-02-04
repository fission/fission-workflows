// Validate package contains validation functions for the common structures used in the
// workflow engine, such as Workflows, Tasks, WorkflowInvocations, etc.
//
// All validate functions return either nil (value is valid) or a validate.Error (value is invalid).
package validate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/graph"
	"github.com/fission/fission-workflows/pkg/types/typedvalues/controlflow"
	"gonum.org/v1/gonum/graph/topo"
)

var (
	ErrObjectEmpty                  = errors.New("no object provided")
	ErrInvalidAPIVersion            = errors.New("unknown API version")
	ErrWorkflowWithoutTasks         = errors.New("workflow needs at least one task")
	ErrUndefinedDependency          = errors.New("task contains undefined dependency")
	ErrTaskIDMissing                = errors.New("task misses an id")
	ErrTaskNotUnique                = errors.New("task is not unique")
	ErrWorkflowWithoutStartTasks    = errors.New("workflow does not contain any start tasks (tasks with 0 dependencies)")
	ErrTaskRequiresFnRef            = errors.New("task requires a function name")
	ErrCircularDependency           = errors.New("workflow contains circular dependency")
	ErrInvalidOutputTask            = errors.New("unknown output task")
	ErrNoParentTaskDependency       = errors.New("dynamic task does not contain parent dependency")
	ErrMultipleParentTaskDependency = errors.New("dynamic task contains multiple parent tasks")
	ErrNoWorkflowInvocation         = errors.New("workflow invocation id is required")
	ErrNoTaskInvocation             = errors.New("task invocation id is required")
	ErrNoFnRef                      = errors.New("function reference is required")
	ErrNoWorkflow                   = errors.New("workflow id is required")
	ErrNoID                         = errors.New("id is required")
	ErrNoStatus                     = errors.New("status is required")
	ErrWorkflowNotReady             = errors.New("workflow is not ready")
)

type Error struct {
	subject string
	errs    []error
}

func (ie Error) Reasons() []error {
	var reasons []error
	for _, e := range ie.errs {
		switch nestedErr := e.(type) {
		case Error:
			reasons = append(reasons, nestedErr.Reasons()...)
		default:
			reasons = append(reasons, nestedErr)
		}
	}
	return ie.errs
}

func (ie Error) Contains(err error) bool {
	for _, e := range ie.errs {
		switch nestedErr := e.(type) {
		case Error:
			if nestedErr.Contains(err) {
				return true
			}
		default:
			if e == err {
				return true
			}
		}
	}
	return false
}

func (ie Error) Error() string {
	vt := ie.subject
	if len(vt) == 0 {
		vt = "Value"
	}
	prefix := fmt.Sprintf("%s is invalid", vt)
	var rs []string
	for k, reason := range ie.Reasons() {
		rs = append(rs, fmt.Sprintf("(%d) %v", k, reason))
	}
	return prefix + ": " + strings.Join(rs, "; ")
}

func (ie Error) getOrNil() error {
	if ie.errs == nil {
		return nil
	}
	return ie
}

func (ie *Error) append(err error) {
	if err == nil {
		return
	}
	ie.errs = append(ie.errs, err)
}

// WorkflowSpec validates the Workflow Specification.
func WorkflowSpec(spec *types.WorkflowSpec) error {
	errs := Error{subject: "WorkflowSpec"}

	if spec == nil {
		errs.append(ErrObjectEmpty)
		return errs.getOrNil()
	}

	if len(spec.ApiVersion) > 0 && !strings.EqualFold(spec.GetApiVersion(), types.WorkflowAPIVersion) {
		errs.append(fmt.Errorf("%v: '%v'", ErrInvalidAPIVersion, spec.GetApiVersion()))
	}

	if len(spec.Tasks) == 0 {
		errs.append(ErrWorkflowWithoutTasks)
	}

	_, ok := spec.Tasks[spec.OutputTask]
	if !ok {
		errs.append(ErrInvalidOutputTask)
	}

	refTable := map[string]*types.TaskSpec{}
	for taskID, task := range spec.Tasks {
		if len(taskID) == 0 {
			errs.append(ErrTaskIDMissing)
		}

		errs.append(TaskSpec(task))

		_, ok := refTable[taskID]
		if ok {
			errs.append(fmt.Errorf("%v: '%v'", ErrTaskNotUnique, taskID))
		}
		refTable[taskID] = task
	}

	// Dependency testing
	var startTasks []*types.TaskSpec
	for taskID, task := range spec.GetTasks() {
		if len(task.Requires) == 0 {
			startTasks = append(startTasks, task)
		}

		// Check for undefined dependencies
		for depName := range task.Requires {
			if _, ok := refTable[depName]; !ok {
				errs.append(fmt.Errorf("%v: '%v->%v'", ErrUndefinedDependency, taskID, depName))
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

	return errs.getOrNil()
}

func TaskSpec(spec *types.TaskSpec) error {
	errs := Error{subject: "TaskSpec"}

	if spec == nil {
		errs.append(ErrObjectEmpty)
		return errs.getOrNil()
	}

	if len(spec.FunctionRef) == 0 {
		errs.append(ErrTaskRequiresFnRef)
	}

	return errs.getOrNil()
}

func DynamicTaskSpec(task *types.TaskSpec) error {
	err := TaskSpec(task)
	if err != nil {
		return err
	}
	errs := Error{
		subject: "TaskSpec.dynamic",
	}

	// Check if there is a parent link
	var parents int
	for _, params := range task.Requires {
		if params.Type == types.TaskDependencyParameters_DYNAMIC_OUTPUT {
			parents++
		}
	}
	if parents == 0 {
		errs.append(ErrNoParentTaskDependency)
	} else if parents > 1 {
		errs.append(ErrMultipleParentTaskDependency)
	}
	return errs.getOrNil()
}

func Task(task *types.Task) error {
	errs := Error{subject: "Task"}
	if task == nil {
		errs.append(ErrObjectEmpty)
		return errs.getOrNil()
	}

	errs.append(TaskSpec(task.Spec))
	errs.append(ObjectMetadata(task.Metadata))

	if task.Status == nil {
		errs.append(ErrNoStatus)
	}

	return errs.getOrNil()
}

func ObjectMetadata(o *types.ObjectMetadata) error {
	errs := Error{subject: "WorkflowInvocationSpec"}

	if o == nil {
		errs.append(ErrObjectEmpty)
		return errs.getOrNil()
	}
	if len(o.Id) == 0 {
		errs.append(ErrNoID)
	}

	return errs.getOrNil()
}

func WorkflowInvocationSpec(spec *types.WorkflowInvocationSpec) error {
	errs := Error{subject: "WorkflowInvocationSpec"}

	if spec == nil {
		errs.append(ErrObjectEmpty)
		return errs.getOrNil()
	}

	if spec.Workflow != nil {
		wf := spec.GetWorkflow()
		errs.append(WorkflowSpec(wf.GetSpec()))
		if !wf.GetStatus().Ready() {
			errs.append(ErrWorkflowNotReady)
		}
		// TODO further validate workflow status
	} else if len(spec.WorkflowId) == 0 {
		errs.append(ErrNoWorkflow)
	}

	return errs.getOrNil()
}

func TaskInvocationSpec(spec *types.TaskInvocationSpec) error {
	errs := Error{subject: "TaskInvocationSpec"}

	if spec == nil {
		errs.append(ErrObjectEmpty)
		return errs.getOrNil()
	}

	if len(spec.InvocationId) == 0 {
		errs.append(ErrNoWorkflowInvocation)
	}

	if len(spec.GetTask().ID()) == 0 {
		errs.append(ErrNoTaskInvocation)
	}

	if spec.GetTask().FnRef() == nil {
		errs.append(ErrNoFnRef)
	}
	return errs.getOrNil()
}

// Format is an utility for printing the error hierarchies of Error
func Format(rawErr error) string {
	return strings.TrimSpace(format(rawErr, 0))
}

// FormatConcise returns a single line description of the errors.
func FormatConcise(rawErr error) string {
	return strings.Replace(Format(rawErr), "\n", ";", -1)
}

func format(rawErr error, depth int) string {
	indent := strings.Repeat("  ", depth)
	switch err := rawErr.(type) {
	case Error:
		result := indent + err.Error() + "\n"
		for _, nested := range err.errs {
			result += format(nested, depth+1)
		}
		if len(err.errs) == 0 {
			result += format(errors.New("unknown error"), depth+1)
		}
		return result
	default:
		return indent + fmt.Sprintf("%v", err) + "\n"
	}
}

func Flow(flow controlflow.Flow) error {
	if flow.IsEmpty() {
		return ErrObjectEmpty
	}
	wf := flow.GetWorkflow()
	if wf != nil {
		return WorkflowSpec(wf)
	}
	return TaskSpec(flow.GetTask())
}

func NewError(subject string, errs ...error) error {
	return Error{
		subject: subject,
		errs:    errs,
	}
}
