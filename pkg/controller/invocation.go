package controller

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/fission/fission-workflows/pkg/api/function"
	"github.com/fission/fission-workflows/pkg/api/invocation"
	"github.com/fission/fission-workflows/pkg/controller/action"
	"github.com/fission/fission-workflows/pkg/controller/expr"
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

var wfiLog = log.WithField("component", "controller-wi")

type InvocationController struct {
	invokeCache   fes.CacheReader
	wfCache       fes.CacheReader
	functionApi   *function.Api
	invocationApi *invocation.Api
	scheduler     *scheduler.WorkflowScheduler
	sub           *pubsub.Subscription
	exprParser    expr.Resolver
	cancelFn      context.CancelFunc

	workQueue chan Action

	// Queued keeps track of which invocations still have actions in the workQueue
	states map[string]*ControlState
	// TODO add active cache
}

func NewInvocationController(invokeCache fes.CacheReader, wfCache fes.CacheReader,
	workflowScheduler *scheduler.WorkflowScheduler, functionApi *function.Api, invocationApi *invocation.Api,
	exprParser expr.Resolver) *InvocationController {
	return &InvocationController{
		invokeCache:   invokeCache,
		wfCache:       wfCache,
		scheduler:     workflowScheduler,
		functionApi:   functionApi,
		invocationApi: invocationApi,
		exprParser:    exprParser,
		workQueue:     make(chan Action, WorkQueueSize),

		// States maintains an active cache of currently running invocations, with execution related data.
		// This state information is considered preemptable and can be removed or lost at any time.
		states: map[string]*ControlState{},
	}
}

func (cr *InvocationController) Init(sctx context.Context) error {
	ctx, cancelFn := context.WithCancel(sctx)
	cr.cancelFn = cancelFn

	// Subscribe to invocation creations and task events.
	selector := labels.InSelector("aggregate.type", "invocation", "function")

	if invokePub, ok := cr.invokeCache.(pubsub.Publisher); ok {
		cr.sub = invokePub.Subscribe(pubsub.SubscriptionOptions{
			Buf:           NotificationBuffer,
			LabelSelector: selector,
		})

		// Invocation Notification lane
		go func(ctx context.Context) {
			for {
				select {
				case notification := <-cr.sub.Ch:
					wfiLog.WithField("labels", notification.Labels()).Debug("Handling invocation notification.")
					switch n := notification.(type) {
					case *fes.Notification:
						cr.HandleNotification(n)
					default:
						wfiLog.WithField("notification", n).Warn("Ignoring unknown notification type")
					}
				case <-ctx.Done():
					wfiLog.WithField("ctx.err", ctx.Err()).Debug("Notification listener closed.")
					return
				}
			}
		}(ctx)
	}

	// workQueue loop
	go func(ctx context.Context) {
		for {
			select {
			case action := <-cr.workQueue:

				// TODO limit goroutine pool size
				go func() {
					wfiLog.WithFields(logrus.Fields{
						"type":   fmt.Sprintf("%T", action),
						"action": fmt.Sprintf("%v", action),
					}).Info("Executing action...")
					cs := cr.controlState(action.Id())
					err := action.Apply()
					if err != nil {
						wfiLog.WithFields(logrus.Fields{
							"action": fmt.Sprintf("%T", action),
						}).Errorf("action failed: %v", err)
						cs.AddError(err)
					} else {
						cs.ResetError()
					}
					cs.DecrementQueueSize()
				}()

			case <-ctx.Done():
				wfiLog.WithField("ctx.err", ctx.Err()).Debug("workQueue closed.")
				return
			}
		}
	}(ctx)

	return nil
}

func (cr *InvocationController) HandleNotification(msg *fes.Notification) error {
	wfiLog.WithFields(logrus.Fields{
		"notification": msg.EventType,
		"labels":       msg.Labels(),
	}).Info("Handling invocation notification!")

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
		wfiLog.WithField("type", msg.EventType).Warn("Controller ignores unknown event.")
	}
	return nil
}

func (cr *InvocationController) HandleTick() error {
	wfiLog.Debug("Controller tick...")
	// Options: refresh projection, send ping, cancel invocation
	// Short loop (invocations the controller is actively tracking) and long loop (to check if there are any orphans)

	// Short control loop
	entities := cr.invokeCache.List()
	for _, entity := range entities {
		wi := aggregates.NewWorkflowInvocation(entity.Id, nil)
		err := cr.invokeCache.Get(wi)
		if err != nil {
			return err
		}

		// Also lock for actions
		cr.evaluate(wi.WorkflowInvocation)
	}
	return nil
}

// TODO return error
func (cr *InvocationController) evaluate(invoc *types.WorkflowInvocation) {
	state := cr.controlState(invoc.Metadata.Id)
	state.Lock()
	defer state.Unlock()

	// Fetch the workflow relevant to the invocation
	wf := aggregates.NewWorkflow(invoc.Spec.WorkflowId, nil)
	err := cr.wfCache.Get(wf)
	if err != nil {
		wfiLog.Errorf("Controller failed to get workflow for invocation '%s': %v", invoc.Spec.WorkflowId, err)
		return
	}
	wfi := &types.WorkflowInstance{
		Workflow:   wf.Workflow,
		Invocation: invoc,
	}

	if wfi.Invocation.Status.Finished() {
		wfiLog.Debugf("No need to evaluate finished invocation %v", wfi.Invocation.Metadata.Id)
		return
	}

	// TODO move to constructor
	ff := &UntilNotEmptyFilter{
		filters: []Filter{
			&PendingActionsFilter{},
			&ErrorLimitExceededFilter{
				invocationApi: cr.invocationApi,
			},
			&CompletedFilter{},
			&TimeoutFilter{
				invocationApi: cr.invocationApi,
			},
			&WorkflowReadyFilter{},
			&CheckCompletedFilter{
				invocationApi: cr.invocationApi,
			},
			&ScheduleFilter{
				scheduler:     cr.scheduler,
				invocationApi: cr.invocationApi,
				functionApi:   cr.functionApi,
				exprParser:    cr.exprParser,
			},
		},
	}
	actions, err := ff.Filter(wfi, state)
	if err != nil {
		// TODO implement retries
		wfiLog.Errorf("Invocation failed: %v", err)
		actions = []Action{&action.Fail{
			Api:          cr.invocationApi,
			InvocationId: wfi.Invocation.Id(),
		}}
	}
	ok := cr.submit(actions...)
	if !ok {
		wfiLog.Warn("Not all actions could be submitted (work queue full)")
	}
}

func (cr *InvocationController) Close() error {
	wfiLog.Info("Closing controller...")
	if invokePub, ok := cr.invokeCache.(pubsub.Publisher); ok {
		err := invokePub.Unsubscribe(cr.sub)
		if err != nil {
			return err
		}
	}

	cr.cancelFn()
	return nil
}

func (cr *InvocationController) submit(actions ...Action) (submitted bool) {
	for _, action := range actions {
		select {
		case cr.workQueue <- action:
			// Ok
			cr.controlState(action.Id()).IncrementQueueSize()
			submitted = true
			wfiLog.WithField("wfi", action.Id()).
				Infof("submitted action: '%s'", reflect.TypeOf(action))
		default:
			// Action overflow
			submitted = false
		}
	}
	return submitted
}

func (cr *InvocationController) controlState(id string) *ControlState {
	s, ok := cr.states[id]
	if ok {
		return s
	} else {
		return NewControlState()
	}
}

//
// Filters
//

type UntilNotEmptyFilter struct {
	filters []Filter
}

func (cf *UntilNotEmptyFilter) Filter(wfi *types.WorkflowInstance, status *ControlState) (actions []Action, err error) {
	for _, f := range cf.filters {
		fa, err := f.Filter(wfi, status)
		if fa != nil && len(fa) > 0 {
			for _, v := range fa {
				actions = append(actions, v)
			}
		}
		if err != nil || actions != nil {
			break
		}
	}
	return actions, err
}

type PendingActionsFilter struct{}

func (fp *PendingActionsFilter) Filter(wfi *types.WorkflowInstance, status *ControlState) (actions []Action, err error) {
	// Check if there are still open actions for this invocation
	if status.QueueSize > 0 {
		wfiLog.Info("Invocation still has pending actions.")
		actions = []Action{}
	}
	return actions, err
}

type ErrorLimitExceededFilter struct {
	invocationApi *invocation.Api
}

func (el *ErrorLimitExceededFilter) Filter(wfi *types.WorkflowInstance, status *ControlState) (actions []Action, err error) {
	// Check if the graph has been failing too often
	if status.ErrorCount > MaxErrorCount {
		wfiLog.Infof("canceling due to error count %v exceeds max error count  %v",
			status.ErrorCount, MaxErrorCount)
		actions = append(actions, &action.Fail{
			Api:          el.invocationApi,
			InvocationId: wfi.Invocation.Id(),
		})
	}
	return actions, err
}

type CompletedFilter struct {
}

func (cf *CompletedFilter) Filter(wfi *types.WorkflowInstance, status *ControlState) (actions []Action, err error) {
	if wfi.Invocation.Status.Finished() {
		wfiLog.Infof("No need to evaluate finished invocation %v", wfi.Invocation.Metadata.Id)
		actions = []Action{}
	}
	return actions, err
}

type TimeoutFilter struct {
	invocationApi *invocation.Api
}

func (tf *TimeoutFilter) Filter(wfi *types.WorkflowInstance, status *ControlState) (actions []Action, err error) {
	// For now: kill after 10 min
	duration := time.Now().Unix() - wfi.Invocation.Metadata.CreatedAt.Seconds
	if duration > int64(InvocationTimeout.Seconds()) {
		wfiLog.Infof("cancelling due to timeout; %v exceeds max timeout %v", duration, int64(InvocationTimeout.Seconds()))
		actions = append(actions, &action.Fail{
			Api:          tf.invocationApi,
			InvocationId: wfi.Invocation.Id(),
		})
	}
	return actions, err
}

type WorkflowReadyFilter struct {
}

func (wr *WorkflowReadyFilter) Filter(wfi *types.WorkflowInstance, status *ControlState) (actions []Action, err error) {
	// Check if workflow is in the right state to use.
	if !wfi.Workflow.Status.Ready() {
		// TODO backoff
		wfiLog.WithField("wf.status", wfi.Workflow.Status.Status).Error("Workflow is not ready yet.")
		actions = []Action{}
	}
	return actions, err
}

type ScheduleFilter struct {
	scheduler     *scheduler.WorkflowScheduler
	invocationApi *invocation.Api
	functionApi   *function.Api
	exprParser    expr.Resolver
}

func (sf *ScheduleFilter) Filter(wfi *types.WorkflowInstance, status *ControlState) (actions []Action, err error) {

	// Request a execution plan from the scheduler
	schedule, err := sf.scheduler.Evaluate(&scheduler.ScheduleRequest{
		Invocation: wfi.Invocation,
		Workflow:   wfi.Workflow,
	})

	// Execute the actions as specified in the execution plan
	for _, a := range schedule.Actions {
		switch a.Type {
		case scheduler.ActionType_ABORT:
			actions = append(actions, &action.Fail{
				Api:          sf.invocationApi,
				InvocationId: wfi.Invocation.Id(),
			})
		case scheduler.ActionType_INVOKE_TASK:
			invokeAction := &scheduler.InvokeTaskAction{}
			err := ptypes.UnmarshalAny(a.Payload, invokeAction)
			if err != nil {
				wfiLog.Errorf("Failed to unpack scheduler action: %v", err)
			}
			actions = append(actions, &action.InvokeTask{
				Wf:   wfi.Workflow,
				Wfi:  wfi.Invocation,
				Expr: sf.exprParser,
				Api:  sf.functionApi,
				Task: invokeAction,
			})
		default:
			logrus.Warnf("Unknown scheduler action: '%v'", a)
		}
	}
	return actions, err
}

type CheckCompletedFilter struct {
	invocationApi *invocation.Api
}

func (cc *CheckCompletedFilter) Filter(wfi *types.WorkflowInstance, status *ControlState) (actions []Action, err error) {
	// Check if the workflow invocation is complete
	tasks := types.GetTasks(wfi.Workflow, wfi.Invocation)
	finished := true
	for id := range tasks {
		t, ok := wfi.Invocation.Status.Tasks[id]
		if !ok || !t.Status.Finished() {
			finished = false
			break
		}
	}
	if finished {
		var finalOutput *types.TypedValue
		if len(wfi.Workflow.Spec.OutputTask) != 0 {
			t, ok := wfi.Invocation.Status.Tasks[wfi.Workflow.Spec.OutputTask]
			if !ok {
				panic("Could not find output task status in completed invocation")
			}
			finalOutput = t.Status.Output
		}

		// TODO submit instead of executing
		err = cc.invocationApi.MarkCompleted(wfi.Invocation.Id(), finalOutput)
	}
	return actions, err
}
