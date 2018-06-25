package workflow

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/util/labels"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/golang/protobuf/ptypes"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	NotificationBuffer   = 100
	defaultEvalQueueSize = 50
	Name                 = "workflow"
)

// TODO add hard limits (cache size, max concurrent invocation)

var (
	wfLog = logrus.WithField("component", "controller.workflow")

	workflowProcessDuration = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "workflows",
		Subsystem: "controller_workflow",
		Name:      "parsed_duration",
		Help:      "Duration of a workflow from a start to a parsed state.",
	})

	workflowStatus = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "workflows",
		Subsystem: "controller_workflow",
		Name:      "status",
		Help:      "Count of the different statuses of workflows.",
	}, []string{"status"})
)

func init() {
	prometheus.MustRegister(workflowProcessDuration, workflowStatus)
}

// WorkflowController is the controller concerned with the lifecycle of workflows. It handles responsibilities, such as
// parsing of workflows.
type Controller struct {
	wfCache    fes.CacheReader
	api        *api.Workflow
	sub        *pubsub.Subscription
	cancelFn   context.CancelFunc
	evalQueue  chan string
	evalCache  *controller.EvalCache
	evalPolicy controller.Rule
}

func NewController(wfCache fes.CacheReader, wfAPI *api.Workflow) *Controller {
	ctr := &Controller{
		wfCache:   wfCache,
		api:       wfAPI,
		evalQueue: make(chan string, defaultEvalQueueSize),
		evalCache: controller.NewEvalCache(),
	}
	ctr.evalPolicy = defaultPolicy(ctr)
	return ctr
}

func (c *Controller) Init(sctx context.Context) error {
	ctx, cancelFn := context.WithCancel(sctx)
	c.cancelFn = cancelFn

	// Subscribe to invocation creations and task events.
	selector := labels.In(fes.PubSubLabelAggregateType, "workflow")
	if invokePub, ok := c.wfCache.(pubsub.Publisher); ok {
		c.sub = invokePub.Subscribe(pubsub.SubscriptionOptions{
			Buffer:       NotificationBuffer,
			LabelMatcher: selector,
		})

		// Workflow Notification listener
		go func(ctx context.Context) {
			for {
				select {
				case notification := <-c.sub.Ch:
					c.handleMsg(notification)
				case <-ctx.Done():
					wfLog.Debug("Notification listener closed.")
					return
				}
			}
		}(ctx)
	}

	// process evaluation queue
	go func(ctx context.Context) {
		for {
			select {
			case eval := <-c.evalQueue:
				controller.EvalQueueSize.WithLabelValues(Name).Dec()
				go c.Evaluate(eval) // TODO limit number of goroutines
			case <-ctx.Done():
				wfLog.Debug("Evaluation queue listener stopped.")
				return
			}
		}
	}(ctx)

	return nil
}

func (c *Controller) handleMsg(msg pubsub.Msg) error {
	wfLog.WithField("labels", msg.Labels()).Debug("Handling invocation notification.")
	switch n := msg.(type) {
	case *fes.Notification:
		c.Notify(n)
	default:
		wfLog.WithField("notification", n).Warn("Ignoring unknown notification type")
	}
	return nil
}

func (c *Controller) Tick(tick uint64) error {
	// TODO short loop: eval cache
	// TODO longer loop: cache
	return nil
}

func (c *Controller) Notify(msg *fes.Notification) error {
	wf, ok := msg.Payload.(*aggregates.Workflow)
	if !ok {
		return fmt.Errorf("received notification of invalid type '%s'. Expected '*aggregates.Workflow'", reflect.TypeOf(msg.Payload))
	}

	c.submitEval(wf.ID())
	return nil
}

func (c *Controller) Evaluate(workflowID string) {
	start := time.Now()
	// Fetch and attempt to claim the evaluation
	evalState := c.evalCache.GetOrCreate(workflowID)
	select {
	case <-evalState.Lock():
		defer evalState.Free()
	default:
		// TODO provide option to wait for a lock
		wfLog.Debugf("Failed to obtain access to workflow %s", workflowID)
		controller.EvalJobs.WithLabelValues(Name, "duplicate").Inc()
		return
	}
	wfLog.Debugf("evaluating workflow %s", workflowID)

	// Fetch the workflow relevant to the invocation
	wf := aggregates.NewWorkflow(workflowID)
	err := c.wfCache.Get(wf)
	// TODO move to rule
	if err != nil && wf.Workflow == nil {
		logrus.Errorf("controller failed to get workflow '%s': %v", workflowID, err)
		controller.EvalJobs.WithLabelValues(Name, "error").Inc()
		return
	}

	// Evaluate invocation
	record := controller.NewEvalRecord() // TODO implement rulepath + cause

	ec := NewEvalContext(evalState, wf.Workflow)

	action := c.evalPolicy.Eval(ec)
	if action == nil {
		controller.EvalJobs.WithLabelValues(Name, "noop").Inc()
		return
	}

	record.Action = action

	// Execute action
	err = action.Apply()
	if err != nil {
		wfLog.Errorf("Action '%T' failed: %v", action, err)
		record.Error = err
	}
	controller.EvalJobs.WithLabelValues(Name, "action").Inc()

	// Record this evaluation
	evalState.Record(record)

	controller.EvalDuration.WithLabelValues(Name, fmt.Sprintf("%T", action)).Observe(float64(time.Now().Sub(start)))
	if wf.GetStatus().Ready() { // TODO only once
		t, _ := ptypes.Timestamp(wf.GetMetadata().GetCreatedAt())
		workflowProcessDuration.Observe(float64(time.Now().Sub(t)))
	}
	workflowStatus.WithLabelValues(wf.GetStatus().GetStatus().String()).Inc()
}

func (c *Controller) Close() error {
	if invokePub, ok := c.wfCache.(pubsub.Publisher); ok {
		err := invokePub.Unsubscribe(c.sub)
		if err != nil {
			wfLog.Errorf("Failed to unsubscribe from workflow cache: %v", err)
		} else {
			wfLog.Info("Unsubscribed from workflow cache")
		}
	}

	c.cancelFn()
	return nil
}

func (c *Controller) submitEval(ids ...string) bool {
	for _, id := range ids {
		select {
		case c.evalQueue <- id:
			controller.EvalQueueSize.WithLabelValues(Name).Inc()
			return true
			// ok
		default:
			wfLog.Warnf("Eval queue is full; dropping eval task for '%v'", id)
			return false
		}
	}
	return true
}

func defaultPolicy(ctr *Controller) controller.Rule {
	return &controller.RuleEvalUntilAction{
		Rules: []controller.Rule{
			&RuleSkipIfReady{},
			&RuleRemoveIfDeleted{
				evalCache: ctr.evalCache,
			},
			&RuleEnsureParsed{
				WfAPI: ctr.api,
			},
		},
	}
}

//
// Workflow-specific actions
//

type ActionParseWorkflow struct {
	WfAPI *api.Workflow
	Wf    *types.Workflow
}

func (ac *ActionParseWorkflow) Apply() error {
	_, err := ac.WfAPI.Parse(ac.Wf)
	return err
}
