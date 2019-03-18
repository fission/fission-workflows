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
	"errors"
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/fission/fission-workflows/pkg/api/store"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/fission/fission-workflows/pkg/util/labels"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/golang/protobuf/ptypes"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

const (
	DefaultMaxRuntime = time.Duration(10) * time.Minute
	PollInterval      = time.Duration(100) * time.Millisecond
	Name              = "workflows"
)

// Runtime provides an abstraction of the workflow engine itself to use as a Task runtime environment.
type Runtime struct {
	api          *api.Invocation
	invocations  *store.Invocations
	workflows    *store.Workflows
	timeout      time.Duration
	pollInterval time.Duration
}

func NewRuntime(api *api.Invocation, invocations *store.Invocations, workflows *store.Workflows) *Runtime {
	return &Runtime{
		api:          api,
		invocations:  invocations,
		workflows:    workflows,
		pollInterval: PollInterval,
		timeout:      DefaultMaxRuntime,
	}
}

func (rt *Runtime) Invoke(spec *types.TaskInvocationSpec, opts ...fnenv.InvokeOption) (*types.TaskInvocationStatus, error) {
	if err := validate.TaskInvocationSpec(spec); err != nil {
		return nil, err
	}

	wfSpec, err := toWorkflowSpec(spec)
	if err != nil {
		return nil, err
	}

	// Note: currently context is not supported in the runtime interface, so we use a background context.
	wfi, err := rt.InvokeWorkflow(wfSpec, opts...)
	if err != nil {
		return nil, err
	}
	return wfi.Status.ToTaskStatus(), nil
}

func (rt *Runtime) InvokeWorkflow(spec *types.WorkflowInvocationSpec, opts ...fnenv.InvokeOption) (*types.WorkflowInvocation, error) {
	cfg := fnenv.ParseInvokeOptions(opts)
	if err := validate.WorkflowInvocationSpec(spec); err != nil {
		return nil, err
	}
	ctx := cfg.Ctx

	span, ctx := opentracing.StartSpanFromContext(cfg.Ctx, "/fnenv/workflows")
	defer span.Finish()
	span.SetTag("workflow", spec.GetWorkflowId())
	span.SetTag("parent", spec.GetParentId())
	span.SetTag("internal", len(spec.GetParentId()) != 0)

	// Check if the workflow required by the invocation exists
	if spec.Workflow == nil {
		awaitWorkflowCtx, cancel := context.WithTimeout(ctx, cfg.AwaitWorkflow)
		wf, err := rt.awaitReadyWorkflow(awaitWorkflowCtx, spec.GetWorkflowId())
		cancel()
		if err != nil {
			span.LogKV("error", err)
			return nil, err
		}
		spec.Workflow = wf
	} else {
		if !spec.Workflow.GetStatus().Ready() {
			err := errors.New("provided workflow is not ready")
			span.LogKV("error", err)
			return nil, err
		}
	}

	span.SetTag("workflow.name", spec.GetWorkflow().GetMetadata().GetName())

	// If debugging mode is enabled, add all inputs to the trace.
	if logrus.GetLevel() == logrus.DebugLevel {
		var inputs interface{}
		var err error
		inputs, err = typedvalues.UnwrapMapTypedValue(spec.GetInputs())
		if err != nil {
			inputs = fmt.Errorf("error: %v", err)
		}
		span.LogKV("inputs", inputs)
	}

	timeStart := time.Now()
	fnenv.FnActive.WithLabelValues(Name).Inc()
	defer fnenv.FnExecTime.WithLabelValues(Name).Observe(float64(time.Since(timeStart)))
	defer fnenv.FnActive.WithLabelValues(Name).Dec()
	defer fnenv.FnCount.WithLabelValues(Name).Inc()

	invocationID, err := rt.api.Invoke(spec, api.WithContext(ctx))
	if err != nil {
		logrus.WithField("fnenv", Name).Errorf("Failed to invoke workflow: %v", err)
		span.LogKV("error", fmt.Errorf("failed to invoke workflow: %v", err))
		return nil, err
	}
	logrus.WithField("fnenv", Name).Infof("Invoked workflow: %s", invocationID)
	span.SetTag("invocation", invocationID)

	// If no deadline was set by the user, use the default deadline.
	var maxRuntime time.Duration
	if spec.Deadline == nil {
		if rt.timeout > 0 {
			maxRuntime = rt.timeout
		} else {
			maxRuntime = DefaultMaxRuntime
		}
		ts, _ := ptypes.TimestampProto(time.Now().Add(maxRuntime))
		spec.Deadline = ts
		rt.timeout = maxRuntime
	}

	// Subscribe and poll for the result
	awaitInvocationCtx, cancel := context.WithTimeout(ctx, maxRuntime)
	invocation, err := rt.awaitInvocationResult(awaitInvocationCtx, invocationID)
	cancel()
	if err != nil {
		span.LogKV("error", err)
		return nil, err
	}
	return invocation, nil
}

// checkForInvocationResult checks if the invocation with the specified ID has completed yet.
// If so it will return the workflow invocation object, otherwise it will return nil.
func (rt *Runtime) checkForInvocationResult(wfiID string) *types.WorkflowInvocation {
	wi, err := rt.invocations.GetInvocation(wfiID)
	if err != nil {
		logrus.Debugf("Could not find workflow invocation in cache: %v", err)
	}
	if wi != nil && wi.GetStatus() != nil && wi.GetStatus().Finished() {
		return wi
	}
	return nil
}

func (rt *Runtime) checkForReadyWorkflow(workflowID string) (*types.Workflow, error) {
	wf, err := rt.workflows.GetWorkflow(workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to find workflow %v for new invocation", workflowID)
	}
	if !wf.GetStatus().Ready() {
		return nil, fmt.Errorf("cannot invoke non-ready workflow %v (status: %v)", workflowID,
			wf.GetStatus().GetStatus().String())
	}
	return wf, nil
}

func (rt *Runtime) awaitReadyWorkflow(ctx context.Context, workflowID string) (wf *types.Workflow, err error) {
	if wf, err = rt.checkForReadyWorkflow(workflowID); err == nil && wf != nil {
		return wf, nil
	}

	// await the parsing of the workflow
	if pub, ok := rt.invocations.CacheReader.(pubsub.Publisher); ok {
		sub := pub.Subscribe(pubsub.SubscriptionOptions{
			Buffer: 1,
			LabelMatcher: labels.And(
				labels.In(fes.PubSubLabelAggregateType, types.TypeWorkflow),
				labels.In(fes.PubSubLabelAggregateID, workflowID),
				labels.In(fes.PubSubLabelEventType,
					append(events.WorkflowTerminalEvents, events.EventWorkflowParsed)...)),
		})
		defer pub.Unsubscribe(sub)

		// Check the cache once to ensure that we did not miss the terminal event while subscribing
		if wf, err = rt.checkForReadyWorkflow(workflowID); err == nil {
			return wf, nil
		}

		select {
		case <-ctx.Done():
			// Check once before cancelling, whether cancelling is needed.
			if result, err := rt.checkForReadyWorkflow(workflowID); result != nil {
				return result, nil
			} else {
				return nil, err
			}
		case <-sub.Ch:
			return rt.checkForReadyWorkflow(workflowID)
		}
	}
	return rt.pollUntilWorkflowResult(ctx, workflowID)
}

func (rt *Runtime) awaitInvocationResult(ctx context.Context, invocationID string) (invocation *types.WorkflowInvocation, err error) {
	if pub, ok := rt.invocations.CacheReader.(pubsub.Publisher); ok {
		sub := pub.Subscribe(pubsub.SubscriptionOptions{
			Buffer: 1,
			LabelMatcher: labels.And(
				labels.In(fes.PubSubLabelAggregateType, types.TypeInvocation),
				labels.In(fes.PubSubLabelAggregateID, invocationID),
				labels.In(fes.PubSubLabelEventType, events.InvocationTerminalEvents...)),
		})
		defer pub.Unsubscribe(sub)
		logrus.Debugf("Listening for termination event for invocation %s", invocationID)

		// Check the cache once to ensure that we did not miss the terminal event while subscribing
		if result := rt.checkForInvocationResult(invocationID); result != nil {
			return result, nil
		}

		// Block until either we received an completion event or the context completed
		select {
		case <-ctx.Done():
			// Check once before cancelling, whether cancelling is needed.
			if result := rt.checkForInvocationResult(invocationID); result != nil {
				return result, nil
			}

			// Cancel the invocation
			err := rt.api.Cancel(invocationID)
			if err == nil {
				err = errors.New(api.ErrInvocationCanceled)
			} else {
				logrus.Errorf("Failed to cancel invocation: %v", err)
			}
			//span.LogKV("error", err)
			return nil, err
		case <-sub.Ch:
			logrus.Debugf("Received terminal event for invocation %s", invocationID)
			return rt.checkForInvocationResult(invocationID), nil
		}
	}

	// Fallback to polling the cache if the cache does not support pubsub.
	logrus.Debug("Workflows store does not support pubsub, falling back to polling.")
	return rt.pollUntilInvocationResult(ctx, invocationID)
}

// pollUntilInvocationResult continuously (or until the context is canceled) polls whether the workflow invocation with the
// specified ID has finished. It either returns the invocation object (if completed) or an error in case of timeouts or
// context cancellation.
func (rt *Runtime) pollUntilInvocationResult(ctx context.Context, wfiID string) (*types.WorkflowInvocation, error) {
	for {
		if result := rt.checkForInvocationResult(wfiID); result != nil {
			return result, nil
		}

		select {
		case <-ctx.Done():
			err := rt.api.Cancel(wfiID)
			if err != nil {
				return nil, err
			}
			return nil, ctx.Err()
		default:
			time.Sleep(rt.pollInterval)
		}
	}
}

func (rt *Runtime) pollUntilWorkflowResult(ctx context.Context, workflowID string) (*types.Workflow, error) {
	for {
		if wf, err := rt.checkForReadyWorkflow(workflowID); err == nil {
			return wf, nil
		}

		select {
		case <-ctx.Done():
			err := rt.api.Cancel(workflowID)
			if err != nil {
				return nil, err
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
		parentID, err := typedvalues.UnwrapString(parentTv)
		if err != nil {
			return nil, fmt.Errorf("invalid parent id %v (%v)", parentTv, err)
		}
		wfSpec.ParentId = parentID
	}
	return wfSpec, nil
}
