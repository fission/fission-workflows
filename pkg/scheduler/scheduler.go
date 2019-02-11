package scheduler

import (
	"time"

	"github.com/fission/fission-workflows/pkg/types"
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

type Policy interface {
	Evaluate(invocation *types.WorkflowInvocation) (*Schedule, error)
}

func init() {
	prometheus.MustRegister(metricEvalCount, metricEvalTime)
}

type InvocationScheduler struct {
	policy Policy
}

func NewInvocationScheduler(policy Policy) *InvocationScheduler {
	return &InvocationScheduler{
		policy: policy,
	}
}

func (ws *InvocationScheduler) Evaluate(invocation *types.WorkflowInvocation) (*Schedule, error) {
	ctxLog := log.WithFields(logrus.Fields{
		"invocation": invocation.ID(),
		"workflow":   invocation.Workflow().ID(),
	})
	timeStart := time.Now()
	defer func() {
		metricEvalTime.Observe(float64(time.Since(timeStart)))
		metricEvalCount.Inc()
	}()

	schedule, err := ws.policy.Evaluate(invocation)
	if err != nil {
		return nil, err
	}

	ctxLog.Debugf("Determined schedule: %v", schedule)
	return schedule, nil
}

func (m *Schedule) Add(action *Action) {
	m.Actions = append(m.Actions, action)
}

func newActionInvoke(task *types.Task) *Action {
	invokeTaskAction, _ := ptypes.MarshalAny(&InvokeTaskAction{
		TaskID: task.ID(),
	})

	return &Action{
		Type:    ActionType_INVOKE_TASK,
		Payload: invokeTaskAction,
	}
}

func newActionAbort(msg string) *Action {
	AbortActionAny, _ := ptypes.MarshalAny(&AbortAction{
		Reason: msg,
	})

	return &Action{
		Type:    ActionType_ABORT,
		Payload: AbortActionAny,
	}
}

func newActionPrepare(taskID string, expectedAt time.Time) *Action {
	ts, _ := ptypes.TimestampProto(expectedAt)
	AbortActionAny, _ := ptypes.MarshalAny(&PrepareTaskAction{
		TaskID:     taskID,
		ExpectedAt: ts,
	})

	return &Action{
		Type:    ActionType_PREPARE_TASK,
		Payload: AbortActionAny,
	}
}
