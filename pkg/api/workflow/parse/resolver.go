package parse

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
)

const (
	ProviderSeparator = ":"
	defaultTimeout    = time.Duration(1) * time.Minute
)

// Resolver contacts function execution runtime clients to resolve the function definitions to concrete function ids.
//
// ParseTask definitions (See types/TaskDef) can contain the following function reference:
// - `<name>` : the function is currently resolved to one of the clients
// - `<client>:<name>` : forces the client that the function needs to be resolved to.
//
// Future:
// - Instead of resolving just to one client, resolve function for all clients, and apply a priority or policy
//   for scheduling (overhead vs. load)
//
type Resolver struct {
	clients map[string]fnenv.Resolver
	timeout time.Duration
}

func NewResolver(client map[string]fnenv.Resolver) *Resolver {
	return &Resolver{
		clients: client,
		timeout: defaultTimeout,
	}
}

type resolvedFn struct {
	fnRef     string
	fnTypeDef *types.ResolvedTask
}

func (ps *Resolver) ResolveTask(task *types.TaskSpec) (*types.ResolvedTask, error) {
	resolved, err := ps.Resolve(task)
	if err != nil {
		return nil, err
	}
	return resolved[task.FunctionRef], nil
}

func (ps *Resolver) ResolveMap(tasks map[string]*types.TaskSpec) (map[string]*types.ResolvedTask, error) {
	var flattened []*types.TaskSpec
	for _, v := range tasks {
		flattened = append(flattened, v)
	}
	return ps.Resolve(flattened...)
}

// Resolve parses the interpreted workflow from a given spec.
//
// It returns a map consisting of the functionRef as the key for easier lookups
func (ps *Resolver) Resolve(tasks ...*types.TaskSpec) (map[string]*types.ResolvedTask, error) {
	// Check for duplicates
	uniqueTasks := map[string]*types.TaskSpec{}
	for _, t := range tasks {
		if _, ok := uniqueTasks[t.FunctionRef]; !ok {
			uniqueTasks[t.FunctionRef] = t
		}
	}

	var lastErr error
	wg := sync.WaitGroup{}
	wg.Add(len(uniqueTasks))
	resolved := map[string]*types.ResolvedTask{}
	resolvedC := make(chan resolvedFn, len(uniqueTasks))

	// Resolve each task in the workflow definition in parallel
	for _, t := range uniqueTasks {
		go func(t *types.TaskSpec, tc chan resolvedFn) {
			err := ps.resolveTaskAndInputs(t, tc)
			if err != nil {
				lastErr = err
			}
			wg.Done()
		}(t, resolvedC)
	}

	// Close channel when all tasks have resolved
	go func() {
		wg.Wait()
		close(resolvedC)
	}()

	// Store results of the resolved tasks
	for t := range resolvedC {
		resolved[t.fnRef] = t.fnTypeDef
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return resolved, nil
}

func (ps *Resolver) resolveTaskAndInputs(task *types.TaskSpec, resolvedC chan resolvedFn) error {
	t, err := ps.resolveTask(task)
	if err != nil {
		return err
	}
	resolvedC <- resolvedFn{task.FunctionRef, t}
	for _, input := range task.Inputs {
		switch input.Type {
		case typedvalues.TypeTask:
			f, err := typedvalues.FormatTask(input)
			if err != nil {
				return err
			}
			err = ps.resolveTaskAndInputs(f, resolvedC)
			if err != nil {
				return err
			}
		case typedvalues.TypeWorkflow:
			wf, err := typedvalues.FormatWorkflow(input)
			if err != nil {
				return err
			}
			for _, tasks := range wf.Tasks {
				err = ps.resolveTaskAndInputs(tasks, resolvedC)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (ps *Resolver) resolveTask(task *types.TaskSpec) (*types.ResolvedTask, error) {
	t := task.FunctionRef
	// Use clients to resolve task to id
	p, err := parseTaskAddress(t)
	if err != nil {
		return nil, err
	}
	if len(p.GetRuntime()) > 0 {
		return ps.resolveForRuntime(t, p.GetRuntime())
	}

	waitFor := len(ps.clients)
	resolved := make(chan *types.ResolvedTask, waitFor)
	defer close(resolved)
	wg := sync.WaitGroup{}
	wg.Add(waitFor)
	var lastErr error
	for cName := range ps.clients {
		go func(cName string) {
			def, err := ps.resolveForRuntime(t, cName)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"err":     err,
					"runtime": cName,
					"fn":      t,
				}).Info("Failed to retrieve function.")
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
		return nil, fmt.Errorf("failed to resolve function '%s' using clients '%v': %v", t, ps.clients, lastErr)
	}
}

func (ps *Resolver) resolveForRuntime(taskName string, runtime string) (*types.ResolvedTask, error) {
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
	return &types.ResolvedTask{
		Src:      taskName,
		Runtime:  runtime,
		Resolved: fnId,
	}, nil
}

// Format: <runtime>:<name> for now
func parseTaskAddress(task string) (*types.ResolvedTask, error) {
	parts := strings.SplitN(task, ProviderSeparator, 2)

	switch len(parts) {
	case 0:
		return nil, fmt.Errorf("could not parse invalid task address '%s'", task)
	case 1:
		return &types.ResolvedTask{
			Src: parts[0],
		}, nil
	default:
		return &types.ResolvedTask{
			Src:     parts[1],
			Runtime: parts[0],
		}, nil
	}
}
