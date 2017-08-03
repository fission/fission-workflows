package scheduler

import (
	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
)

type WorkflowScheduler struct {
}

func (ws *WorkflowScheduler) Evaluate(request *ScheduleRequest) (*Schedule, error) {
	ctxLog := log.WithFields(log.Fields{
		"workflow": request.Workflow,
		"invoke":   request.Invocation,
	})

	ctxLog.Info("Scheduler evaluating...")

	openTasks := map[string][]string{} // bool = nothing
	// Fill open tasks
	for id, t := range request.Workflow.Spec.Src.Tasks {
		_, ok := request.Invocation.Status.Tasks[id] // TODO Ignore failed tasks for now
		if !ok {
			openTasks[id] = t.Dependencies
		}
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
	scheduledNodes := []*ScheduledNode{}
	for _, taskId := range horizon {
		scheduledNodes = append(scheduledNodes, &ScheduledNode{
			Id: taskId,
		})
	}

	schedule := &Schedule{
		InvocationId: request.Invocation.Metadata.Id,
		CreatedAt:    ptypes.TimestampNow(),
		Nodes:        scheduledNodes,
	}

	ctxLog.WithField("schedule", schedule).Info("Determined schedule")

	return schedule, nil
}
