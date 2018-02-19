package action

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/api/function"
	"github.com/fission/fission-workflows/pkg/api/invocation"
	"github.com/fission/fission-workflows/pkg/controller/expr"
	"github.com/fission/fission-workflows/pkg/scheduler"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/sirupsen/logrus"
)

// Abort aborts an invocation.
type Abort struct {
	Api          *invocation.Api
	InvocationId string
}

func (a *Abort) Id() string {
	return a.InvocationId // Invocation
}

func (a *Abort) Apply() error {
	logrus.WithField("Wfi", a.Id()).Info("Applying abort action")
	return a.Api.Cancel(a.InvocationId)
}

// Fail halts an invocation.
type Fail struct {
	Api          *invocation.Api
	InvocationId string
	Err          error
}

func (a *Fail) Id() string {
	return a.InvocationId // Invocation
}

func (a *Fail) Apply() error {
	logrus.WithField("Wfi", a.Id()).Info("Executing fail action")
	return a.Api.Fail(a.InvocationId, a.Err)
}

// InvokeTask invokes a function
type InvokeTask struct {
	Wf   *types.Workflow
	Wfi  *types.WorkflowInvocation
	Expr expr.Resolver
	Api  *function.Api
	Task *scheduler.InvokeTaskAction
}

func (a *InvokeTask) Id() string {
	return a.Wfi.Id() // Invocation
}

func (a *InvokeTask) Apply() error {
	actionLog := logrus.WithField("Wfi", a.Id())
	// Find Task (static or dynamic)
	task, ok := types.GetTask(a.Wf, a.Wfi, a.Task.Id)
	if !ok {
		return fmt.Errorf("task '%v' could not be found", a.Id())
	}
	actionLog.Infof("Invoking function '%s' for Task '%s'", task.Spec.FunctionRef, a.Task.Id)

	// Check if resolved
	if task.Status.Resolved == nil {
		return fmt.Errorf("no resolved Task could be found for FunctionRef '%v'", task.Spec.FunctionRef)
	}

	// Resolve the inputs
	inputs := map[string]*types.TypedValue{}
	queryScope := expr.NewScope(a.Wf, a.Wfi)
	for inputKey, val := range a.Task.Inputs {
		resolvedInput, err := a.Expr.Resolve(queryScope, a.Task.Id, val)
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
		TaskId: a.Task.Id,
		Type:   task.Status.Resolved,
		Inputs: inputs,
	}

	_, err := a.Api.Invoke(a.Wfi.Metadata.Id, fnSpec)
	if err != nil {
		actionLog.WithFields(logrus.Fields{
			"id": a.Wfi.Metadata.Id,
		}).Errorf("Failed to execute task: %v", err)
		return err
	}

	return nil
}
