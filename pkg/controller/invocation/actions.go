package invocation

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/api/function"
	"github.com/fission/fission-workflows/pkg/api/invocation"
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/controller/expr"
	"github.com/fission/fission-workflows/pkg/scheduler"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/sirupsen/logrus"
)

//
// Invocation-specific actions
//

// ActonAbort aborts an invocation.
type ActonAbort struct {
	Api          *invocation.Api
	InvocationId string
}

func (a *ActonAbort) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureInvocationContext(cec)
	a.InvocationId = ec.Invocation().Id()
	return a
}

func (a *ActonAbort) Apply() error {
	wfiLog.Info("Applying action: abort")
	return a.Api.Cancel(a.InvocationId)
}

// ActionFail halts an invocation.
type ActionFail struct {
	Api          *invocation.Api
	InvocationId string
	Err          error
}

func (a *ActionFail) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureInvocationContext(cec)
	a.InvocationId = ec.Invocation().Id()
	return a
}

func (a *ActionFail) Apply() error {
	wfiLog.Info("Applying action: fail")
	return a.Api.Fail(a.InvocationId, a.Err)
}

// ActionInvokeTask invokes a function
type ActionInvokeTask struct {
	Wf   *types.Workflow
	Wfi  *types.WorkflowInvocation
	Api  *function.Api
	Task *scheduler.InvokeTaskAction
}

func (a *ActionInvokeTask) Eval(cec controller.EvalContext) controller.Action {
	panic("not implemented")
}

func (a *ActionInvokeTask) Apply() error {
	wfiLog.Infof("Invoking task: %v", a.Task.Id)
	// Find Task (static or dynamic)
	task, ok := types.GetTask(a.Wf, a.Wfi, a.Task.Id)
	if !ok {
		return fmt.Errorf("task '%v' could not be found", a.Wfi.Id())
	}
	wfiLog.Infof("Invoking function '%s' for Task '%s'", task.Spec.FunctionRef, a.Task.Id)

	// Check if resolved
	if task.Status.FnRef == nil {
		return fmt.Errorf("no resolved Task could be found for FunctionRef '%v'", task.Spec.FunctionRef)
	}

	// ResolveTask the inputs
	inputs := map[string]*types.TypedValue{}
	queryScope := expr.NewScope(a.Wf, a.Wfi)
	for inputKey, val := range a.Task.Inputs {
		resolvedInput, err := expr.Resolve(queryScope, a.Task.Id, val)
		if err != nil {
			wfiLog.WithFields(logrus.Fields{
				"val":      val,
				"inputKey": inputKey,
			}).Errorf("Failed to parse input: %v", err)
			return err
		}

		inputs[inputKey] = resolvedInput
		wfiLog.WithFields(logrus.Fields{
			"val":      val,
			"key":      inputKey,
			"resolved": resolvedInput,
		}).Infof("Resolved expression")
	}

	// Invoke
	fnSpec := &types.TaskInvocationSpec{
		FnRef:        task.Status.FnRef,
		TaskId:       a.Task.Id,
		InvocationId: a.Wfi.Id(),
		Inputs:       inputs,
	}

	_, err := a.Api.Invoke(fnSpec)
	if err != nil {
		wfiLog.WithFields(logrus.Fields{
			"id": a.Wfi.Metadata.Id,
		}).Errorf("Failed to execute task: %v", err)
		return err
	}

	return nil
}
