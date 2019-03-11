package invocation

import (
	"context"
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/controller/expr"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/opentracing/opentracing-go"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

//
// Invocation-specific actions
//

// ActonAbort aborts an invocation.
type ActonAbort struct {
	API          *api.Invocation
	InvocationID string
}

func (a *ActonAbort) Eval(cec controller.EvalContext) []controller.Action {
	ec := EnsureInvocationContext(cec)
	a.InvocationID = ec.Invocation().ID()
	return []controller.Action{a}
}

func (a *ActonAbort) Apply() error {
	log.Info("Applying action: abort")
	return a.API.Cancel(a.InvocationID)
}

// ActionFail halts an invocation.
type ActionFail struct {
	API          *api.Invocation
	InvocationID string
	Err          error
}

func (a *ActionFail) Eval(cec controller.EvalContext) []controller.Action {
	ec := EnsureInvocationContext(cec)
	a.InvocationID = ec.Invocation().ID()
	if a.Err == nil {
		if s, ok := ec.EvalState().Last(); ok {
			a.Err = s.Error
		}
	}
	if a.Err == nil {
		a.Err = errors.New("unknown error has occurred")
	}
	return []controller.Action{a}
}

func (a *ActionFail) Apply() error {
	log.Infof("Applying action: fail (%v)", a.Err)
	return a.API.Fail(a.InvocationID, a.Err)
}

// ActionInvokeTask invokes a function
type ActionInvokeTask struct {
	ec         *controller.EvalState
	Wfi        *types.WorkflowInvocation
	API        *api.Task
	TaskID     string
	StateStore *expr.Store
}

func (a *ActionInvokeTask) Eval(cec controller.EvalContext) []controller.Action {
	panic("not implemented")
}

func (a *ActionInvokeTask) logger() logrus.FieldLogger {
	return logrus.WithFields(logrus.Fields{
		"invocation": a.Wfi.ID(),
		"workflow":   a.Wfi.Workflow().ID(),
		"task":       a.TaskID,
	})
}

func (a *ActionInvokeTask) String() string {
	return fmt.Sprintf("task/run(%s)", a.TaskID)
}

func (a *ActionInvokeTask) Apply() error {
	taskID := a.TaskID
	log := a.logger()
	span := opentracing.StartSpan(fmt.Sprintf("/task/%s", taskID), opentracing.ChildOf(a.ec.Span().Context()))
	span.SetTag("task", taskID)
	defer span.Finish()

	// Find task
	task, ok := a.Wfi.Task(taskID)
	if !ok {
		err := fmt.Errorf("task '%v' could not be found", a.Wfi.ID())
		span.LogKV("error", err)
		return err
	}

	span.SetTag("fnref", task.GetStatus().GetFnRef())
	if logrus.GetLevel() == logrus.DebugLevel {
		var err error
		var inputs interface{}
		inputs, err = typedvalues.UnwrapMapTypedValue(task.GetSpec().GetInputs())
		if err != nil {
			inputs = fmt.Sprintf("error: %v", err)
		}
		span.LogKV("inputs", inputs)
	}

	// Check if function has been resolved
	if task.Status.FnRef == nil {
		err := fmt.Errorf("no resolved DefaultTask could be found for FunctionRef '%v'", task.Spec.FunctionRef)
		span.LogKV("error", err)
		return err
	}

	// Pre-execution: Resolve expression inputs
	var inputs map[string]*typedvalues.TypedValue
	if len(task.GetSpec().GetInputs()) > 0 {
		var err error
		exprEvalStart := time.Now()
		inputs, err = a.resolveInputs(task.GetSpec().GetInputs())
		exprEvalDuration.Observe(float64(time.Now().Sub(exprEvalStart)))
		if err != nil {
			log.Error(err)
			span.LogKV("error", err)
			return err
		}

		if logrus.GetLevel() == logrus.DebugLevel {
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
	spec := &types.TaskInvocationSpec{
		FnRef:        task.Status.FnRef,
		TaskId:       a.TaskID,
		InvocationId: a.Wfi.ID(),
		Inputs:       inputs,
	}
	log.Infof("Executing function: %v", spec.GetFnRef().Format())
	if logrus.GetLevel() == logrus.DebugLevel {
		i, err := typedvalues.UnwrapMapTypedValue(spec.GetInputs())
		if err != nil {
			log.Errorf("Failed to format inputs for debugging: %v", err)
		} else {
			log.Debugf("Using inputs: %v", i)
		}
	}

	ctx := opentracing.ContextWithSpan(context.Background(), span)
	updated, err := a.API.Invoke(spec, api.WithContext(ctx), api.PostTransformer(a.postTransformer))
	if err != nil {
		log.Errorf("Failed to execute task: %v", err)
		span.LogKV("error", err)
		return err
	}
	span.SetTag("status", updated.GetStatus().GetStatus().String())
	if !updated.GetStatus().Successful() {
		span.LogKV("error", updated.GetStatus().GetError().String())
	}
	if logrus.GetLevel() == logrus.DebugLevel {
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

func (a *ActionInvokeTask) postTransformer(ti *types.TaskInvocation) error {
	task, _ := a.Wfi.Task(a.TaskID)
	if ti.GetStatus().Successful() {
		output := task.GetSpec().GetOutput()
		if output != nil {
			if output.ValueType() == typedvalues.TypeExpression {
				tv, err := a.resolveOutput(ti, output)
				if err != nil {
					return err
				}
				output = tv
			}
			ti.GetStatus().Output = output
		}
		outputHeaders := task.GetSpec().GetOutputHeaders()
		if outputHeaders != nil {
			if outputHeaders.ValueType() == typedvalues.TypeExpression {
				tv, err := a.resolveOutputHeaders(ti, outputHeaders)
				if err != nil {
					return err
				}
				outputHeaders = tv
			}
			ti.GetStatus().OutputHeaders = outputHeaders
		}

	}
	return nil
}

func (a *ActionInvokeTask) resolveOutput(ti *types.TaskInvocation, outputExpr *typedvalues.TypedValue) (*typedvalues.TypedValue, error) {
	// Inherit scope if invocation has a parent
	var parentScope *expr.Scope
	if len(a.Wfi.Spec.ParentId) != 0 {
		var ok bool
		parentScope, ok = a.StateStore.Get(a.Wfi.Spec.ParentId)
		if !ok {
			a.logger().Warn("Could not find parent scope (%s) of scope (%s)", a.Wfi.Spec.ParentId, a.Wfi.ID())
		}
	}

	// Setup the scope for the expressions
	scope, err := expr.NewScope(parentScope, a.Wfi)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create scope for task '%v'", a.TaskID)
	}
	a.StateStore.Set(a.Wfi.ID(), scope)

	// Add the current output
	scope.Tasks[a.TaskID].Output = typedvalues.MustUnwrap(ti.GetStatus().GetOutput())

	// Resolve the output expression
	resolvedOutput, err := expr.Resolve(scope, a.TaskID, outputExpr)
	if err != nil {
		return nil, err
	}
	return resolvedOutput, nil
}

func (a *ActionInvokeTask) resolveOutputHeaders(ti *types.TaskInvocation, outputHeadersExpr *typedvalues.TypedValue) (*typedvalues.TypedValue, error) {
	// Inherit scope if invocation has a parent
	var parentScope *expr.Scope
	if len(a.Wfi.Spec.ParentId) != 0 {
		var ok bool
		parentScope, ok = a.StateStore.Get(a.Wfi.Spec.ParentId)
		if !ok {
			a.logger().Warn("Could not find parent scope (%s) of scope (%s)", a.Wfi.Spec.ParentId, a.Wfi.ID())
		}
	}

	// Setup the scope for the expressions
	scope, err := expr.NewScope(parentScope, a.Wfi)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create scope for task '%v'", a.TaskID)
	}
	a.StateStore.Set(a.Wfi.ID(), scope)

	// Add the current outputHeaders
	scope.Tasks[a.TaskID].OutputHeaders = typedvalues.MustUnwrap(ti.GetStatus().GetOutputHeaders())

	// Resolve the outputHeaders expression
	resolvedOutputHeaders, err := expr.Resolve(scope, a.TaskID, outputHeadersExpr)
	if err != nil {
		return nil, err
	}
	return resolvedOutputHeaders, nil
}

func (a *ActionInvokeTask) resolveInputs(inputs map[string]*typedvalues.TypedValue) (map[string]*typedvalues.TypedValue, error) {
	// Inherit scope if invocation has a parent
	var parentScope *expr.Scope
	if len(a.Wfi.Spec.ParentId) != 0 {
		var ok bool
		parentScope, ok = a.StateStore.Get(a.Wfi.Spec.ParentId)
		if !ok {
			a.logger().Warn("Could not find parent scope (%s) of scope (%s)", a.Wfi.Spec.ParentId, a.Wfi.ID())
		}
	}

	taskID := a.TaskID

	// Setup the scope for the expressions
	scope, err := expr.NewScope(parentScope, a.Wfi)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create scope for task '%v'", taskID)
	}
	a.StateStore.Set(a.Wfi.ID(), scope)

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

type actionPrepareTask struct {
	taskSpec   *types.TaskInvocationSpec
	expectedAt time.Time
	api        *api.Task
}

func (a *actionPrepareTask) Apply() error {
	return a.api.Prepare(a.taskSpec, a.expectedAt)
}
