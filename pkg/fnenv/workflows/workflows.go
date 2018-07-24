package workflows

import (
	"context"
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fnenv"
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
		parentID, err := typedvalues.FormatString(parentTv)
		if err != nil {
			return nil, fmt.Errorf("invalid parent id %v (%v)", parentTv, err)
		}
		wfSpec.ParentId = parentID
	}

	// Note: currently context is not supported in the runtime interface, so we use a background context.
	wfi, err := rt.InvokeWorkflow(context.Background(), wfSpec)
	if err != nil {
		return nil, err
	}
	return wfi.Status.ToTaskStatus(), nil
}

func (rt *Runtime) InvokeWorkflow(ctx context.Context, spec *types.WorkflowInvocationSpec) (*types.WorkflowInvocation, error) {
	timeStart := time.Now()
	defer fnenv.FnExecTime.WithLabelValues(Name).Observe(float64(time.Since(timeStart)))
	fnenv.FnActive.WithLabelValues(Name).Inc()
	wfiID, err := rt.api.Invoke(spec)
	if err != nil {
		logrus.Errorf("Failed to invoke workflow: %v", err)
		return nil, err
	}

	timedCtx, cancelFn := context.WithTimeout(ctx, invokeSyncTimeout)
	defer cancelFn()
	var result *types.WorkflowInvocation
	for {
		wi := aggregates.NewWorkflowInvocation(wfiID)
		err := rt.wfiCache.Get(wi)
		if err != nil {
			logrus.Debugf("Could not find workflow invocation in cache: %v", err)
		}
		if wi != nil && wi.GetStatus() != nil && wi.GetStatus().Finished() {
			result = wi.WorkflowInvocation
			break
		}

		select {
		case <-timedCtx.Done():
			err := rt.api.Cancel(wfiID)
			if err != nil {
				logrus.Errorf("Failed to cancel workflow invocation: %v", err)
			}
			return nil, timedCtx.Err()
		default:
			// TODO polling is a temporary shortcut; needs optimizing.
			time.Sleep(invokeSyncPollingInterval)
		}
	}
	fnenv.FnActive.WithLabelValues(Name).Dec()
	fnenv.FnCount.WithLabelValues(Name).Inc()

	return result, nil
}
