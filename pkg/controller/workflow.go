package controller

import (
	"reflect"

	"time"

	"context"

	"fmt"
	"github.com/fission/fission-workflows/pkg/api/workflow"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/util/backoff"
	"github.com/fission/fission-workflows/pkg/util/labels"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
)

var wfLog = log.WithField("controller", "controller-wf")

// WorkflowController is the controller concerned with the lifecycle of workflows. It handles responsibilities, such as
// parsing of workflows.
type WorkflowController struct {
	wfCache        fes.CacheReader
	api            *workflow.Api
	workQueue      chan Action
	backoffHandler backoff.Map // Specifies the time after which the wf is allowed to be reevaluated again
	sub            *pubsub.Subscription
	cancelFn       context.CancelFunc
}

func NewWorkflowController(wfCache fes.CacheReader, wfApi *workflow.Api) *WorkflowController {
	return &WorkflowController{
		wfCache:        wfCache,
		api:            wfApi,
		workQueue:      make(chan Action, WorkQueueSize),
		backoffHandler: backoff.NewMap(),
	}
}

func (ctr *WorkflowController) Init(sctx context.Context) error {
	ctx, cancelFn := context.WithCancel(sctx)
	ctr.cancelFn = cancelFn

	// Subscribe to invocation creations and task events.
	selector := labels.InSelector("aggregate.type", "workflow")
	if invokePub, ok := ctr.wfCache.(pubsub.Publisher); ok {
		ctr.sub = invokePub.Subscribe(pubsub.SubscriptionOptions{
			Buf:           NotificationBuffer,
			LabelSelector: selector,
		})
	}

	// Workflow Notification lane
	go func(ctx context.Context) {
		for {
			select {
			case notification := <-ctr.sub.Ch:
				wfLog.WithField("labels", notification.Labels()).Info("Handling workflow notification.")
				switch n := notification.(type) {
				case *fes.Notification:
					ctr.HandleNotification(n)
				default:
					wfLog.WithField("notification", n).Warn("Ignoring unknown notification type")
				}
			case <-ctx.Done():
				wfLog.WithField("ctx.err", ctx.Err()).Debug("Notification listener closed.")
				return
			}
		}
	}(ctx)

	// Action WorkQueue Handler
	go func(ctx context.Context) {
		for {
			select {
			case action := <-ctr.workQueue:
				ctr.HandleAction(action)
			case <-ctx.Done():
				wfLog.WithField("ctx.err", ctx.Err()).Info("WorkQueue handler closed.")
				return
			}
		}
	}(ctx)

	return nil
}

func (ctr *WorkflowController) HandleTick() error {
	// Assume that all workflows are in cache
	now := time.Now()
	for _, a := range ctr.wfCache.List() {
		locked := ctr.backoffHandler.Locked(a.Id, now)
		if locked {
			continue
		}

		wfEntity, err := ctr.wfCache.GetAggregate(a)
		if err != nil {
			return fmt.Errorf("failed to retrieve: %v", err)
		}

		wf, ok := wfEntity.(*aggregates.Workflow)
		if !ok {
			wfLog.WithField("wfEntity", wfEntity).WithField("type", reflect.TypeOf(wfEntity)).
				Error("Unexpected type in wfCache")
			panic(fmt.Sprintf("unexpected type '%v' in wfCache", reflect.TypeOf(wfEntity)))
		}

		err = ctr.evaluate(wf.Workflow)
		if err != nil {
			return fmt.Errorf("failed to evaluate workflow: %v", err)
		}
	}
	return nil
}

func (ctr *WorkflowController) HandleNotification(msg *fes.Notification) error {
	wf, ok := msg.Payload.(*aggregates.Workflow)
	if !ok {
		//wfLog.WithField("type", reflect.TypeOf(msg.Payload)).Error("Received notification of invalid type")
		return fmt.Errorf("received notification of invalid type '%s'. Expected '*aggregates.Workflow'", reflect.TypeOf(msg.Payload))
	}

	// Let's ignore back-off state, because something might have altered the state of the workflow
	// Check if the target workflow is not in a back-off state
	//locked := ctr.backoffHandler.Locked(wf.Aggregate().Id, time.Now())
	//if locked {
	//	return errors.New("not handling notification; wf is in back-off state")
	//}

	return ctr.evaluate(wf.Workflow)
}

func (ctr *WorkflowController) evaluate(wf *types.Workflow) error {
	switch wf.Status.Status {
	case types.WorkflowStatus_READY:
		// Alright.
	case types.WorkflowStatus_DELETED:
		wfLog.WithField("wf", wf.Metadata.Id).Warn("Should be removed")
	case types.WorkflowStatus_UNKNOWN:
		fallthrough
	case types.WorkflowStatus_FAILED:
		ok := ctr.submit(&parseWorkflowAction{
			wfApi: ctr.api,
			wf:    wf,
		})
		if !ok {
			wfLog.Warn("Failed to submit action.")
		}
	}
	return nil
}

func (ctr *WorkflowController) submit(action Action) (submitted bool) {
	select {
	case ctr.workQueue <- action:
		// Ok
		submitted = true
	default:
		// Action overflow
	}
	return submitted
}

func (ctr *WorkflowController) HandleAction(action Action) {
	actionLog := wfLog.WithField("action", action.Id())
	actionLog.Info("Handling action.")
	err := action.Apply()
	if err != nil {
		actionLog.Errorf("Failed to perform action: %v", err)
		bkf := ctr.backoffHandler.Backoff(action.Id())
		actionLog.Infof("Set evaluation backoff to %d ms (attempt: %v)",
			bkf.Lockout.Nanoseconds()/1000, bkf.Attempts)
	}
}

func (ctr *WorkflowController) Close() error {
	wfLog.Info("Closing controller...")
	if invokePub, ok := ctr.wfCache.(pubsub.Publisher); ok {
		err := invokePub.Unsubscribe(ctr.sub)
		if err != nil {
			return err
		}
	}

	ctr.cancelFn()
	return nil
}

//
// Actions
//

type parseWorkflowAction struct {
	wfApi *workflow.Api
	wf    *types.Workflow
}

func (ac *parseWorkflowAction) Id() string {
	return ac.wf.Metadata.Id
}

func (ac *parseWorkflowAction) Apply() error {
	_, err := ac.wfApi.Parse(ac.wf)
	return err
}
