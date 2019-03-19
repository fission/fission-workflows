package controller

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/fission/fission-workflows/pkg/api/store"
	"github.com/fission/fission-workflows/pkg/controller/ctrl"
	"github.com/fission/fission-workflows/pkg/controller/executor"
	"github.com/fission/fission-workflows/pkg/controller/expr"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/scheduler"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/typedvalues/controlflow"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/golang/protobuf/ptypes"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

const (
	DefaultMaxRuntime = 10 * time.Minute
)

// InvocationController is the controller for ensuring the processing of a single workflow invocation.
type InvocationController struct {
	invocationID  string
	executor      *executor.LocalExecutor
	invocationAPI *api.Invocation
	taskAPI       *api.Task
	scheduler     *scheduler.InvocationScheduler
	StateStore    *expr.Store // Future: just grab the initial state of the parent, instead of constantly rebuilding it.
	span          opentracing.Span
	logger        *logrus.Entry

	errorCount int
}

func NewInvocationController(invocationID string, executor *executor.LocalExecutor, invocationAPI *api.Invocation,
	taskAPI *api.Task, scheduler *scheduler.InvocationScheduler, stateStore *expr.Store,
	span opentracing.Span, logger *logrus.Entry) *InvocationController {

	return &InvocationController{
		invocationID:  invocationID,
		executor:      executor,
		invocationAPI: invocationAPI,
		taskAPI:       taskAPI,
		scheduler:     scheduler,
		StateStore:    stateStore,
		span:          span,
		logger:        logger,
	}
}

func (c *InvocationController) Eval(ctx context.Context, processValue *ctrl.Event) ctrl.Result {
	// Ensure that the entity is a workflow invocation
	invocation, ok := processValue.Updated.(*types.WorkflowInvocation)
	if !ok {
		return ctrl.Err{Err: fmt.Errorf("entity expected %T, but was %T",
			&types.WorkflowInvocation{}, processValue.Updated)}
	}

	// The reason that we skip this event is that it is quickly followed by the TaskSucceeded event, which causes a
	// duplicate execution due to a race condition between the evaluation of the TaskAdded and the processing of the
	// TaskSucceeded event. In the future we should eliminate this special case by improving the task status tracking.
	if processValue.Event.GetType() == events.EventInvocationTaskAdded {
		return ctrl.Done{}
	}

	// Ensure that it is the correct invocation
	if invocation.ID() != c.invocationID {
		return ctrl.Err{Err: fmt.Errorf("invocation ID expected %v, but was %v", c.invocationID, invocation.ID())}
	}

	// Ensure that the workflow is present in the invocation
	if invocation.Workflow() == nil {
		err := errors.New("workflow is not present in the invocation")
		c.executor.Submit(&executor.Task{
			TaskID:  invocation.ID() + ".fail",
			GroupID: invocation.ID(),
			Apply: func() error {
				return c.invocationAPI.Fail(invocation.ID(), err)
			},
		})
		return ctrl.Err{Err: err}
	}

	// Do not evaluate as long as there still tasks to be executed
	if activeTaskCount := c.executor.GetGroupTasks(invocation.ID()); activeTaskCount > 0 {
		return ctrl.Err{Err: fmt.Errorf("invocation still has %d open task(s) to be executed", activeTaskCount)}
	}

	// Check if the invocation is not in a terminal state
	if invocation.GetStatus().Finished() {
		return ctrl.Done{Msg: fmt.Sprintf("invocation is in a terminal state (%v)",
			invocation.GetStatus().GetStatus().String())}
	}

	// Check if the deadline has not been exceeded
	deadline, err := ptypes.Timestamp(invocation.GetSpec().GetDeadline())
	if err != nil {
		createdAt, err := ptypes.Timestamp(invocation.GetMetadata().GetCreatedAt())
		if err != nil {
			err := errors.New("failed to read deadline and createdAt")
			c.executor.Submit(&executor.Task{
				TaskID:  invocation.ID() + ".fail",
				GroupID: invocation.ID(),
				Apply: func() error {
					return c.invocationAPI.Fail(invocation.ID(), err)
				},
			})
			return ctrl.Err{Err: err}
		}
		deadline = createdAt.Add(DefaultMaxRuntime)
	}
	if time.Now().After(deadline) {
		err := errors.New("deadline exceeded")
		c.executor.Submit(&executor.Task{
			TaskID:  invocation.ID() + ".fail",
			GroupID: invocation.ID(),
			Apply: func() error {
				return c.invocationAPI.Fail(invocation.ID(), err)
			},
		})
		return ctrl.Err{Err: err}
	}

	// Check if we did not exceed the error count
	if c.errorCount > 0 {
		err := errors.New("error count exceeded")
		c.executor.Submit(&executor.Task{
			TaskID:  invocation.ID() + ".fail",
			GroupID: invocation.ID(),
			Apply: func() error {
				return c.invocationAPI.Fail(invocation.ID(), err)
			},
		})
		return ctrl.Err{Err: err}
	}

	// Check if all tasks have finished
	if allTasksFinished(invocation) {
		output, outputHeaders, err := determineTaskOutput(invocation)
		if err != nil {
			c.executor.Submit(&executor.Task{
				TaskID:  invocation.ID() + ".fail",
				GroupID: invocation.ID(),
				Apply: func() error {
					return c.invocationAPI.Fail(invocation.ID(), err)
				},
			})
			return ctrl.Err{Err: err}
		} else {
			c.executor.Submit(&executor.Task{
				TaskID:  invocation.ID() + ".success",
				GroupID: invocation.ID(),
				Apply: func() error {
					return c.invocationAPI.Complete(invocation.ID(), output, outputHeaders)
				},
			})
			return ctrl.Success{Msg: "all tasks of the invocation have completed"}
		}
	}

	// Defer the heuristic part of the evaluation to the scheduler.
	schedule, err := c.scheduler.Evaluate(invocation)
	if err != nil {
		return ctrl.Err{Err: err}
	}

	// If the scheduler indicates to fail, fail the invocation immediately.
	if abortAction := schedule.GetAbort(); abortAction != nil {
		err := errors.New(abortAction.Reason)
		c.executor.Submit(&executor.Task{
			TaskID:  invocation.ID() + ".fail",
			GroupID: invocation.ID(),
			Apply: func() error {
				return c.invocationAPI.Fail(invocation.ID(), err)
			},
		})
		return ctrl.Err{Err: err}
	}

	// Prepare (prewarm) the tasks listed in the schedule.
	for _, action := range schedule.GetPrepareTasks() {
		c.executor.Submit(&executor.Task{
			TaskID:  fmt.Sprintf("%s.prewarm.%s", invocation.ID(), action.TaskID),
			GroupID: invocation.ID(),
			Apply: func() error {
				task, ok := invocation.Task(action.TaskID)
				if !ok || task == nil {
					return fmt.Errorf("no task in workflow with ID: %s", action.TaskID)
				}
				taskRunSpec := types.NewTaskInvocationSpec(invocation.ID(), task)
				return c.taskAPI.Prepare(taskRunSpec, action.GetExpectedAtTime())
			},
		})
	}

	// Execute the tasks listed in the schedule.
	for _, action := range schedule.GetRunTasks() {
		c.executor.Submit(&executor.Task{
			TaskID:  fmt.Sprintf("%s.run.%s", invocation.ID(), action.TaskID),
			GroupID: invocation.ID(),
			Apply: func() error {
				return c.execTask(invocation, action.GetTaskID())
			},
		})
	}

	return ctrl.Success{
		Msg: fmt.Sprintf("scheduled execution of %d tasks and preparation of %d tasks",
			len(schedule.GetRunTasks()), len(schedule.GetPrepareTasks())),
	}
}

func (c *InvocationController) execTask(invocation *types.WorkflowInvocation, taskID string) error {
	log := c.logger
	span := opentracing.StartSpan(fmt.Sprintf("/task/%s", taskID), opentracing.ChildOf(c.span.Context()))
	span.SetTag("task", taskID)
	defer span.Finish()

	// Find task
	task, ok := invocation.Task(taskID)
	if !ok {
		err := fmt.Errorf("task '%v' could not be found", invocation.ID())
		span.LogKV("error", err)
		return err
	}

	span.SetTag("fnref", task.GetStatus().GetFnRef())
	if log.Level == logrus.DebugLevel {
		var err error
		var inputs interface{}
		inputs, err = typedvalues.UnwrapMapTypedValue(task.GetSpec().GetInputs())
		if err != nil {
			inputs = fmt.Sprintf("error: %v", err)
		}
		span.LogKV("inputs", inputs)
	}

	// Check if function has been resolved
	if task.GetStatus().GetFnRef() == nil {
		err := fmt.Errorf("no resolved task could be found for FunctionRef '%v'", task.Spec.FunctionRef)
		span.LogKV("error", err)
		return err
	}

	// Pre-execution: Resolve expression inputs
	var inputs map[string]*typedvalues.TypedValue
	if len(task.GetSpec().GetInputs()) > 0 {
		var err error
		inputs, err = c.resolveInputs(invocation, task.ID(), task.GetSpec().GetInputs())
		if err != nil {
			log.Error(err)
			span.LogKV("error", err)
			return err
		}

		if log.Level == logrus.DebugLevel {
			var err error
			var resolvedInputs interface{}
			resolvedInputs, err = typedvalues.UnwrapMapTypedValue(inputs)
			if err != nil {
				resolvedInputs = fmt.Sprintf("error: %v", err)
			}
			span.LogKV("resolved_inputs", resolvedInputs)
		}
	}

	// Invoke task
	taskRunSpec := types.NewTaskInvocationSpec(invocation.ID(), task)
	taskRunSpec.Inputs = inputs
	log.Infof("Executing function: %v", taskRunSpec.GetFnRef().Format())
	if log.Level == logrus.DebugLevel {
		i, err := typedvalues.UnwrapMapTypedValue(taskRunSpec.GetInputs())
		if err != nil {
			log.Errorf("Failed to format inputs for debugging: %v", err)
		} else {
			log.Debugf("Using inputs: %v", i)
		}
	}
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	updated, err := c.taskAPI.Invoke(taskRunSpec, api.WithContext(ctx), api.AwaitWorklow(10*time.Second),
		api.PostTransformer(func(ti *types.TaskInvocation) error {
			return c.transformTaskRunOutputs(invocation, ti)
		}))
	if err != nil {
		span.LogKV("error", err)
		return err
	}

	span.SetTag("status", updated.GetStatus().GetStatus().String())
	if !updated.GetStatus().Successful() {
		span.LogKV("error", updated.GetStatus().GetError().String())
	}
	if log.Level == logrus.DebugLevel {
		var err error
		var output interface{}
		output, err = typedvalues.Unwrap(updated.GetStatus().GetOutput())
		if err != nil {
			output = fmt.Sprintf("error: %v", err)
		}
		span.LogKV("output", output)
	}
	return nil
}

func (c *InvocationController) resolveInputs(invocation *types.WorkflowInvocation, taskID string,
	inputs map[string]*typedvalues.TypedValue) (map[string]*typedvalues.TypedValue, error) {
	// Inherit scope if invocation has a parent
	log := c.logger
	var parentScope *expr.Scope
	if len(invocation.Spec.ParentId) != 0 {
		var ok bool
		parentScope, ok = c.StateStore.Get(invocation.Spec.ParentId)
		if !ok {
			log.Warnf("Could not find parent scope (%s) of scope (%s)", invocation.Spec.ParentId, invocation.ID())
		}
	}

	// Setup the scope for the expressions
	scope, err := expr.NewScope(parentScope, invocation)
	if err != nil {
		return nil, fmt.Errorf("failed to create scope for task '%v': %v", taskID, err)
	}
	c.StateStore.Set(invocation.ID(), scope)

	// Resolve each of the inputs (based on priority)
	resolvedInputs := map[string]*typedvalues.TypedValue{}
	for _, input := range typedvalues.Prioritize(inputs) {
		resolvedInput, err := expr.Resolve(scope, taskID, input.Val)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve input field %v: %v", input.Key, err)
		}
		resolvedInputs[input.Key] = resolvedInput
		if input.Val.ValueType() == typedvalues.TypeExpression {
			log.Infof("Input field resolved '%v': %v -> %v", input.Key,
				util.Truncate(typedvalues.MustUnwrap(input.Val), 100),
				util.Truncate(typedvalues.MustUnwrap(resolvedInput), 100))
		} else {
			log.Infof("Input field loaded '%v': %v", input.Key,
				util.Truncate(typedvalues.MustUnwrap(resolvedInput), 100))
		}

		// Update the scope with the resolved type
		scope.Tasks[taskID].Inputs[input.Key] = typedvalues.MustUnwrap(resolvedInput)
	}
	return resolvedInputs, nil
}

func (c *InvocationController) resolveOutput(invocation *types.WorkflowInvocation, ti *types.TaskInvocation,
	outputExpr *typedvalues.TypedValue) (*typedvalues.TypedValue, error) {
	log := c.logger

	// Inherit scope if invocation has a parent
	taskID := ti.GetSpec().GetTask().GetMetadata().GetId()
	var parentScope *expr.Scope
	if len(invocation.Spec.ParentId) != 0 {
		var ok bool
		parentScope, ok = c.StateStore.Get(invocation.Spec.ParentId)
		if !ok {
			log.Warnf("Could not find parent scope (%s) of scope (%s)", invocation.Spec.ParentId, invocation.ID())
		}
	}

	// Setup the scope for the expressions
	scope, err := expr.NewScope(parentScope, invocation)
	if err != nil {
		return nil, fmt.Errorf("failed to create scope for task '%v': %v", taskID, err)
	}
	c.StateStore.Set(invocation.ID(), scope)

	// Add the current output
	scope.Tasks[taskID].Output = typedvalues.MustUnwrap(ti.GetStatus().GetOutput())

	// Resolve the output expression
	resolvedOutput, err := expr.Resolve(scope, taskID, outputExpr)
	if err != nil {
		return nil, err
	}
	return resolvedOutput, nil
}

func (c *InvocationController) transformTaskRunOutputs(invocation *types.WorkflowInvocation, ti *types.TaskInvocation) error {
	log := c.logger
	task := ti.GetSpec().GetTask()
	if !ti.GetStatus().Successful() {
		// Nothing to transform
		return nil
	}

	// If there is an output set for the task, replace the actual output of the task runOnce with this.
	output := task.GetSpec().GetOutput()
	if output != nil {
		// Resolve the task output if it is an expression.
		if output.ValueType() == typedvalues.TypeExpression {
			tv, err := c.resolveOutput(invocation, ti, output)
			if err != nil {
				return err
			}
			output = tv
		}
		log.Debugf("replaced the task run output (old: %v, new: %v)", ti.GetStatus().Output, output)
		ti.GetStatus().Output = output
	}

	// If there are output headers set for the task, replace the actual output headers of the task runOnce with these.
	outputHeaders := task.GetSpec().GetOutputHeaders()
	if outputHeaders != nil {
		// Resolve the header outputs if it is an expression.
		if outputHeaders.ValueType() == typedvalues.TypeExpression {
			tv, err := c.resolveOutputHeaders(invocation, ti, outputHeaders)
			if err != nil {
				return err
			}
			outputHeaders = tv
		}
		log.Debugf("replaced the task run output headers (old: %v, new: %v)",
			ti.GetStatus().OutputHeaders, outputHeaders)
		ti.GetStatus().OutputHeaders = outputHeaders
	}

	return nil
}

func (c *InvocationController) resolveOutputHeaders(invocation *types.WorkflowInvocation, ti *types.TaskInvocation,
	outputHeadersExpr *typedvalues.TypedValue) (*typedvalues.TypedValue, error) {

	taskID := ti.GetSpec().GetTask().GetMetadata().GetId()
	// Inherit scope if invocation has a parent
	var parentScope *expr.Scope
	if len(invocation.Spec.ParentId) != 0 {
		var ok bool
		parentScope, ok = c.StateStore.Get(invocation.Spec.ParentId)
		if !ok {
			c.logger.Warnf("Could not find parent scope (%s) of scope (%s)", invocation.Spec.ParentId, invocation.ID())
		}
	}

	// Setup the scope for the expressions
	scope, err := expr.NewScope(parentScope, invocation)
	if err != nil {
		return nil, fmt.Errorf("failed to create scope for task '%v': %v", taskID, err)
	}
	c.StateStore.Set(invocation.ID(), scope)

	// Add the current outputHeaders
	scope.Tasks[taskID].OutputHeaders = typedvalues.MustUnwrap(ti.GetStatus().GetOutputHeaders())

	// Resolve the outputHeaders expression
	resolvedOutputHeaders, err := expr.Resolve(scope, taskID, outputHeadersExpr)
	if err != nil {
		return nil, err
	}
	return resolvedOutputHeaders, nil
}

func determineTaskOutput(invocation *types.WorkflowInvocation) (output *typedvalues.TypedValue,
	outputHeaders *typedvalues.TypedValue, err error) {

	success := true
	wf := invocation.GetSpec().GetWorkflow()
	for id := range invocation.Tasks() {
		task := invocation.Status.Tasks[id]
		if !task.GetStatus().Successful() {
			success = false
			break
		}
	}

	var finalOutput *typedvalues.TypedValue
	var finalOutputHeaders *typedvalues.TypedValue
	if len(wf.GetSpec().GetOutputTask()) != 0 {
		finalOutput = controlflow.ResolveTaskOutput(wf.Spec.OutputTask, invocation)
		finalOutputHeaders = controlflow.ResolveTaskOutputHeaders(wf.Spec.OutputTask, invocation)
	}

	if success {
		return finalOutput, finalOutputHeaders, nil
	} else {
		return nil, nil, errors.New("one or more tasks in the workflow have failed")
	}
}

func allTasksFinished(invocation *types.WorkflowInvocation) bool {
	finished := true
	for id := range invocation.Tasks() {
		task, ok := invocation.Status.Tasks[id]
		if !ok || !task.GetStatus().Finished() {
			finished = false
			break
		}
	}
	return finished
}

// InvocationMetaController is the component responsible for the full integration of the invocations reconciliation loop.
//
// Specifically, the meta-controller is responsible for the following:
// - It starts or registers the sensors to ensure that events are routed to this control system.
// - It manages all of the workflow controllers.
// - It provides an executor pool for controllers to submit their tasks to.
type InvocationMetaController struct {
	sensors     []ctrl.Sensor
	executor    *executor.LocalExecutor
	runOnce     *sync.Once
	invocations *store.Invocations
	system      *ctrl.System
}

func NewInvocationMetaController(executor *executor.LocalExecutor, invocations *store.Invocations,
	invocationAPI *api.Invocation, taskAPI *api.Task, scheduler *scheduler.InvocationScheduler, stateStore *expr.Store,
	pollInterval time.Duration) *InvocationMetaController {
	return &InvocationMetaController{
		sensors: []ctrl.Sensor{
			NewInvocationNotificationSensor(invocations),
			NewInvocationStorePollSensor(invocations, pollInterval),
		},
		executor: executor,
		runOnce:  &sync.Once{},
		system: ctrl.NewSystem(func(event *ctrl.Event) (ctrl ctrl.Controller, err error) {
			spanCtx, err := fes.ExtractTracingFromEventMetadata(event.Event.GetMetadata())
			if err != nil {
				logrus.Debugf("Could not extract span from event metadata: %v", err)
			}
			var span opentracing.Span
			if spanCtx != nil {
				span = opentracing.StartSpan("/controller/eval", opentracing.FollowsFrom(spanCtx))
			} else {
				span = opentracing.StartSpan("/controller/eval")
			}
			invocationID := event.Aggregate.Id
			if len(invocationID) == 0 {
				return nil, fmt.Errorf("invocation ID missing in event: %v %v", event.Aggregate, event.Event.GetType())
			}
			return NewInvocationController(invocationID, executor, invocationAPI, taskAPI, scheduler,
				stateStore, span, logrus.WithField("key", invocationID)), nil
		}),
	}
}

func (c *InvocationMetaController) Run() {
	c.runOnce.Do(func() {
		go c.run()
	})
}

func (c *InvocationMetaController) run() {
	// Start the task executor
	c.executor.Start()

	// Start the sensors
	for _, sensor := range c.sensors {
		err := sensor.Start(c.system)
		if err != nil {
			panic(err)
		}
	}

	// Run control system
	c.system.Run()

}

func (c *InvocationMetaController) Close() error {
	err := c.executor.Close()
	err = c.system.Close()
	for _, sensor := range c.sensors {
		err = sensor.Close()
	}
	return err
}

// InvocationNotificationSensor watches the invocations store notifications for workflow events.
type InvocationNotificationSensor struct {
	invocations *store.Invocations
	done        func()
	closeC      <-chan struct{}
}

func NewInvocationNotificationSensor(invocations *store.Invocations) *InvocationNotificationSensor {
	ctx, done := context.WithCancel(context.Background())
	return &InvocationNotificationSensor{
		invocations: invocations,
		done:        done,
		closeC:      ctx.Done(),
	}
}

func (s *InvocationNotificationSensor) Start(evalQueue ctrl.EvalQueue) error {
	go s.Run(evalQueue)
	return nil
}

func (s *InvocationNotificationSensor) Run(evalQueue ctrl.EvalQueue) {
	sub := s.invocations.GetInvocationUpdates()
	if sub == nil {
		logrus.Warn("Workflow store does not support pubsub.")
		return
	}
	logrus.Debug("Listening for invocation events")
	for {
		select {
		case msg := <-sub.Ch:
			notification, err := sub.ToNotification(msg)
			if err != nil {
				logrus.Warnf("Failed to convert pubsub message to notification: %v", err)
			}
			evalQueue.Submit(notification)
		case <-s.closeC:
			err := sub.Close()
			if err != nil {
				logrus.Error(err)
			}
			logrus.Info("Notification listener stopped.")
			return
		}
	}
}

func (s *InvocationNotificationSensor) Close() error {
	s.done()
	return nil
}

// InvocationStorePollSensor polls the invocations store on a set interval.
type InvocationStorePollSensor struct {
	*ctrl.PollSensor
	invocations *store.Invocations
}

func NewInvocationStorePollSensor(invocations *store.Invocations, interval time.Duration) *InvocationStorePollSensor {
	s := &InvocationStorePollSensor{
		invocations: invocations,
	}
	s.PollSensor = ctrl.NewPollSensor(interval, s.Poll)
	return s
}

func (s *InvocationStorePollSensor) Poll(evalQueue ctrl.EvalQueue) {
	for _, aggregate := range s.invocations.List() {
		// Ignore non-workflow entities in workflow store
		if aggregate.Type != types.TypeInvocation {
			logrus.Warnf("Non-invocation entity in invocations store: %v", aggregate)
			continue
		}

		// Get actual workflow
		wf, err := s.invocations.GetInvocation(aggregate.GetId())
		if err != nil {
			logrus.Warnf("Could not retrieve entity from invocations store: %v", aggregate)
			continue
		}

		// Check if the status is not in a terminal state
		switch wf.GetStatus().GetStatus() {
		case types.WorkflowInvocationStatus_ABORTED, types.WorkflowInvocationStatus_FAILED, types.WorkflowInvocationStatus_SUCCEEDED:
			continue
		default:
			// nop
		}

		// Submit evaluation for the workflow invocation
		// The workqueue within in the control system ensures that invocations that are already queued for execution
		// will be ignored.
		evalQueue.Submit(&ctrl.Event{
			Old:     wf,
			Updated: wf,
			Event: &fes.Event{
				Type:      EventRefresh,
				Aggregate: &aggregate,
				Timestamp: ptypes.TimestampNow(),
			},
			Aggregate: aggregate,
		})
	}
}
