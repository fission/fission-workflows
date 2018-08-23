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
	"github.com/imdario/mergo"
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
	wfiLog.Info("Applying action: abort")
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
	wfiLog.Infof("Applying action: fail (%v)", a.Err)
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

func (a *ActionInvokeTask) Apply() error {
	log := a.logger()
	span := opentracing.StartSpan("pkg/controller/invocation/actions/InvokeTask",
		opentracing.ChildOf(a.ec.Span()))

	// Find task
	task, ok := types.GetTask(a.Wf, a.Wfi, a.Task.Id)
	if !ok {
		return fmt.Errorf("task '%v' could not be found", a.Wfi.ID())
	}

	// Check if function has been resolved
	if task.Status.FnRef == nil {
		return fmt.Errorf("no resolved Task could be found for FunctionRef '%v'", task.Spec.FunctionRef)
	}

	// Pre-execution: Resolve expression inputs
	exprEvalStart := time.Now()
	inputs, err := a.resolveInputs(a.Task.Inputs)
	exprEvalDuration.Observe(float64(time.Now().Sub(exprEvalStart)))
	if err != nil {
		log.Error(err)
		return err
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
		i, err := typedvalues.FormatTypedValueMap(typedvalues.DefaultParserFormatter, spec.GetInputs())
		if err != nil {
			log.Errorf("Failed to format inputs for debugging: %v", err)
		} else {
			log.Debugf("Using inputs: %v", i)
		}
	}
	ctx := opentracing.ContextWithSpan(context.Background(), span)
	_, err = a.API.Invoke(spec, api.WithContext(ctx), api.PostTransformer(a.postTransformer))
	if err != nil {
		log.Errorf("Failed to execute task: %v", err)
		return err
	}
	span.Finish()
	return nil
}

func (a *ActionInvokeTask) postTransformer(ti *types.TaskInvocation) error {
	task, _ := types.GetTask(a.Wf, a.Wfi, a.Task.Id)
	if ti.GetStatus().Successful() {
		output := task.GetSpec().GetOutput()
		if output != nil {
			if output.GetType() == typedvalues.TypeExpression {
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

func (a *ActionInvokeTask) resolveOutput(ti *types.TaskInvocation, outputExpr *types.TypedValue) (*types.TypedValue, error) {
	log := a.logger()

	// Setup the scope for the expressions
	scope, err := expr.NewScope(a.Wf, a.Wfi)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create scope for task '%v'", a.Task.Id)
	}
	a.StateStore.Set(a.Wfi.ID(), scope)

	// Inherit scope if this invocation is part of a dynamic invocation
	if len(a.Wfi.Spec.ParentId) != 0 {
		parentScope, ok := a.StateStore.Get(a.Wfi.Spec.ParentId)
		if ok {
			err := mergo.Merge(scope, parentScope)
			if err != nil {
				log.Errorf("Failed to inherit parent scope: %v", err)
			}
		}
	}

	// Add the current output
	scope.Tasks[a.Task.Id].Output = typedvalues.MustFormat(ti.GetStatus().GetOutput())

	// Resolve the output expression
	resolvedOutput, err := expr.Resolve(scope, a.Task.Id, outputExpr)
	if err != nil {
		return nil, err
	}
	return resolvedOutput, nil
}

func (a *ActionInvokeTask) resolveInputs(inputs map[string]*types.TypedValue) (map[string]*types.TypedValue, error) {
	log := a.logger()

	// Setup the scope for the expressions
	scope, err := expr.NewScope(a.Wf, a.Wfi)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create scope for task '%v'", a.Task.Id)
	}
	a.StateStore.Set(a.Wfi.ID(), scope)

	// Inherit scope if this invocation is part of a dynamic invocation
	if len(a.Wfi.Spec.ParentId) != 0 {
		parentScope, ok := a.StateStore.Get(a.Wfi.Spec.ParentId)
		if ok {
			err := mergo.Merge(scope, parentScope)
			if err != nil {
				log.Errorf("Failed to inherit parent scope: %v", err)
			}
		}
	}

	// Resolve each of the inputs (based on priority)
	resolvedInputs := map[string]*types.TypedValue{}
	for _, input := range typedvalues.Prioritize(inputs) {
		resolvedInput, err := expr.Resolve(scope, a.Task.Id, input.Val)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve input field %v: %v", input.Key, err)
		}
		resolvedInputs[input.Key] = resolvedInput
		log.Infof("Resolved field %v: %v -> %v", input.Key, typedvalues.MustFormat(input.Val),
			util.Truncate(typedvalues.MustFormat(resolvedInput), 100))

		// Update the scope with the resolved type
		scope.Tasks[a.Task.Id].Inputs[input.Key] = typedvalues.MustFormat(resolvedInput)
	}
	return resolvedInputs, nil
}
