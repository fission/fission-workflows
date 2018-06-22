package scheduler

import (
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/graph"
	"github.com/golang/protobuf/ptypes"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var log = logrus.WithField("component", "scheduler")

var (
	metricEvalTime = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "workflows",
		Subsystem: "scheduler",
		Name:      "eval_time",
		Help:      "Statistics of scheduler evaluations",
	})
	metricEvalCount = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "workflows",
		Subsystem: "scheduler",
		Name:      "eval_count",
		Help:      "Number of evaluations",
	})
)

func init() {
	prometheus.MustRegister(metricEvalCount, metricEvalTime)
}

type WorkflowScheduler struct {
}

func (ws *WorkflowScheduler) Evaluate(request *ScheduleRequest) (*Schedule, error) {
	ctxLog := log.WithFields(logrus.Fields{
		"invocation": request.Invocation.ID(),
		"workflow":   request.Workflow.ID(),
	})
	timeStart := time.Now()
	defer func() {
		metricEvalTime.Observe(float64(time.Since(timeStart)))
		metricEvalCount.Inc()
	}()

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

			msg := fmt.Sprintf("Task '%v' failed", t.Invocation.ID())
			if err := t.Invocation.GetStatus().GetError(); err != nil {
				msg = err.Message
			}

			AbortActionAny, _ := ptypes.MarshalAny(&AbortAction{
				Reason: msg,
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
			Id:     taskDef.Task.ID(),
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
