package workflow

import (
	"fmt"
	"strings"

	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/sirupsen/logrus"
)

type Parser struct {
	clients map[string]function.Resolver
	// TODO allow aliases
}

func NewParser(client map[string]function.Resolver) *Parser {
	return &Parser{client}
}

// Parse parses the interpreted workflow from a given spec.
func (ps *Parser) Parse(spec *types.WorkflowSpec) (*types.WorkflowStatus, error) {

	// TODO make resolving parallel
	taskTypes := map[string]*types.TaskTypeDef{}
	for taskId, task := range spec.GetTasks() {
		if len(task.GetType()) > 0 && !strings.EqualFold(task.GetType(), "function") {
			return nil, fmt.Errorf("Unknown type: '%s'", task.GetType())
		}

		taskDef, err := ps.resolveTask(task)
		if err != nil {
			return nil, err
		}
		taskTypes[taskId] = taskDef
	}

	return &types.WorkflowStatus{
		ResolvedTasks: taskTypes,
	}, nil
}

func (ps *Parser) resolveTask(task *types.Task) (*types.TaskTypeDef, error) {
	t := task.GetName()
	// Use clients to resolve task to id
	parsed, _ := parseTaskAddress(t)
	if len(parsed.GetRuntime()) > 0 {
		return ps.resolveForRuntime(t, parsed.GetRuntime())
	}

	// TODO resolve function for all clients, and apply a priority or policy for scheduling (overhead vs. load)
	for cName := range ps.clients {
		def, err := ps.resolveForRuntime(t, cName)
		if err == nil {
			return def, nil
		}
	}
	return nil, fmt.Errorf("Failed to resolve function '%s' using clients '%v'.", t, ps.clients)
}

func (ps *Parser) resolveForRuntime(taskName string, runtime string) (*types.TaskTypeDef, error) {
	resolver, ok := ps.clients[runtime]
	if !ok {
		return nil, fmt.Errorf("Runtime '%s' could not be found.", runtime)
	}

	fnId, err := resolver.Resolve(taskName)
	if err != nil {
		return nil, err
	}

	logrus.WithFields(logrus.Fields{
		"src":      taskName,
		"runtime":  runtime,
		"resolved": fnId,
	}).Info("Resolved task")
	return &types.TaskTypeDef{
		Src:      taskName,
		Runtime:  runtime,
		Resolved: fnId,
	}, nil
}

// Format: <runtime>:<name> for now
func parseTaskAddress(task string) (*types.TaskTypeDef, error) {
	parts := strings.SplitN(task, ":", 2)
	if len(parts) == 1 {
		return &types.TaskTypeDef{
			Src: parts[0],
		}, nil
	}

	return &types.TaskTypeDef{
		Src:     parts[1],
		Runtime: parts[0],
	}, nil
}
