package workflow

import (
	"errors"

	"fmt"
	"strings"

	"github.com/fission/fission-workflow/pkg/types"
)

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

// Pre-parse
func (vl *Validator) Validate(spec *types.WorkflowSpec) error {
	if len(spec.GetName()) == 0 {
		return errors.New("Workflow requires a name")
	}

	// Version is not important right now

	src := spec.GetSrc()

	if src == nil {
		return errors.New("No source (src) provided")
	}

	if len(src.ApiVersion) > 0 && !strings.EqualFold(src.GetApiVersion(), "v1") {
		return fmt.Errorf("Unknown API version: '%v'.", src.GetApiVersion())
	}

	refTable := map[string]*types.Task{}

	for taskId, task := range src.Tasks {
		if len(taskId) == 0 {
			return errors.New("Task requires an id")
		}

		if len(task.Name) == 0 {
			return errors.New("Task requires an name")
		}

		_, ok := refTable[taskId]
		if ok {
			return fmt.Errorf("Task id '%s' is not unique", taskId)
		}
		refTable[taskId] = task

		if len(task.GetType()) > 0 && !strings.EqualFold(task.GetType(), "function") {
			return fmt.Errorf("Unknown task type '%s'", task.GetType())
		}
	}

	if len(src.Tasks) == 0 {
		return errors.New("Workflow needs at least one task")
	}

	// TODO check for cycles (?)
	// Dependency testing
	startTasks := []*types.Task{}
	for taskId, task := range src.GetTasks() {
		if len(task.GetDependencies()) == 0 {
			startTasks = append(startTasks, task)
		}

		for depName, _ := range task.GetDependencies() {
			if _, ok := refTable[depName]; !ok {
				return fmt.Errorf("Task '%s' ocntains undefined dependency '%s'", taskId, depName)
			}
		}
	}

	// Check if there are starting points
	if len(startTasks) == 0 {
		return errors.New("Workflow does not contain any start tasks (tasks with no dependencies)")
	}

	// Check output task (TODO make this implicit)
	if _, ok := src.GetTasks()[src.OutputTask]; !ok {
		return fmt.Errorf("Invalid outputTask '%s'", src.OutputTask)
	}

	return nil
}
