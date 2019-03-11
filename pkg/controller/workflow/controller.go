package workflow

import (
	"context"
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/api/store"
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/util/gopool"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/golang/protobuf/ptypes"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	defaultEvalQueueSize  = 50
	Name                  = "workflow"
	maxParallelExecutions = 100
)

// TODO add hard limits (cache size, max concurrent invocation)

var (
	log = logrus.WithField("component", "controller.workflow")

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
	workflows  *store.Workflows
	api        *api.Workflow
	sub        *pubsub.Subscription
	cancelFn   context.CancelFunc
	evalQueue  chan string
	evalCache  *controller.EvalStore
	evalPolicy controller.Rule
}

func NewController(workflows *store.Workflows, api *api.Workflow) *Controller {
	ctr := &Controller{
		workflows: workflows,
		api:       api,
		evalQueue: make(chan string, defaultEvalQueueSize),
		evalCache: &controller.EvalStore{},
	}
	ctr.evalPolicy = defaultPolicy(ctr)
	return ctr
}

func (c *Controller) Init(sctx context.Context) error {
	ctx, cancelFn := context.WithCancel(sctx)
	c.cancelFn = cancelFn

	// Subscribe to invocation creations and task events.
	go func(ctx context.Context) {
		sub := c.workflows.GetWorkflowUpdates()
		if sub == nil {
			log.Warn("Workflow cache does not support pubsub.")
			return
		}
		for {
			select {
			case msg := <-sub.Ch:
				update, err := sub.ToNotification(msg)
				if err != nil {
					log.Warnf("Failed to convert pubsub message to notification: %v", err)
				}
				err = c.Notify(update)
				if err != nil {
					log.Errorf("Failed to notify controller of update: %v", err)
				}
			case <-ctx.Done():
				err := sub.Close()
				if err != nil {
					log.Error(err)
				}
				log.Info("Notification listener stopped.")
				return
			}
		}
	}(ctx)

	// process evaluation queue
	pool := gopool.New(maxParallelExecutions)
	go func(ctx context.Context) {
		for {
			select {
			case eval := <-c.evalQueue:
				pool.Submit(ctx, func() {
					controller.EvalQueueSize.WithLabelValues(Name).Dec()
					c.Evaluate(eval)
				})
			case <-ctx.Done():
				log.Debug("Evaluation queue listener stopped.")
				return
			}
		}
	}(ctx)

	return nil
}

func (c *Controller) Tick(tick uint64) error {
	// TODO short loop: eval cache
	// TODO longer loop: cache
	return nil
}

func (c *Controller) Notify(notification *fes.Notification) error {
	workflow, err := store.ParseNotificationToWorkflow(notification)
	if err != nil {
		return err
	}
	spanCtx, err := fes.ExtractTracingFromEvent(notification.Event)
	if err != nil {
		logrus.Warnf("Failed to extract opentracing metadata from event %v", notification.Event.Id)
	}
	c.evalCache.LoadOrStore(workflow.ID(), spanCtx)
	c.submitEval(workflow.ID())
	return nil
}

func (c *Controller) Evaluate(workflowID string) {
	start := time.Now()
	// Fetch and attempt to claim the evaluation
	evalState, ok := c.evalCache.Load(workflowID)
	if !ok {
		logrus.Warnf("Skipping evaluation of unknown workflow: %v", workflowID)
		return
	}
	select {
	case <-evalState.Lock():
		defer evalState.Free()
	default:
		// TODO provide option to wait for a lock
		log.Debugf("Failed to obtain access to workflow %s", workflowID)
		controller.EvalJobs.WithLabelValues(Name, "duplicate").Inc()
		return
	}
	log.Debugf("evaluating workflow %s", workflowID)

	// Fetch the workflow relevant to the invocation
	wf, err := c.workflows.GetWorkflow(workflowID)
	// TODO move to rule
	if err != nil && wf == nil {
		logrus.Errorf("controller failed to get workflow '%s': %v", workflowID, err)
		controller.EvalJobs.WithLabelValues(Name, "error").Inc()
		return
	}

	// Evaluate invocation
	record := controller.NewEvalRecord() // TODO implement rulepath + cause

	ec := NewEvalContext(evalState, wf)

	actions := c.evalPolicy.Eval(ec)
	if actions == nil {
		controller.EvalJobs.WithLabelValues(Name, "noop").Inc()
		return
	}

	// Execute action
	for _, action := range actions {
		err = action.Apply()
		if err != nil {
			log.Errorf("Action '%T' failed: %v", action, err)
			record.Error = err
			break
		}
	}
	controller.EvalJobs.WithLabelValues(Name, "action").Inc()

	// Record this evaluation
	evalState.Record(record)

	controller.EvalDuration.WithLabelValues(Name, fmt.Sprintf("%T", actions)).Observe(float64(time.Now().Sub(start)))
	if wf.GetStatus().Ready() { // TODO only once
		t, _ := ptypes.Timestamp(wf.GetMetadata().GetCreatedAt())
		workflowProcessDuration.Observe(float64(time.Now().Sub(t)))
	}
	workflowStatus.WithLabelValues(wf.GetStatus().GetStatus().String()).Inc()
}

func (c *Controller) Close() error {
	err := c.evalCache.Close()
	c.cancelFn()
	return err
}

func (c *Controller) submitEval(ids ...string) bool {
	for _, id := range ids {
		select {
		case c.evalQueue <- id:
			controller.EvalQueueSize.WithLabelValues(Name).Inc()
			return true
			// ok
		default:
			log.Warnf("Eval queue is full; dropping eval task for '%v'", id)
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
