package workflow

import (
	"fmt"
	"strings"

	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/types"
)

type Parser struct {
	clients map[string]function.Resolver
}

func NewParser(client map[string]function.Resolver) *Parser {
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

// TODO support specific runtime
func (ps *Parser) parseTask(task *types.Task) (*types.TaskTypeDef, error) {
	// TODO Split up for different task types
	t := task.GetName()
	var err error
	var fnId, clientName string
	for cName, client := range ps.clients { // TODO priority-based or store all resolved functions
		fnId, err = client.Resolve(t)
		clientName = cName
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to resolve function '%s' using clients '%v', because '%v'.", t, ps.clients, err)
	}

	return &types.TaskTypeDef{
		Src:      task.GetName(),
		Runtime:  clientName,
		Resolved: fnId,
	}, nil
}
