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

func newRunTaskAction(taskID string) *RunTaskAction {
	return &RunTaskAction{
		TaskID: taskID,
	}
}

func newAbortAction(msg string) *AbortAction {
	return &AbortAction{
		Reason: msg,
	}
}

func newPrepareTaskAction(taskID string, expectedAt time.Time) *PrepareTaskAction {
	ts, _ := ptypes.TimestampProto(expectedAt)
	return &PrepareTaskAction{
		TaskID:     taskID,
		ExpectedAt: ts,
	}
}

func (m *Schedule) AddRunTask(action *RunTaskAction) {
	m.RunTasks = append(m.RunTasks, action)
}

func (m *Schedule) AddPrepareTask(action *PrepareTaskAction) {
	m.PrepareTasks = append(m.PrepareTasks, action)
}

func (m *Schedule) Actions() (actions []interface{}) {
	if m.Abort != nil {
		actions = append(actions, m.Abort)
	}
	if len(m.PrepareTasks) > 0 {
		for _, t := range m.PrepareTasks {
			actions = append(actions, t)
		}
	}
	if len(m.RunTasks) > 0 {
		for _, t := range m.RunTasks {
			actions = append(actions, t)
		}
	}
	return actions
}
