package fnenv

import (
	"fmt"
	"sync"
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/sirupsen/logrus"
)

const (
	defaultTimeout = time.Duration(1) * time.Minute
)

// MetaResolver contacts function execution runtime clients to resolve the function definitions to concrete function ids.
//
// ParseTask definitions (See types/TaskDef) can contain the following function reference:
// - `<name>` : the function is currently resolved to one of the clients
// - `<client>:<name>` : forces the client that the function needs to be resolved to.
//
// Future:
// - Instead of resolving just to one client, resolve function for all clients, and apply a priority or policy
//   for scheduling (overhead vs. load)
//
type MetaResolver struct {
	clients map[string]RuntimeResolver
	timeout time.Duration
}

func NewMetaResolver(client map[string]RuntimeResolver) *MetaResolver {
	return &MetaResolver{
		clients: client,
		timeout: defaultTimeout,
	}
}

func (ps *MetaResolver) Resolve(targetFn string) (types.FnRef, error) {
	ref, err := types.ParseFnRef(targetFn)
	if err != nil {
		if err == types.ErrNoRuntime {
			ref = types.FnRef{
				ID: targetFn,
			}
		} else {
			return types.FnRef{}, err
		}
	}

	if ref.Runtime != "" {
		return ps.resolveForRuntime(ref.ID, ref.Runtime)
	}

	waitFor := len(ps.clients)
	resolved := make(chan types.FnRef, waitFor)
	defer close(resolved)
	wg := sync.WaitGroup{}
	wg.Add(waitFor)
	var lastErr error
	for cName := range ps.clients {
		go func(cName string) {
			def, err := ps.resolveForRuntime(ref.ID, cName)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"err":     err,
					"runtime": cName,
					"fn":      targetFn,
				}).Info("Function not found.")
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
		return types.FnRef{}, fmt.Errorf("failed to resolve function '%s' using clients '%v': %v",
			targetFn, ps.clients, lastErr)
	}
}

func (ps *MetaResolver) resolveForRuntime(targetFn string, runtime string) (types.FnRef, error) {
	dst, ok := ps.clients[runtime]
	if !ok {
		return types.FnRef{}, ErrInvalidRuntime
	}
	rsv, err := dst.Resolve(targetFn)
	if err != nil {
		return types.FnRef{}, err
	}
	return types.FnRef{
		Runtime: runtime,
		ID:      rsv,
	}, nil
}

//
// Helper functions
//

// ResolveTask resolved all the tasks in the provided workflow spec.
//
// In case there are no functions resolved, an empty slice is returned.
func ResolveTasks(ps Resolver, tasks map[string]*types.TaskSpec) (map[string]*types.FnRef, error) {
	var flattened []*types.TaskSpec
	for _, v := range tasks {
		flattened = append(flattened, v)
	}
	return ResolveTask(ps, flattened...)
}

// ResolveTask resolved the interpreted workflow from a given spec.
//
// It returns a map consisting of the original functionRef as the key.
func ResolveTask(ps Resolver, tasks ...*types.TaskSpec) (map[string]*types.FnRef, error) {
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
	resolved := map[string]*types.FnRef{}
	resolvedC := make(chan sourceFnRef, len(uniqueTasks))

	// ResolveTask each task in the workflow definition in parallel
	for k, t := range uniqueTasks {
		go func(k string, t *types.TaskSpec, tc chan sourceFnRef) {
			err := resolveTask(ps, k, t, tc)
			if err != nil {
				lastErr = err
			}
			wg.Done()
		}(k, t, resolvedC)
	}

	// Close channel when all tasks have resolved
	go func() {
		wg.Wait()
		close(resolvedC)
	}()

	// Store results of the resolved tasks
	for t := range resolvedC {
		resolved[t.src] = t.FnRef
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return resolved, nil
}

// resolveTaskAndInputs traverses the inputs of a task to resolve nested functions and workflows.
func resolveTask(ps Resolver, id string, task *types.TaskSpec, resolvedC chan sourceFnRef) error {
	if task == nil || resolvedC == nil {
		return nil
	}
	t, err := ps.Resolve(task.FunctionRef)
	if err != nil {
		return err
	}
	resolvedC <- sourceFnRef{
		src:   id,
		FnRef: &t,
	}

	return nil
}

// sourceFnRef wraps the FnRef with the source field which contains the original function target.
type sourceFnRef struct {
	*types.FnRef
	src string
}
