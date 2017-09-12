package parse

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

	if spec == nil {
		return errors.New("No source (spec) provided")
	}

	if len(spec.ApiVersion) > 0 && !strings.EqualFold(spec.GetApiVersion(), "v1") {
		return fmt.Errorf("Unknown API version: '%v'.", spec.GetApiVersion())
	}

	refTable := map[string]*types.Task{}

	for taskId, task := range spec.Tasks {
		if len(taskId) == 0 {
			return errors.New("Task requires an id")
		}

		if len(task.FunctionRef) == 0 {
			return errors.New("Task requires a function name")
		}

		_, ok := refTable[taskId]
		if ok {
			return fmt.Errorf("Task id '%s' is not unique", taskId)
		}
		refTable[taskId] = task
	}

	if len(spec.Tasks) == 0 {
		return errors.New("Workflow needs at least one task")
	}

	// Dependency testing
	startTasks := []*types.Task{}
	for taskId, task := range spec.GetTasks() {
		if len(task.Requires) == 0 {
			startTasks = append(startTasks, task)
		}

		for depName := range task.Requires {
			if _, ok := refTable[depName]; !ok {
				return fmt.Errorf("Task '%s' ocntains undefined dependency '%s'", taskId, depName)
			}
		}
	}

	// Check if there are starting points
	if len(startTasks) == 0 {
		return errors.New("Workflow does not contain any start tasks (tasks with no dependencies)")
	}

	// Check output task
	if _, ok := spec.GetTasks()[spec.OutputTask]; !ok {
		return fmt.Errorf("Invalid outputTask '%s'", spec.OutputTask)
	}

	return nil
}
