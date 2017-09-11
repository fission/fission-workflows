package parse

import (
	"fmt"
	"strings"

	"sync"

	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
)

// Resolver contacts function execution runtime clients to resolve the function definitions to concrete function ids.
//
// Task definitions (See types/TaskDef) can contain the following function reference:
// - `<name>` : the function is currently resolved to one of the clients
// - `<client>:<name>` : forces the client that the function needs to be resolved to.
//
// Future:
// - Instead of resolving just to one client, resolve function for all clients, and apply a priority or policy
//   for scheduling (overhead vs. load)
//
type Resolver struct {
	clients map[string]function.Resolver
}

func NewResolver(client map[string]function.Resolver) *Resolver {
	return &Resolver{client}
}

// Resolve parses the interpreted workflow from a given spec.
func (ps *Resolver) Resolve(spec *types.WorkflowSpec) (*types.WorkflowStatus, error) {

	taskSize := len(spec.GetTasks())
	wg := sync.WaitGroup{}
	wg.Add(taskSize)
	var lastErr error

	taskTypes := map[string]*types.TaskTypeDef{}
	for taskId, task := range spec.GetTasks() {
		go func(taskId string, task *types.Task) {
			lastErr = ps.resolveTaskAndInputs(task, taskTypes)
			wg.Done()
		}(taskId, task)
	}
	wg.Wait() // for all tasks to be resolved

	if lastErr != nil {
		return nil, lastErr
	}
	return &types.WorkflowStatus{
		ResolvedTasks: taskTypes,
	}, nil
}

func (ps *Resolver) resolveTaskAndInputs(task *types.Task, resolved map[string]*types.TaskTypeDef) error {
	t, err := ps.resolveTask(task)
	if err != nil {
		return err
	}
	resolved[task.Name] = t
	for _, input := range task.Inputs {
		if input.Type == typedvalues.TYPE_FLOW {
			f, err := typedvalues.Format(input)
			if err != nil {
				return err
			}
			switch it := f.(type) {
			case *types.Task:
				err = ps.resolveTaskAndInputs(it, resolved)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (ps *Resolver) resolveTask(task *types.Task) (*types.TaskTypeDef, error) {
	t := task.GetName()
	// Use clients to resolve task to id
	p, err := parseTaskAddress(t)
	if err != nil {
		return nil, err
	}
	if len(p.GetRuntime()) > 0 {
		return ps.resolveForRuntime(t, p.GetRuntime())
	}

	waitFor := len(ps.clients)
	resolved := make(chan *types.TaskTypeDef, waitFor)
	wg := sync.WaitGroup{}
	wg.Add(waitFor)
	var lastErr error
	for cName := range ps.clients {
		go func(cName string) {
			def, err := ps.resolveForRuntime(t, cName)
			if err != nil {
				logrus.Error(err)
				lastErr = err
			} else {
				resolved <- def
			}
			wg.Done()
		}(cName)
	}
	wg.Wait() // for all clients to resolve

	// For now just select the first resolved
	select {
	case result := <-resolved:
		return result, nil
	default:
		return nil, fmt.Errorf("failed to resolve function '%s' using clients '%v'", t, ps.clients)
	}
}

func (ps *Resolver) resolveForRuntime(taskName string, runtime string) (*types.TaskTypeDef, error) {
	resolver, ok := ps.clients[runtime]
	if !ok {
		return nil, fmt.Errorf("runtime '%s' could not be found", runtime)
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

	switch len(parts) {
	case 0:
		return nil, fmt.Errorf("Could not parse invalid task address '%s'", task)
	case 1:
		return &types.TaskTypeDef{
			Src: parts[0],
		}, nil
	default:
		return &types.TaskTypeDef{
			Src:     parts[1],
			Runtime: parts[0],
		}, nil
	}
}
