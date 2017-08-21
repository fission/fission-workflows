package scheduler

import (
	"fmt"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
	"math"
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
	for id, t := range request.Workflow.Spec.Tasks {
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
		if len(task.GetDependencies()) == 0 {
			horizon[id] = task
			break
		}

		completedDeps := 0
		for depName := range task.GetDependencies() {
			if _, ok := openTasks[depName]; !ok {
				completedDeps = completedDeps + 1
				break
			}
		}
		log.WithFields(log.Fields{
			"completedDeps": completedDeps,
			"task":          id,
			"max":           int(math.Max(float64(task.DependenciesAwait), float64(len(task.GetDependencies())-1))),
		}).Infof("Checking if dependencies have been satisfied")
		if completedDeps > int(math.Max(float64(task.DependenciesAwait), float64(len(task.GetDependencies())-1))) {
			horizon[id] = task
		}
	}

	ctxLog.WithField("horizon", horizon).Debug("Determined horizon")

	// Determine schedule nodes
	for taskId, taskDef := range horizon {
		// Fetch input

		invokeTaskAction, _ := ptypes.MarshalAny(&InvokeTaskAction{
			Id:     taskId,
			Inputs: taskDef.Inputs, // TODO interpolated inputs
		})

		schedule.Actions = append(schedule.Actions, &Action{
			Type:    ActionType_INVOKE_TASK,
			Payload: invokeTaskAction,
		})
	}

	ctxLog.WithField("schedule", schedule).Info("Determined schedule")

	return schedule, nil
}
