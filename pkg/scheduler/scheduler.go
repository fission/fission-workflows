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

	openTasks := map[string][]string{} // bool = nothing
	// Fill open tasks
	for id, t := range request.Workflow.Spec.Src.Tasks {
		invokedTask, ok := request.Invocation.Status.Tasks[id] // TODO Ignore failed tasks for now
		if !ok {
			openTasks[id] = t.Dependencies
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

	// Determine graph
	horizon := []string{}
	for t, deps := range openTasks {
		free := true
		for _, dep := range deps {
			if _, ok := openTasks[dep]; ok {
				free = false
				break
			}
		}
		if free {
			horizon = append(horizon, t)
		}
	}

	ctxLog.WithField("horizon", horizon).Debug("Determined horizon")

	// Determine schedule nodes
	for _, taskId := range horizon {
		invokeTaskAction, _ := ptypes.MarshalAny(&InvokeTaskAction{
			Id: taskId,
		})

		schedule.Actions = append(schedule.Actions, &Action{
			Type:    ActionType_INVOKE_TASK,
			Payload: invokeTaskAction,
		})
	}

	ctxLog.WithField("schedule", schedule).Info("Determined schedule")

	return schedule, nil
}
