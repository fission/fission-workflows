package controller

import (
	"context"

	"fmt"

	"github.com/fission/fission-workflows/pkg/api/function"
	"github.com/fission/fission-workflows/pkg/api/invocation"
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
	sub           *pubsub.Subscription
	exprParser    query.ExpressionParser
	cancelFn      context.CancelFunc
	workQueue     chan Action
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
		workQueue:     make(chan Action, 50),
	}
}

func (cr *InvocationController) Init() error {
	ctx, cancelFn := context.WithCancel(context.Background())
	cr.cancelFn = cancelFn

	// Subscribe to invocation creations and task events.
	selector := labels.InSelector("aggregate.type", "invocation", "function")

	if invokePub, ok := cr.invokeCache.(pubsub.Publisher); ok {
		cr.sub = invokePub.Subscribe(pubsub.SubscriptionOptions{
			Buf:           NOTIFICATION_BUFFER,
			LabelSelector: selector,
		})
	}

	// Invocation Notification lane
	go func(ctx context.Context) {
		for {
			select {
			case notification := <-cr.sub.Ch:
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

	// workQueue loop
	go func(ctx context.Context) {
		for {
			select {
			case action := <-cr.workQueue:
				err := action.Apply()
				if err != nil {
					logrus.WithField("action", action).Error("WorkflowInvocation action failed")
				}
			case <-ctx.Done():
				logrus.WithField("ctx.err", ctx.Err()).Debug("WorkflowInvocation workQueue closed.")
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
			cr.submit(&abortAction{
				api:          cr.invocationApi,
				invocationId: invoc.Metadata.Id,
			})
		case scheduler.ActionType_INVOKE_TASK:
			invokeAction := &scheduler.InvokeTaskAction{}
			ptypes.UnmarshalAny(action.Payload, invokeAction)
			cr.submit(&invokeTaskAction{
				wf:   wf.Workflow,
				wfi:  invoc,
				expr: cr.exprParser,
				api:  cr.functionApi,
				task: invokeAction,
			})
		}
	}
}

func (cr *InvocationController) submit(action Action) (submitted bool) {
	select {
	case cr.workQueue <- action:
		// Ok
		submitted = true
	default:
		// Action overflow
	}
	return submitted
}

func (cr *InvocationController) Close() error {
	logrus.Debug("Closing controller...")
	if invokePub, ok := cr.invokeCache.(pubsub.Publisher); ok {
		err := invokePub.Unsubscribe(cr.sub)
		if err != nil {
			return err
		}
	}

	cr.cancelFn()

	return nil
}

//
// Actions
//

// abortAction aborts an invocation
type abortAction struct {
	api          *invocation.Api
	invocationId string
}

func (a *abortAction) Id() string {
	return a.invocationId
}

func (a *abortAction) Apply() error {
	logrus.Infof("aborting: '%v'", a.invocationId)
	return a.api.Cancel(a.invocationId)
}

// invokeTaskAction invokes a function
type invokeTaskAction struct {
	wf   *types.Workflow
	wfi  *types.WorkflowInvocation
	expr query.ExpressionParser
	api  *function.Api
	task *scheduler.InvokeTaskAction
}

func (a *invokeTaskAction) Id() string {
	return a.task.Id
}

func (a *invokeTaskAction) Apply() error {
	// Find task (static or dynamic)
	task, ok := a.wfi.Status.DynamicTasks[a.task.Id]
	if !ok {
		task, ok = a.wf.Spec.Tasks[a.task.Id]
		if !ok {
			return fmt.Errorf("unknown task '%v'", a.task.Id)
		}
	}
	logrus.Infof("Invoking function '%s' for task '%s'", task.FunctionRef, a.task.Id)

	// Resolve type of the task
	taskDef, ok := a.wf.Status.ResolvedTasks[task.FunctionRef]
	if !ok {
		return fmt.Errorf("no resolved task could be found for FunctionRef'%v'", task.FunctionRef)
	}

	// Resolve the inputs
	inputs := map[string]*types.TypedValue{}
	queryScope := query.NewScope(a.wf, a.wfi)
	for inputKey, val := range a.task.Inputs {
		resolvedInput, err := a.expr.Resolve(queryScope, queryScope.Tasks[a.task.Id], nil, val)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"val":      val,
				"inputKey": inputKey,
			}).Errorf("Failed to parse input: %v", err)
			return fmt.Errorf("failed to parse input '%v'", val)
		}

		inputs[inputKey] = resolvedInput
		logrus.WithFields(logrus.Fields{
			"val":      val,
			"key":      inputKey,
			"resolved": resolvedInput,
		}).Infof("Resolved expression")
	}

	// Invoke
	fnSpec := &types.TaskInvocationSpec{
		TaskId: a.task.Id,
		Type:   taskDef,
		Inputs: inputs,
	}

	// TODO concurrent invocations
	_, err := a.api.Invoke(a.wfi.Metadata.Id, fnSpec)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":  a.wfi.Metadata.Id,
			"err": err,
		}).Errorf("Failed to execute task")
		return err
	}

	return nil
}
