package scheduler

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/graph"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("component", "scheduler")

type WorkflowScheduler struct {
}

func (ws *WorkflowScheduler) Evaluate(request *ScheduleRequest) (*Schedule, error) {
	ctxLog := log.WithFields(logrus.Fields{
		"wfi": request.Invocation.Metadata.Id,
	})

	schedule := &Schedule{
		InvocationId: request.Invocation.Metadata.Id,
		CreatedAt:    ptypes.TimestampNow(),
		Actions:      []*Action{},
	}

	ctxLog.Debug("Scheduler evaluating...")
	cwf := types.GetTaskContainers(request.Workflow, request.Invocation)

	// Fill open tasks
	openTasks := map[string]*types.TaskInstance{}
	for id, t := range cwf {
		if t.Invocation == nil || t.Invocation.Status.Status == types.TaskInvocationStatus_UNKNOWN {
			openTasks[id] = t
			continue
		}
		if t.Invocation.Status.Status == types.TaskInvocationStatus_FAILED {
			AbortActionAny, _ := ptypes.MarshalAny(&AbortAction{
				Reason: fmt.Sprintf("taskContainer '%s' failed!", t.Invocation),
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

	depGraph := graph.Parse(graph.NewTaskInstanceIterator(openTasks))
	horizon := graph.Roots(depGraph)

	// Determine schedule nodes
	for _, node := range horizon {
		taskDef := node.(*graph.TaskInstanceNode)
		// Fetch input
		// TODO might be Status.Inputs instead of Spec.Inputs
		inputs := taskDef.Task.Spec.Inputs
		invokeTaskAction, _ := ptypes.MarshalAny(&InvokeTaskAction{
			Id:     taskDef.Task.Id(),
			Inputs: inputs,
		})

		schedule.Actions = append(schedule.Actions, &Action{
			Type:    ActionType_INVOKE_TASK,
			Payload: invokeTaskAction,
		})
	}

	ctxLog.WithField("schedule", len(schedule.Actions)).Info("Determined schedule")
	return schedule, nil
}
