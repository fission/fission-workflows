package workflows

import (
	"context"
	"errors"
	"time"

	"github.com/fission/fission-workflows/pkg/api/invocation"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/sirupsen/logrus"
)

const (
	invokeSyncTimeout         = time.Duration(10) * time.Minute
	invokeSyncPollingInterval = time.Duration(100) * time.Millisecond
	Name                      = "workflows"
)

// Runtime provides an abstraction of the workflow engine itself to use as a Function runtime environment.
type Runtime struct {
	api      *invocation.Api
	wfiCache fes.CacheReader
}

func NewRuntime(api *invocation.Api, wfiCache fes.CacheReader) *Runtime {
	return &Runtime{
		api:      api,
		wfiCache: wfiCache,
	}
}

func (rt *Runtime) Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocationStatus, error) {
	wfiId, err := rt.api.Invoke(spec.ToWorkflowSpec())
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
			logrus.Warnf("Failed to get workflow invocation from cache: %v", err)
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

func CreateFnRef(wfId string) types.FnRef {
	return types.FnRef{
		Runtime: Name,
		ID:      wfId,
	}
}
