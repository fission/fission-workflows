package controller

import (
	"context"

	"github.com/fission/fission-workflows/pkg/api/function"
	"github.com/fission/fission-workflows/pkg/api/invocation"
	"github.com/fission/fission-workflows/pkg/controller/actions"
	"github.com/fission/fission-workflows/pkg/controller/query"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/scheduler"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/fission/fission-workflows/pkg/util/labels"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

const (
	NOTIFICATION_BUFFER = 100
)

type InvocationController struct {
	invokeCache   fes.CacheReader
	wfCache       fes.CacheReader
	functionApi   *function.Api
	invocationApi *invocation.Api
	scheduler     *scheduler.WorkflowScheduler
	invocSub      *pubsub.Subscription
	exprParser    query.ExpressionParser
	cancelFn      context.CancelFunc
	// TODO add active cache
}

func NewInvocationController(invokeCache fes.CacheReader, wfCache fes.CacheReader,
	workflowScheduler *scheduler.WorkflowScheduler, functionApi *function.Api, invocationApi *invocation.Api,
	exprParser query.ExpressionParser) *InvocationController {
	return &InvocationController{
		invokeCache:   invokeCache,
		wfCache:       wfCache,
		scheduler:     workflowScheduler,
		functionApi:   functionApi,
		invocationApi: invocationApi,
		exprParser:    exprParser,
	}
}

func (cr *InvocationController) Init() error {
	ctx, cancelFn := context.WithCancel(context.Background())
	cr.cancelFn = cancelFn

	// Subscribe to invocation creations and task events.
	selector := labels.InSelector("aggregate.type", "invocation", "function")

	if invokePub, ok := cr.invokeCache.(pubsub.Publisher); ok {
		cr.invocSub = invokePub.Subscribe(pubsub.SubscriptionOptions{
			Buf:           NOTIFICATION_BUFFER,
			LabelSelector: selector,
		})
	}

	// Invocation Notification lane
	go func(ctx context.Context) {
		for {
			select {
			case notification := <-cr.invocSub.Ch:
				logrus.WithField("labels", notification.Labels()).Info("Handling invocation notification.")
				switch n := notification.(type) {
				case *fes.Notification:
					cr.HandleNotification(n)
				default:
					logrus.WithField("notification", n).Warn("Ignoring unknown notification type")
				}
			case <-ctx.Done():
				logrus.WithField("ctx.err", ctx.Err()).Debug("Notification listener closed.")
				return
			}
		}
	}(ctx)
	return nil
}

func (cr *InvocationController) HandleNotification(msg *fes.Notification) {
	logrus.WithFields(logrus.Fields{
		"notification": msg.EventType,
		"labels":       msg.Labels(),
	}).Info("controller event trigger!")

	switch msg.EventType {
	case events.Invocation_INVOCATION_CREATED.String():
		fallthrough
	case events.Function_TASK_SUCCEEDED.String():
		fallthrough
	case events.Function_TASK_FAILED.String():
		// Decide which task to execute next
		invoc, ok := msg.Payload.(*aggregates.WorkflowInvocation)
		if !ok {
			panic(msg)
		}

		cr.evaluate(invoc.WorkflowInvocation)
	default:
		logrus.WithField("type", msg.EventType).Warn("Controller ignores unknown event.")
	}
}

func (cr *InvocationController) HandleTick() {
	logrus.Debug("Controller tick...")
	// Options: refresh projection, send ping, cancel invocation
	// Short loop (invocations the controller is actively tracking) and long loop (to check if there are any orphans)

	// Long loop
	entities := cr.invokeCache.List()
	for _, entity := range entities {
		wi := aggregates.NewWorkflowInvocation(entity.Id, nil)
		err := cr.invokeCache.Get(wi)
		if err != nil {
			logrus.Error(err)
			return
		}

		cr.evaluate(wi.WorkflowInvocation)
	}
}

func (cr *InvocationController) evaluate(invoc *types.WorkflowInvocation) {
	wf := aggregates.NewWorkflow(invoc.Spec.WorkflowId, nil)
	err := cr.wfCache.Get(wf)
	if err != nil {
		logrus.Errorf("Failed to get workflow for invocation '%s': %v", invoc.Spec.WorkflowId, err)
		return
	}

	if wf.Status.Status != types.WorkflowStatus_READY {
		logrus.WithField("wf.status", wf.Status.Status).Error("Workflow has not been parsed yet.")
		return
	}

	schedule, err := cr.scheduler.Evaluate(&scheduler.ScheduleRequest{
		Invocation: invoc,
		Workflow:   wf.Workflow,
	})

	if len(schedule.Actions) == 0 { // TODO controller should verify (it is an invariant)
		var output *types.TypedValue
		if t, ok := invoc.Status.Tasks[wf.Spec.OutputTask]; ok {
			output = t.Status.Output
		} else {
			logrus.Infof("Output task '%v' does not exist", wf.Spec.OutputTask)
			return
		}
		err := cr.invocationApi.MarkCompleted(invoc.Metadata.Id, output)
		if err != nil {
			panic(err)
		}
	}

	for _, action := range schedule.Actions {
		switch action.Type {
		case scheduler.ActionType_ABORT:
			err := actions.Abort(invoc.Metadata.Id, cr.invocationApi)
			if err != nil {
				logrus.Error(err)
			}
		case scheduler.ActionType_INVOKE_TASK:
			invokeAction := &scheduler.InvokeTaskAction{}
			ptypes.UnmarshalAny(action.Payload, invokeAction)
			err := actions.InvokeTask(invokeAction, wf.Workflow, invoc, cr.exprParser, cr.functionApi)
			if err != nil {
				logrus.Error(err)
			}
		}
	}
}

func (cr *InvocationController) Close() error {
	logrus.Debug("Closing controller...")
	if invokePub, ok := cr.invokeCache.(pubsub.Publisher); ok {
		err := invokePub.Unsubscribe(cr.invocSub)
		if err != nil {
			return err
		}
	}

	cr.cancelFn()

	return nil
}
