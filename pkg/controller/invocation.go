package controller

import (
	"context"

	"fmt"

	"time"

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
)

const (
	NOTIFICATION_BUFFER = 100
	INVOCATION_TIMEOUT  = time.Duration(10) * time.Minute
)

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
	errorCount  int
	recentError error
	inQueue     int
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
		workQueue:     make(chan Action, 50),

		// States maintains an active cache of currently running invocations, with execution related data.
		// This state information is considered preemptable and can be removed or lost at any time.
		states:        map[string]ControlState{},
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
	}

	// workQueue loop
	go func(ctx context.Context) {
		for {
			select {
			case action := <-cr.workQueue:
				state := cr.states[action.Id()]

				if state.inQueue > 0 {
					state.inQueue -= 1
				}
				cr.states[action.Id()] = state

				go func() { // TODO limit goroutine pool size
					err := action.Apply()
					state := cr.states[action.Id()]
					if err != nil {
						logrus.WithField("action", action).Error("WorkflowInvocation action failed")
						state.recentError = err
						state.errorCount += 1
					} else {
						state.errorCount = 0
					}
					cr.states[action.Id()] = state
				}()

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

	// Short control loop
	entities := cr.invokeCache.List()
	for _, entity := range entities {
		wi := aggregates.NewWorkflowInvocation(entity.Id, nil)
		err := cr.invokeCache.Get(wi)
		if err != nil {
			logrus.Error(err)
			return
		}

		// Check if we actually need to evaluate
		if wi.Status.Status.Finished() {
			// TODO remove finished wfi from active cache
			continue
		}

		// TODO check if workflow invocation is in a backoff

		cr.evaluate(wi.WorkflowInvocation)
	}
}

func (cr *InvocationController) evaluate(invoc *types.WorkflowInvocation) {
	state := cr.states[invoc.Metadata.Id]

	// Check if there are still open actions
	if state.inQueue > 0 {
		return
	}

	// Check if the graph has been failing too often
	if state.errorCount > 5 {
		logrus.Infof("canceling invocation %v due to error count", invoc.Metadata.Id)
		err := cr.invocationApi.Cancel(invoc.Metadata.Id) // TODO just submit?
		if err != nil {
			logrus.Errorf("failed to cancel timed out invocation: %v", err)
		}
		return
	}

	// Check if the workflow invocation is in the right state
	if invoc.Status.Status.Finished() {
		logrus.Infof("No need to evaluate finished invocation %v", invoc.Metadata.Id)
		return
	}

	// For now: kill after 10 min
	if (time.Now().Unix() - invoc.Metadata.CreatedAt.Seconds) > int64(INVOCATION_TIMEOUT.Seconds()) {
		logrus.Infof("canceling timeout invocation %v", invoc.Metadata.Id)
		err := cr.invocationApi.Cancel(invoc.Metadata.Id) // TODO just submit?
		if err != nil {
			logrus.Errorf("failed to cancel timed out invocation: %v", err)
		}
		return
	}

	// Fetch the workflow relevant to the invocation
	wf := aggregates.NewWorkflow(invoc.Spec.WorkflowId, nil)
	err := cr.wfCache.Get(wf)
	if err != nil {
		logrus.Errorf("Controller failed to get workflow for invocation '%s': %v", invoc.Spec.WorkflowId, err)
		return
	}

	// Check if workflow is in the right state to use.
	if wf.Status.Status != types.WorkflowStatus_READY {
		logrus.WithField("wf.status", wf.Status.Status).Error("Workflow has not been parsed yet.")
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
			logrus.Errorf("failed to mark invocation as complete: %v", err)
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

func (cr *InvocationController) submit(action Action) (submitted bool) {
	select {
	case cr.workQueue <- action:
		// Ok
		state := cr.states[action.Id()]
		state.inQueue += 1
		cr.states[action.Id()] = state
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
	return a.invocationId // Invocation
}

func (a *abortAction) Apply() error {
	logrus.Infof("aborting: '%v'", a.invocationId)
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
	queryScope := expr.NewScope(a.wf, a.wfi)
	for inputKey, val := range a.task.Inputs {
		resolvedInput, err := a.expr.Resolve(queryScope, queryScope.Tasks[a.task.Id], nil, val)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"val":      val,
				"inputKey": inputKey,
			}).Errorf("Failed to parse input: %v", err)
			return err
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
