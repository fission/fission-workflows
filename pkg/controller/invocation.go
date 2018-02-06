package controller

import (
	"context"

	"fmt"

	"time"

	"sync"

	"github.com/fission/fission-workflows/pkg/api/function"
	"github.com/fission/fission-workflows/pkg/api/invocation"
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
	"reflect"
	"sync/atomic"
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
	states map[string]ControlState
	// TODO add active cache
}

type ControlState struct {
	ErrorCount  uint32
	RecentError error
	QueueSize   uint32
	lock        sync.Mutex
}

func (cs ControlState) AddError(err error) uint32 {
	cs.RecentError = err
	return atomic.AddUint32(&cs.ErrorCount, 1)
}

func (cs ControlState) ResetError() {
	cs.RecentError = nil
	cs.ErrorCount = 0
}

func (cs ControlState) IncrementQueueSize() uint32 {
	return atomic.AddUint32(&cs.QueueSize, 1)
}

func (cs ControlState) DecrementQueueSize() uint32 {
	return atomic.AddUint32(&cs.QueueSize, ^uint32(0)) // TODO avoid overflow
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
		states: map[string]ControlState{},
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

				go func() { // TODO limit goroutine pool size
					err := action.Apply()
					if err != nil {
						wfiLog.WithField("action", action).Errorf("action failed: %v", err)
						cr.states[action.Id()].AddError(err)
					} else {
						cr.states[action.Id()].ResetError()
					}
					cr.states[action.Id()].DecrementQueueSize()
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
	}).Info("Handling notification!")

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

		cr.evaluate(wi.WorkflowInvocation)
	}
	return nil
}

// TODO return error
func (cr *InvocationController) evaluate(invoc *types.WorkflowInvocation) {
	state := cr.states[invoc.Metadata.Id]
	state.lock.Lock()
	defer state.lock.Unlock()

	// Check if there are still open actions for this invocation
	if state.QueueSize > 0 {
		return
	}

	// Check if we actually need to evaluate
	if invoc.Status.Status.Finished() {
		// TODO remove finished wfi from active cache
		return
	}

	// TODO check if workflow invocation is in a back-off

	// Check if the graph has been failing too often
	if state.ErrorCount > MaxErrorCount {
		wfiLog.Infof("canceling due to error count %v exceeds max error count  %v", state.ErrorCount, MaxErrorCount)
		ok := cr.submit(&abortAction{
			api:          cr.invocationApi,
			invocationId: invoc.Metadata.Id,
		})
		if !ok {
			wfiLog.Error("failed to cancel timed out invocation.")
		}
		return
	}

	// Check if the workflow invocation is in the right state
	if invoc.Status.Status.Finished() {
		wfiLog.Infof("No need to evaluate finished invocation %v", invoc.Metadata.Id)
		return
	}

	// For now: kill after 10 min
	duration := time.Now().Unix() - invoc.Metadata.CreatedAt.Seconds
	if duration > int64(InvocationTimeout.Seconds()) {
		wfiLog.Infof("cancelling due to timeout; %v exceeds max timeout %v", duration, int64(InvocationTimeout.Seconds()))
		ok := cr.submit(&abortAction{
			api:          cr.invocationApi,
			invocationId: invoc.Metadata.Id,
		})
		if !ok {
			wfiLog.Error("failed to cancel timed out invocation.")
		}
		return
	}

	// Fetch the workflow relevant to the invocation
	wf := aggregates.NewWorkflow(invoc.Spec.WorkflowId, nil)
	err := cr.wfCache.Get(wf)
	if err != nil {
		wfiLog.Errorf("Controller failed to get workflow for invocation '%s': %v", invoc.Spec.WorkflowId, err)
		return
	}

	// Check if workflow is in the right state to use.
	if wf.Status.Status != types.WorkflowStatus_READY {
		// TODO backoff
		wfiLog.WithField("wf.status", wf.Status.Status).Error("Workflow has not been parsed yet.")
		return
	}

	// Check if the workflow invocation is complete
	tasks := types.Tasks(wf.Workflow, invoc)
	finished := true
	for id := range tasks {
		t, ok := invoc.Status.Tasks[id]
		if !ok || !t.Status.Status.Finished() {
			finished = false
			break
		}
	}
	if finished {
		var finalOutput *types.TypedValue
		if len(wf.Spec.OutputTask) != 0 {
			t, ok := invoc.Status.Tasks[wf.Spec.OutputTask]
			if !ok {
				panic("Could not find output task status in completed invocation")
			}
			finalOutput = t.Status.Output
		}

		err := cr.invocationApi.MarkCompleted(invoc.Metadata.Id, finalOutput) // TODO just submit?
		if err != nil {
			wfiLog.Errorf("failed to mark invocation as complete: %v", err)
			return
		}
	}

	// Request a execution plan from the scheduler
	schedule, err := cr.scheduler.Evaluate(&scheduler.ScheduleRequest{
		Invocation: invoc,
		Workflow:   wf.Workflow,
	})

	// Execute the actions as specified in the execution plan
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

func (cr *InvocationController) submit(action Action) (submitted bool) {
	select {
	case cr.workQueue <- action:
		// Ok
		cr.states[action.Id()].IncrementQueueSize()
		submitted = true
		wfiLog.WithField("wfi", action.Id()).
			Infof("submitted action: '%s'", reflect.TypeOf(action))
	default:
		// Action overflow
	}
	return submitted
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
	return a.invocationId // Invocation
}

func (a *abortAction) Apply() error {
	wfiLog.WithField("wfi", a.Id()).Info("Applying abort action")
	return a.api.Cancel(a.invocationId)
}

// invokeTaskAction invokes a function
type invokeTaskAction struct {
	wf   *types.Workflow
	wfi  *types.WorkflowInvocation
	expr expr.Resolver
	api  *function.Api
	task *scheduler.InvokeTaskAction
}

func (a *invokeTaskAction) Id() string {
	return a.wfi.Metadata.Id // Invocation
}

func (a *invokeTaskAction) Apply() error {
	actionLog := wfiLog.WithField("wfi", a.Id())
	// Find task (static or dynamic)
	task, ok := a.wfi.Status.DynamicTasks[a.task.Id]
	if !ok {
		task, ok = a.wf.Spec.Tasks[a.task.Id]
		if !ok {
			return fmt.Errorf("unknown task '%v'", a.task.Id)
		}
	}
	actionLog.Infof("Invoking function '%s' for task '%s'", task.FunctionRef, a.task.Id)

	// Resolve type of the task
	taskDef, ok := a.wf.Status.ResolvedTasks[task.FunctionRef]
	if !ok {
		return fmt.Errorf("no resolved task could be found for FunctionRef '%v'", task.FunctionRef)
	}

	// Resolve the inputs
	inputs := map[string]*types.TypedValue{}
	queryScope := expr.NewScope(a.wf, a.wfi)
	for inputKey, val := range a.task.Inputs {
		resolvedInput, err := a.expr.Resolve(queryScope, queryScope.Tasks[a.task.Id], nil, val)
		if err != nil {
			actionLog.WithFields(logrus.Fields{
				"val":      val,
				"inputKey": inputKey,
			}).Errorf("Failed to parse input: %v", err)
			return err
		}

		inputs[inputKey] = resolvedInput
		actionLog.WithFields(logrus.Fields{
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

	_, err := a.api.Invoke(a.wfi.Metadata.Id, fnSpec)
	if err != nil {
		actionLog.WithFields(logrus.Fields{
			"id":  a.wfi.Metadata.Id,
			"err": err,
		}).Errorf("Failed to execute task")
		return err
	}

	return nil
}
