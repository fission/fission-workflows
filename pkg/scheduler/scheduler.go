package scheduler

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/graph"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("component", "scheduler")

// TODO document other approach: scheduler as a controller
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

	ctxLog.Info("Scheduler evaluating...")
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

	// Determine horizon (aka tasks that can be executed now)
	//horizon := map[string]*types.TaskInstance{}
	//for id, task := range openTasks {
	//	// TODO replace graph implementation, with marked and await counter
	//	completedDeps := len(task.Task.Spec.Requires) - len(depGraph.To(graph.Get(depGraph, id)))
	//
	//	ctxLog.WithFields(logrus.Fields{
	//		"completedDeps": completedDeps,
	//		"task":          id,
	//		"max":           int(math.Max(float64(task.Task.Spec.Await), float64(len(task.Task.Spec.Requires)-1))),
	//	}).Infof("Checking if dependencies have been satisfied")
	//	if completedDeps >= int(math.Max(float64(task.Task.Spec.Await), float64(len(task.Task.Spec.Requires)))) {
	//		log.WithField("task", task.Task.Id()).Info("task found on horizon")
	//		horizon = append(horizon, &graph.TaskInstanceNode{TaskInstance: task})
	//	}
	//	// TODO fix graph representation (how to find second horizon)
	//}

	ctxLog.WithField("horizon", horizon).Info("Determined horizon")

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

	if len(schedule.Actions) > 0 {
		ctxLog.WithField("schedule", len(schedule.Actions)).
			WithField("invocation", schedule.InvocationId).
			Info("Determined schedule")
	}
	return schedule, nil
}
