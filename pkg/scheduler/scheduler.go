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

func (ws *WorkflowScheduler) Evaluate(invocation *types.WorkflowInvocation) (*Schedule, error) {
	ctxLog := log.WithFields(logrus.Fields{
		"invocation": invocation.ID(),
		"workflow":   invocation.Workflow().ID(),
	})
	timeStart := time.Now()
	defer func() {
		metricEvalTime.Observe(float64(time.Since(timeStart)))
		metricEvalCount.Inc()
	}()

	schedule := &Schedule{
		InvocationId: invocation.ID(),
		CreatedAt:    ptypes.TimestampNow(),
		Actions:      []*Action{},
	}

	ctxLog.Debug("Scheduler evaluating...")

	// Fill open tasks
	openTasks := map[string]*types.TaskInvocation{}
	for id, task := range invocation.Tasks() {
		taskRun, ok := invocation.TaskInvocation(id)
		if !ok {
			// TODO extract this to types
			taskRun = &types.TaskInvocation{
				Metadata: types.NewObjectMetadata(id),
				Spec: &types.TaskInvocationSpec{
					Task:         task,
					FnRef:        task.GetStatus().GetFnRef(),
					TaskId:       task.ID(),
					InvocationId: invocation.ID(),
					Inputs:       task.GetSpec().GetInputs(),
				},
				Status: &types.TaskInvocationStatus{
					Status: types.TaskInvocationStatus_UNKNOWN,
				},
			}
		}
		switch taskRun.GetStatus().GetStatus() {
		case types.TaskInvocationStatus_UNKNOWN:
			openTasks[id] = taskRun
		case types.TaskInvocationStatus_FAILED:
			msg := fmt.Sprintf("Task '%v' failed", task.ID())
			if err := task.GetStatus().GetError(); err != nil {
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
		taskDef := node.(*graph.TaskInvocationNode)
		// Fetch input
		inputs := taskDef.Task().Spec.Inputs
		invokeTaskAction, _ := ptypes.MarshalAny(&InvokeTaskAction{
			Id:     taskDef.Task().ID(),
			Inputs: inputs,
		})

		schedule.Actions = append(schedule.Actions, &Action{
			Type:    ActionType_INVOKE_TASK,
			Payload: invokeTaskAction,
		})
	}

	ctxLog.WithField("schedule", len(schedule.Actions)).Debug("Determined schedule")
	return schedule, nil
}
