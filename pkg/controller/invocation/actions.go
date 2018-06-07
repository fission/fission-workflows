package invocation

import (
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/controller/expr"
	"github.com/fission/fission-workflows/pkg/scheduler"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/imdario/mergo"
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
	Wf         *types.Workflow
	Wfi        *types.WorkflowInvocation
	API        *api.Task
	Task       *scheduler.InvokeTaskAction
	StateStore *expr.Store
}

func (a *ActionInvokeTask) Eval(cec controller.EvalContext) controller.Action {
	panic("not implemented")
}

func (a *ActionInvokeTask) resolveInputs() (map[string]*types.TypedValue, error) {
	// Resolve the inputs
	scope, err := expr.NewScope(a.Wf, a.Wfi)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create scope for task '%v'", a.Task.Id)
	}
	a.StateStore.Set(a.Wfi.ID(), scope)

	// Inherit scope if this invocation is part of a dynamic decision
	if len(a.Wfi.Spec.ParentId) != 0 {
		parentScope, ok := a.StateStore.Get(a.Wfi.Spec.ParentId)
		if ok {
			err := mergo.Merge(scope, parentScope)
			if err != nil {
				logrus.Errorf("Failed to inherit parent scope: %v", err)
			}
		}
	}
	inputs := map[string]*types.TypedValue{}
	for _, input := range typedvalues.Prioritize(a.Task.Inputs) {
		resolvedInput, err := expr.Resolve(scope, a.Task.Id, input.Val)
		if err != nil {
			wfiLog.WithFields(logrus.Fields{
				"val": input.Key,
				"key": input.Val,
			}).Errorf("Failed to resolve input: %v", err)
			return nil, err
		}

		inputs[input.Key] = resolvedInput
		wfiLog.WithFields(logrus.Fields{
			"key": input.Key,
		}).Infof("Resolved input: %v -> %v", typedvalues.MustFormat(input.Val), typedvalues.MustFormat(resolvedInput))

		// Update the scope with the resolved type
		scope.Tasks[a.Task.Id].Inputs[input.Key] = typedvalues.MustFormat(resolvedInput)
	}
	return inputs, nil
}

func (a *ActionInvokeTask) Apply() error {
	wfiLog.Infof("Running task: %v", a.Task.Id)
	// Find Task (static or dynamic)
	task, ok := types.GetTask(a.Wf, a.Wfi, a.Task.Id)
	if !ok {
		return fmt.Errorf("task '%v' could not be found", a.Wfi.ID())
	}
	wfiLog.Infof("Invoking function '%s' for Task '%s'", task.Spec.FunctionRef, a.Task.Id)

	// Check if function has been resolved
	if task.Status.FnRef == nil {
		return fmt.Errorf("no resolved Task could be found for FunctionRef '%v'", task.Spec.FunctionRef)
	}

	// Resolve inputs
	exprEvalStart := time.Now()
	inputs, err := a.resolveInputs()
	exprEvalDuration.Observe(float64(time.Now().Sub(exprEvalStart)))
	if err != nil {
		return err
	}

	// Invoke
	fnSpec := &types.TaskInvocationSpec{
		FnRef:        task.Status.FnRef,
		TaskId:       a.Task.Id,
		InvocationId: a.Wfi.ID(),
		Inputs:       inputs,
	}

	_, err = a.API.Invoke(fnSpec)
	if err != nil {
		wfiLog.WithFields(logrus.Fields{
			"id": a.Wfi.Metadata.Id,
		}).Errorf("Failed to execute task: %v", err)
		return err
	}

	return nil
}
