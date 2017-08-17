package workflow

import (
	"fmt"
	"strings"

	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/types"
)

type Parser struct {
	client function.Resolver
}

func NewParser(client function.Resolver) *Parser {
	return &Parser{client}
}

// Parse parses the interpreted workflow from a given spec.
func (ps *Parser) Parse(spec *types.WorkflowSpec) (*types.WorkflowStatus, error) {
	src := spec.GetSrc()

	// TODO paralize this resolving
	taskTypes := map[string]*types.TaskTypeDef{}
	for taskId, task := range src.GetTasks() {
		if len(task.GetType()) > 0 && !strings.EqualFold(task.GetType(), "function") {
			return nil, fmt.Errorf("Unknown type: '%s'", task.GetType())
		}

		taskDef, err := ps.parseTask(task)
		if err != nil {
			return nil, err
		}
		taskTypes[taskId] = taskDef
	}

	return &types.WorkflowStatus{
		ResolvedTasks: taskTypes,
	}, nil
}

func (ps *Parser) parseTask(task *types.Task) (*types.TaskTypeDef, error) {
	// TODO Split up for different task types
	name := task.GetName()
	fnId, err := ps.client.Resolve(name)
	if err != nil {
		return nil, fmt.Errorf("Failed to resolve function '%s' using client '%v', because '%v'.", name, ps.client, err)
	}

	return &types.TaskTypeDef{
		Src:      task.GetName(),
		Resolved: fnId,
	}, nil
}
