package scheduler

import (
	"fmt"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
)

type WorkflowScheduler struct {
}

func (ws *WorkflowScheduler) Evaluate(request *ScheduleRequest) (*Schedule, error) {
	schedule := &Schedule{
		InvocationId: request.Invocation.Metadata.Id,
		CreatedAt:    ptypes.TimestampNow(),
		Actions:      []*Action{},
	}

	ctxLog := log.WithFields(log.Fields{
		"workflow": request.Workflow,
		"invoke":   request.Invocation,
	})

	ctxLog.Info("Scheduler evaluating...")

	openTasks := map[string]*types.Task{} // bool = nothing
	// Fill open tasks
	for id, t := range request.Workflow.Spec.Src.Tasks {
		invokedTask, ok := request.Invocation.Status.Tasks[id]
		if !ok {
			openTasks[id] = t
			continue
		}
		if invokedTask.Status.Status == types.FunctionInvocationStatus_FAILED {
			AbortActionAny, _ := ptypes.MarshalAny(&AbortAction{
				Reason: fmt.Sprintf("Task '%s' failed!", invokedTask),
			})

			abortAction := &Action{
				Type:    ActionType_ABORT,
				Payload: AbortActionAny,
			}
			schedule.Actions = append(schedule.Actions, abortAction)
		}
	}
	if len(schedule.Actions) > 0 {
		return schedule, nil
	}

	// Determine horizon (aka tasks that can be executed now)
	horizon := map[string]*types.Task{}
	for id, task := range openTasks {
		free := true
		for depName := range task.GetDependencies() {
			if _, ok := openTasks[depName]; ok {
				free = false
				break
			}
		}
		if free {
			horizon[id] = task
		}
	}

	ctxLog.WithField("horizon", horizon).Debug("Determined horizon")

	// Determine schedule nodes
	for taskId, task := range horizon {
		// Fetch input
		inputs := map[string]string{}
		if len(task.GetDependencies()) == 0 {
			inputs[types.INPUT_MAIN] = request.Invocation.Spec.Input
		} else {

			for depName, dep := range task.GetDependencies() {
				// TODO check for overwrites (especially the alias)
				inputs[depName] = string(request.Invocation.Status.Tasks[depName].Status.Output)
				if len(dep.Alias) > 0 {
					inputs[dep.Alias] = inputs[depName]
				}

			}
			// Decide which one is the main input
		}

		invokeTaskAction, _ := ptypes.MarshalAny(&InvokeTaskAction{
			Id:    taskId,
			Input: inputs,
		})

		schedule.Actions = append(schedule.Actions, &Action{
			Type:    ActionType_INVOKE_TASK,
			Payload: invokeTaskAction,
		})
	}

	ctxLog.WithField("schedule", schedule).Info("Determined schedule")

	return schedule, nil
}
