package scheduler

import (
	"fmt"

	"math"

	"github.com/fission/fission-workflows/pkg/types"
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
		"workflow": request.Workflow.Metadata.Id,
		"invoke":   request.Invocation.Metadata.Id,
	})

	ctxLog.Info("Scheduler evaluating...")

	cwf := types.CalculateTaskDependencyGraph(request.Workflow, request.Invocation)

	openTasks := map[string]*types.TaskStatus{} // bool = nothing
	// Fill open tasks
	for id, t := range cwf {
		if t.Invocation == nil {
			openTasks[id] = t
			continue
		}
		if t.Invocation.Status.Status == types.TaskInvocationStatus_FAILED {
			AbortActionAny, _ := ptypes.MarshalAny(&AbortAction{
				Reason: fmt.Sprintf("TaskStatus '%s' failed!", t.Invocation),
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
	horizon := map[string]*types.TaskStatus{}
	for id, task := range openTasks {
		if len(task.Requires) == 0 {
			horizon[id] = task
			break
		}

		completedDeps := 0
		for depName := range task.Requires {
			t, ok := cwf[depName]
			if !ok {
				panic(fmt.Sprintf("Unknown task dependency: %v", depName))
			}

			if t.Invocation != nil && t.Invocation.Status.Status.Finished() {
				completedDeps = completedDeps + 1
			}
		}

		log.WithFields(log.Fields{
			"completedDeps": completedDeps,
			"task":          id,
			"max":           int(math.Max(float64(task.Await), float64(len(task.Requires)-1))),
		}).Infof("Checking if dependencies have been satisfied")
		if completedDeps >= int(math.Max(float64(task.Await), float64(len(task.Requires)))) {
			log.WithField("task", task.Id).Info("task found on horizon")
			horizon[id] = task
		}
	}

	ctxLog.WithField("horizon", horizon).Debug("Determined horizon")

	// Determine schedule nodes
	for taskId, taskDef := range horizon {
		// Fetch input
		inputs := taskDef.Inputs
		invokeTaskAction, _ := ptypes.MarshalAny(&InvokeTaskAction{
			Id:     taskId,
			Inputs: inputs,
		})

		schedule.Actions = append(schedule.Actions, &Action{
			Type:    ActionType_INVOKE_TASK,
			Payload: invokeTaskAction,
		})
	}

	ctxLog.WithField("schedule", len(schedule.Actions)).
		WithField("invocation", schedule.InvocationId).
		Info("Determined schedule")

	return schedule, nil
}
