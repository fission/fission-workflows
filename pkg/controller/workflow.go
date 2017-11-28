package controller

import (
	"reflect"

	"time"

	"context"

	"github.com/fission/fission-workflows/pkg/api/workflow"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/util/backoff"
	"github.com/fission/fission-workflows/pkg/util/labels"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/sirupsen/logrus"
)

// Part of the controller concerned with workflows

type WorkflowController struct {
	wfCache   fes.CacheReader
	api       *workflow.Api
	workQueue chan Action
	wfBackoff backoff.Map // Specifies the time after which the wf is allowed to be reevaluated again

	wfSub    *pubsub.Subscription
	cancelFn context.CancelFunc
}

func NewWorkflowController(wfCache fes.CacheReader, wfApi *workflow.Api) *WorkflowController {
	return &WorkflowController{
		wfCache:   wfCache,
		api:       wfApi,
		workQueue: make(chan Action, 50),
		wfBackoff: backoff.NewMap(),
	}
}

func (ctr *WorkflowController) Init() error {
	ctx, cancelFn := context.WithCancel(context.Background())
	ctr.cancelFn = cancelFn

	// Subscribe to invocation creations and task events.
	selector := labels.InSelector("aggregate.type", "workflow")

	if invokePub, ok := ctr.wfCache.(pubsub.Publisher); ok {
		ctr.wfSub = invokePub.Subscribe(pubsub.SubscriptionOptions{
			Buf:           NOTIFICATION_BUFFER,
			LabelSelector: selector,
		})
	}

	// Workflow Notification lane
	go func(ctx context.Context) {
		for {
			select {
			case notification := <-ctr.wfSub.Ch:
				logrus.WithField("labels", notification.Labels()).Info("Handling workflow notification.")
				switch n := notification.(type) {
				case *fes.Notification:
					ctr.HandleNotification(n)
				default:
					logrus.WithField("notification", n).Warn("Ignoring unknown notification type")
				}
			case <-ctx.Done():
				logrus.WithField("ctx.err", ctx.Err()).Debug("Notification listener closed.")
				return
			}
		}
	}(ctx)

	// Action workQueue
	go func(ctx context.Context) {
		for {
			select {
			case action := <-ctr.workQueue:
				ctr.HandleAction(action)
			case <-ctx.Done():
				logrus.WithField("ctx.err", ctx.Err()).Debug("Action workQueue worker closed.")
				return
			}
		}
	}(ctx)

	return nil
}

func (ctr *WorkflowController) HandleTick() {
	// Assume that all workflows are in cache
	now := time.Now()
	for _, a := range ctr.wfCache.List() {
		locked := ctr.wfBackoff.Locked(a.Id, now)
		if locked {
			continue
		}

		wfEntity, err := ctr.wfCache.GetAggregate(a)
		if err != nil {
			logrus.Errorf("Failed to retrieve: %v", err)
			continue
		}

		wf, ok := wfEntity.(*aggregates.Workflow)
		if !ok {
			logrus.WithField("wfEntity", wfEntity).WithField("type", reflect.TypeOf(wfEntity)).
				Error("Unexpected type in wfCache")
			continue
		}

		err = ctr.evaluate(wf.Workflow)
		if err != nil {
			logrus.Errorf("Failed to evaluate: %v", err)
			continue
		}
	}
}

func (ctr *WorkflowController) HandleNotification(msg *fes.Notification) {
	wf, ok := msg.Payload.(*aggregates.Workflow)
	if !ok {
		logrus.WithField("type", reflect.TypeOf(msg.Payload)).Error("Received notification of invalid type")
		return
	}

	// Check if the target workflow is not in a backoff
	//locked := ctr.wfBackoff.Locked(wf.Aggregate().Id, time.Now())
	//if locked {
	//	return
	//}

	err := ctr.evaluate(wf.Workflow)
	if err != nil {
		logrus.WithField("wf", wf.Metadata.Id).WithField("err", err).
			Error("Failed to evaluate workflow")
		return
	}
}

func (ctr *WorkflowController) evaluate(wf *types.Workflow) error {
	switch wf.Status.Status {
	case types.WorkflowStatus_READY:
		// Alright.
	case types.WorkflowStatus_DELETED:
		logrus.WithField("wf", wf.Metadata.Id).Warn("Should be removed")
	case types.WorkflowStatus_UNKNOWN:
		fallthrough
	case types.WorkflowStatus_FAILED:
		ctr.submit(&parseWorkflowAction{
			wfApi: ctr.api,
			wf:    wf,
		})
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
	actionLog := logrus.WithField("action", action.Id())
	actionLog.Info("Handling action.")
	err := action.Apply()
	if err != nil {
		actionLog.Errorf("Failed to perform action: %v", err)
		bkf := ctr.wfBackoff.Backoff(action.Id())
		actionLog.Infof("Set evaluation backoff to %d ms (attempt: %v)",
			bkf.Lockout.Nanoseconds()/1000, bkf.Attempts)
	}
}

func (ctr *WorkflowController) Close() error {
	logrus.Debug("Closing controller...")
	if invokePub, ok := ctr.wfCache.(pubsub.Publisher); ok {
		err := invokePub.Unsubscribe(ctr.wfSub)
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
