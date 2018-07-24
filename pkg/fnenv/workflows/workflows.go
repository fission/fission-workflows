// package workflows exposes the workflow engine itself as a function environment to improve recursion.
//
// Although other runtimes could also be used to proxy calls to the workflow engine, that would be an unnecessarily
// expensive operation. The call would pass from the workflow engine to the function environment back to the workflow
// engine. Shortcutting this round-trip by avoiding leaving the workflow engine, reduces the critical path of the call.
//
// Besides the performance, recursive workflow calls happen in the context of a higher-level workflow.
// To avoid confusing users and cluttering external (logging) systems, this package enables these workflows to remain
// largely opaque to the user.
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
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/fission/fission-workflows/pkg/util/labels"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/sirupsen/logrus"
)

const (
	Timeout      = time.Duration(10) * time.Minute
	PollInterval = time.Duration(100) * time.Millisecond
	Name         = "workflows"
)

// Runtime provides an abstraction of the workflow engine itself to use as a Task runtime environment.
type Runtime struct {
	api          *api.Invocation
	wfiCache     fes.CacheReader
	timeout      time.Duration
	pollInterval time.Duration
}

func NewRuntime(api *api.Invocation, wfiCache fes.CacheReader) *Runtime {
	return &Runtime{
		api:          api,
		wfiCache:     wfiCache,
		pollInterval: PollInterval,
		timeout:      Timeout,
	}
}

func (rt *Runtime) Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocationStatus, error) {
	if err := validate.TaskInvocationSpec(spec); err != nil {
		return nil, err
	}

	wfSpec, err := toWorkflowSpec(spec)
	if err != nil {
		return nil, err
	}

	// Note: currently context is not supported in the runtime interface, so we use a background context.
	wfi, err := rt.InvokeWorkflow(context.Background(), wfSpec)
	if err != nil {
		return nil, err
	}
	return wfi.Status.ToTaskStatus(), nil
}

func (rt *Runtime) InvokeWorkflow(ctx context.Context, spec *types.WorkflowInvocationSpec) (*types.WorkflowInvocation, error) {
	if err := validate.WorkflowInvocationSpec(spec); err != nil {
		return nil, err
	}

	timeStart := time.Now()
	fnenv.FnActive.WithLabelValues(Name).Inc()
	defer fnenv.FnExecTime.WithLabelValues(Name).Observe(float64(time.Since(timeStart)))
	defer fnenv.FnActive.WithLabelValues(Name).Dec()
	defer fnenv.FnCount.WithLabelValues(Name).Inc()

	logrus.Warn("INVOKE", spec)
	wfiID, err := rt.api.Invoke(spec)
	if err != nil {
		logrus.WithField("fnenv", Name).Errorf("Failed to invoke workflow: %v", err)
		return nil, err
	}
	logrus.WithField("fnenv", Name).Infof("Invoked workflow: %s", wfiID)

	timedCtx, cancelFn := context.WithTimeout(ctx, rt.timeout)
	defer cancelFn()
	if pub, ok := rt.wfiCache.(pubsub.Publisher); ok {
		sub := pub.Subscribe(pubsub.SubscriptionOptions{
			Buffer: 1,
			LabelMatcher: labels.And(
				labels.In(fes.PubSubLabelAggregateType, aggregates.TypeWorkflowInvocation),
				labels.In(fes.PubSubLabelAggregateID, wfiID),
				labels.In(fes.PubSubLabelEventType, events.Invocation_INVOCATION_COMPLETED.String(),
					events.Invocation_INVOCATION_CANCELED.String(), events.Invocation_INVOCATION_FAILED.String())),
		})
		defer pub.Unsubscribe(sub)

		// Check the cache once to ensure that we did not miss the complete event
		if result := rt.checkForResult(wfiID); result != nil {
			return result, nil
		}

		// Block until either we received an completion event or the context completed
		select {
		case <-sub.Ch:
		case <-timedCtx.Done():
		}
		return rt.checkForResult(wfiID), timedCtx.Err()
	}

	// Fallback to polling the cache if the cache does not support pubsub.
	return rt.pollUntilResult(timedCtx, wfiID)
}

// checkForResult checks if the invocation with the specified ID has completed yet.
// If so it will return the workflow invocation object, otherwise it will return nil.
func (rt *Runtime) checkForResult(wfiID string) *types.WorkflowInvocation {
	wi := aggregates.NewWorkflowInvocation(wfiID)
	err := rt.wfiCache.Get(wi)
	if err != nil {
		logrus.Debugf("Could not find workflow invocation in cache: %v", err)
	}
	if wi != nil && wi.GetStatus() != nil && wi.GetStatus().Finished() {
		return wi.WorkflowInvocation
	}

	return nil
}

// pollUntilResult continuously (or until the context is canceled) polls whether the workflow invocation with the
// specified ID has finished. It either returns the invocation object (if completed) or an error in case of timeouts or
// context cancellation.
func (rt *Runtime) pollUntilResult(ctx context.Context, wfiID string) (*types.WorkflowInvocation, error) {
	for {
		if result := rt.checkForResult(wfiID); result != nil {
			return result, nil
		}

		select {
		case <-ctx.Done():
			err := rt.api.Cancel(wfiID)
			if err != nil {
				logrus.Errorf("Failed to cancel workflow invocation: %v", err)
			}
			return nil, ctx.Err()
		default:
			time.Sleep(rt.pollInterval)
		}
	}
}

func toWorkflowSpec(spec *types.TaskInvocationSpec) (*types.WorkflowInvocationSpec, error) {

	// Prepare inputs
	wfSpec := spec.ToWorkflowSpec()
	if parentTv, ok := spec.Inputs[types.InputParent]; ok {
		parentID, err := typedvalues.FormatString(parentTv)
		if err != nil {
			return nil, fmt.Errorf("invalid parent id %v (%v)", parentTv, err)
		}
		wfSpec.ParentId = parentID
	}
	return wfSpec, nil
}
