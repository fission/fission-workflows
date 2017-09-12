package actions

import (
	"fmt"

	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/controller/query"
	"github.com/fission/fission-workflow/pkg/scheduler"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/sirupsen/logrus"
)

func InvokeTask(action *scheduler.InvokeTaskAction, wf *types.Workflow, invoc *types.WorkflowInvocation,
	queryParser query.ExpressionParser, api *function.Api) error {

	task, ok := invoc.Status.DynamicTasks[action.Id]
	if !ok {
		task, ok = wf.Spec.Tasks[action.Id]
		if !ok {
			return fmt.Errorf("unknown task '%v'", action.Id)
		}
	}

	// Resolve type of the task
	taskDef, ok := wf.Status.ResolvedTasks[task.FunctionRef]
	if !ok {
		return fmt.Errorf("no resolved task could be found for task '%v'", task.FunctionRef)
	}

	// Resolve the inputs
	inputs := map[string]*types.TypedValue{}
	queryScope := query.NewScope(wf, invoc)
	for inputKey, val := range action.Inputs {
		resolvedInput, err := queryParser.Resolve(queryScope, queryScope.Tasks[action.Id], val)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"val":      val,
				"inputKey": inputKey,
			}).Warnf("Failed to parse input: %v", err)
			continue
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
		TaskId: action.Id,
		Type:   taskDef,
		Inputs: inputs,
	}

	// TODO concurrent invocations
	_, err := api.Invoke(invoc.Metadata.Id, fnSpec)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"id":  invoc.Metadata.Id,
			"err": err,
		}).Errorf("Failed to execute task")
		return err
	}
	return nil
}
