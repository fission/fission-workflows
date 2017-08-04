package workflow

import (
	"fmt"
	"strings"

	"github.com/fission/fission"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission/controller/client"
)

type Parser struct {
	client *client.Client // TODO abstract away fission
}

func NewParser(client *client.Client) *Parser {
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
	fn, err := ps.client.FunctionGet(&fission.Metadata{
		Name: name,
	})
	if err != nil {
		return nil, err
	}

	return &types.TaskTypeDef{
		Src:      task.GetName(),
		Resolved: fn.Uid,
	}, nil
}
