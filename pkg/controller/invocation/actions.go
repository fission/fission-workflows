package invocation

import (
	"context"
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/controller/expr"
	"github.com/fission/fission-workflows/pkg/scheduler"
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

func (a *ActonAbort) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureInvocationContext(cec)
	a.InvocationID = ec.Invocation().ID()
	return a
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

func (a *ActionFail) Eval(cec controller.EvalContext) controller.Action {
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
	return a
}

func (a *ActionFail) Apply() error {
	log.Infof("Applying action: fail (%v)", a.Err)
	return a.API.Fail(a.InvocationID, a.Err)
}

// ActionInvokeTask invokes a function
type ActionInvokeTask struct {
	ec         *controller.EvalState
	Wf         *types.Workflow
	Wfi        *types.WorkflowInvocation
	API        *api.Task
	Task       *scheduler.InvokeTaskAction
	StateStore *expr.Store
}

func (a *ActionInvokeTask) Eval(cec controller.EvalContext) controller.Action {
	panic("not implemented")
}

func (a *ActionInvokeTask) logger() logrus.FieldLogger {
	return logrus.WithFields(logrus.Fields{
		"invocation": a.Wfi.ID(),
		"workflow":   a.Wf.ID(),
		"task":       a.Task.Id,
	})
}

func (a *ActionInvokeTask) String() string {
	return fmt.Sprintf("task/run(%s)", a.Task.GetId())
}

func (a *ActionInvokeTask) Apply() error {
	log := a.logger()
	span := opentracing.StartSpan(fmt.Sprintf("/task/%s", a.Task.GetId()), opentracing.ChildOf(a.ec.Span().Context()))
	span.SetTag("task", a.Task.GetId())
	defer span.Finish()

	// Find task
	task, ok := types.GetTask(a.Wf, a.Wfi, a.Task.Id)
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
		err := fmt.Errorf("no resolved Task could be found for FunctionRef '%v'", task.Spec.FunctionRef)
		span.LogKV("error", err)
		return err
	}

	// Pre-execution: Resolve expression inputs
	var inputs map[string]*typedvalues.TypedValue
	if len(a.Task.Inputs) > 0 {
		var err error
		exprEvalStart := time.Now()
		inputs, err = a.resolveInputs(a.Task.Inputs)
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
		TaskId:       a.Task.Id,
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
	task, _ := types.GetTask(a.Wf, a.Wfi, a.Task.Id)
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
	scope, err := expr.NewScope(parentScope, a.Wf, a.Wfi)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create scope for task '%v'", a.Task.Id)
	}
	a.StateStore.Set(a.Wfi.ID(), scope)

	// Add the current output
	scope.Tasks[a.Task.Id].Output = typedvalues.MustUnwrap(ti.GetStatus().GetOutput())

	// Resolve the output expression
	resolvedOutput, err := expr.Resolve(scope, a.Task.Id, outputExpr)
	if err != nil {
		return nil, err
	}
	return resolvedOutput, nil
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

	// Setup the scope for the expressions
	scope, err := expr.NewScope(parentScope, a.Wf, a.Wfi)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create scope for task '%v'", a.Task.Id)
	}
	a.StateStore.Set(a.Wfi.ID(), scope)

	// Resolve each of the inputs (based on priority)
	resolvedInputs := map[string]*typedvalues.TypedValue{}
	for _, input := range typedvalues.Prioritize(inputs) {
		resolvedInput, err := expr.Resolve(scope, a.Task.Id, input.Val)
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
		scope.Tasks[a.Task.Id].Inputs[input.Key] = typedvalues.MustUnwrap(resolvedInput)
	}
	return resolvedInputs, nil
}
