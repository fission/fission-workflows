package workflows

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/sirupsen/logrus"
)

const (
	invokeSyncTimeout         = time.Duration(10) * time.Minute
	invokeSyncPollingInterval = time.Duration(100) * time.Millisecond
	Name                      = "workflows"
)

// Runtime provides an abstraction of the workflow engine itself to use as a Task runtime environment.
type Runtime struct {
	api      *api.Invocation
	wfiCache fes.CacheReader
}

func NewRuntime(api *api.Invocation, wfiCache fes.CacheReader) *Runtime {
	return &Runtime{
		api:      api,
		wfiCache: wfiCache,
	}
}

// TODO support async
func (rt *Runtime) Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocationStatus, error) {
	if err := validate.TaskInvocationSpec(spec); err != nil {
		return nil, err
	}

	// Prepare inputs
	wfSpec := spec.ToWorkflowSpec()
	if parentTv, ok := spec.Inputs[types.InputParent]; ok {
		parentId, err := typedvalues.FormatString(parentTv)
		if err != nil {
			return nil, fmt.Errorf("invalid parent id %v (%v)", parentTv, err)
		}
		wfSpec.ParentId = parentId
	}

	// Invoke workflow
	wfiId, err := rt.api.Invoke(wfSpec)
	if err != nil {
		logrus.Errorf("Failed to invoke workflow: %v", err)
		return nil, err
	}

	timeout, cancelFn := context.WithTimeout(context.TODO(), invokeSyncTimeout)
	defer cancelFn()
	var result *types.WorkflowInvocation
	for {
		wi := aggregates.NewWorkflowInvocation(wfiId)
		err := rt.wfiCache.Get(wi)
		if err != nil {
			logrus.Debugf("Could not find workflow invocation in cache: %v", err)
		}
		if wi != nil && wi.GetStatus() != nil && wi.GetStatus().Finished() {
			result = wi.WorkflowInvocation
			break
		}

		select {
		case <-timeout.Done():
			err := rt.api.Cancel(wfiId)
			if err != nil {
				logrus.Errorf("Failed to cancel workflow invocation: %v", err)
			}
			return nil, errors.New("timeout occurred")
		default:
			// TODO polling is a temporary shortcut; needs optimizing.
			time.Sleep(invokeSyncPollingInterval)
		}
	}

	return result.Status.ToTaskStatus(), nil

}
