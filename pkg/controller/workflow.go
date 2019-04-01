package controller

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/api/store"
	"github.com/fission/fission-workflows/pkg/controller/ctrl"
	"github.com/fission/fission-workflows/pkg/controller/executor"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
)

const (
	parseTask = "task"
)

// WorkflowController is the controller for ensuring the processing of a single workflow.
type WorkflowController struct {
	api        *api.Workflow
	executor   *executor.LocalExecutor
	errorCount int
	workflowID string
}

func NewWorkflowController(api *api.Workflow, executor *executor.LocalExecutor, workflowID string) *WorkflowController {
	return &WorkflowController{api: api, executor: executor, workflowID: workflowID}
}

func (c *WorkflowController) Eval(ctx context.Context, processValue *ctrl.Event) ctrl.Result {
	workflow, ok := processValue.Updated.(*types.Workflow)
	if !ok {
		return ctrl.Err{Err: fmt.Errorf("entity expected %T, but was %T", &types.Workflow{}, processValue.Updated)}
	}

	// ensure that it is the correct workflow
	if workflow.ID() != c.workflowID {
		return ctrl.Err{Err: fmt.Errorf("workflow ID expected %v, but was %v", c.workflowID, workflow.ID())}
	}

	// Do not evaluate as long as there still tasks to be executed
	if c.executor.GetGroupTasks(workflow.ID()) > 0 {
		return ctrl.Err{Err: errors.New("still executing tasks for workflow")}
	}

	switch workflow.GetStatus().GetStatus() {
	case types.WorkflowStatus_READY, types.WorkflowStatus_DELETED:
		// Stop tracking this workflow because it has reached a terminal state
		return ctrl.Done{}
	case types.WorkflowStatus_FAILED:
		// The previous parsing has failed. We retry the parsing but with a increasing backoff.

		// TODO revert back to retrying once Fission can deal with specialization timeouts
		//backoff := time.Duration(c.errorCount) * time.Second
		//log.Infof("Backing off for %v before trying to parse workflow again", backoff)
		//c.executor.SubmitAfter(&executor.Task{
		//	TaskID:  workflow.ID() + "." + parseTask,
		//	GroupID: workflow.ID(),
		//	Apply: func() error {
		//		_, err := c.api.Parse(workflow)
		//		return err
		//	},
		//}, backoff)
		//return ctrl.Success{Msg: "retrying parsing of the workflow"}
		return ctrl.Done{}
	case types.WorkflowStatus_QUEUED:
		// The workflow has not yet been processed, so we try to parse it.
		c.errorCount = 0
		c.executor.Submit(&executor.Task{
			TaskID:  workflow.ID() + "." + parseTask,
			GroupID: workflow.ID(),
			Apply: func() error {
				_, err := c.api.Parse(workflow)
				return err
			},
		})
		return ctrl.Success{Msg: "parsing of the workflow"}
	default:
		return ctrl.Err{Err: errors.New("unknown workflow state")}
	}
}

// WorkflowMetaController is the component responsible for the full integration of the workflows reconciliation loop.
//
// Specifically, the meta-controller is responsible for the following:
// - It starts or registers the sensors to ensure that events are routed to this control system.
// - It manages all of the workflow controllers.
// - It provides an executor pool for controllers to submit their tasks to.
type WorkflowMetaController struct {
	system    *ctrl.System
	api       *api.Workflow
	executor  *executor.LocalExecutor
	workflows *store.Workflows
	run       *sync.Once
	sensors   []ctrl.Sensor
}

func NewWorkflowMetaController(api *api.Workflow, workflows *store.Workflows, executor *executor.LocalExecutor,
	storePollInterval time.Duration) *WorkflowMetaController {

	return &WorkflowMetaController{
		api:       api,
		executor:  executor,
		run:       &sync.Once{},
		workflows: workflows,
		sensors: []ctrl.Sensor{
			NewWorkflowNotificationSensor(workflows),
			NewWorkflowStorePollSensor(workflows, storePollInterval),
		},
		system: ctrl.NewSystem(func(event *ctrl.Event) (ctrl ctrl.Controller, err error) {
			return NewWorkflowController(api, executor, event.Aggregate.Id), nil
		}),
	}
}

func (c *WorkflowMetaController) Run() {
	c.run.Do(func() {
		// Start the task executor
		c.executor.Start()

		// Start the sensors
		for _, sensor := range c.sensors {
			err := sensor.Start(c.system)
			if err != nil {
				panic(err)
			}
		}

		// Run control system
		c.system.Run()
	})
}

func (c *WorkflowMetaController) Close() error {
	err := c.executor.Close()
	err = c.system.Close()
	for _, sensor := range c.sensors {
		err = sensor.Close()
	}
	return err
}

// WorkflowNotificationSensor watches the workflow store notifications for workflow events.
type WorkflowNotificationSensor struct {
	workflows *store.Workflows
	done      func()
	closeC    <-chan struct{}
}

func NewWorkflowNotificationSensor(workflows *store.Workflows) *WorkflowNotificationSensor {
	ctx, done := context.WithCancel(context.Background())
	return &WorkflowNotificationSensor{
		workflows: workflows,
		done:      done,
		closeC:    ctx.Done(),
	}
}

func (s *WorkflowNotificationSensor) Start(evalQueue ctrl.EvalQueue) error {
	go s.Run(evalQueue)
	return nil
}

func (s *WorkflowNotificationSensor) Run(evalQueue ctrl.EvalQueue) {
	sub := s.workflows.GetWorkflowUpdates()
	if sub == nil {
		log.Warn("Workflow store does not support pubsub.")
		return
	}
	log.Debug("Listening for workflow events")
	for {
		select {
		case msg := <-sub.Ch:
			notification, err := sub.ToNotification(msg)
			if err != nil {
				log.Warnf("Failed to convert pubsub message to notification: %v", err)
			}
			evalQueue.Submit(notification)
		case <-s.closeC:
			err := sub.Close()
			if err != nil {
				log.Error(err)
			}
			log.Info("Notification listener stopped.")
			return
		}
	}
}

func (s *WorkflowNotificationSensor) Close() error {
	s.done()
	return nil
}

// WorkflowStorePollSensor polls the workflows store on a set interval.
type WorkflowStorePollSensor struct {
	*ctrl.PollSensor
	workflows *store.Workflows
}

func NewWorkflowStorePollSensor(workflows *store.Workflows, interval time.Duration) *WorkflowStorePollSensor {
	s := &WorkflowStorePollSensor{
		workflows: workflows,
	}
	s.PollSensor = ctrl.NewPollSensor(interval, s.Poll)
	return s
}

func (s *WorkflowStorePollSensor) Poll(evalQueue ctrl.EvalQueue) {
	for _, aggregate := range s.workflows.List() {
		// Ignore non-workflow entities in workflow store
		if aggregate.Type != types.TypeWorkflow {
			log.Warnf("Non-workflow entity in workflows store: %v", aggregate)
			continue
		}

		// Get actual workflow
		wf, err := s.workflows.GetWorkflow(aggregate.GetId())
		if err != nil {
			log.Warnf("Could not retrieve entity from workflows store: %v", aggregate)
			continue
		}

		// Check if the status is not in a terminal state
		switch wf.GetStatus().GetStatus() {
		case types.WorkflowStatus_DELETED, types.WorkflowStatus_READY:
			continue
		default:
			// nop
		}
		log.Debugf("WorkflowStorePoll: Found missing workflow %v (state: %v)", aggregate.GetId(),
			wf.GetStatus().GetStatus().String())

		// Submit evaluation for the workflow
		// The workqueue within in the control system ensures that workflows that are already queued for execution
		// will be ignored.
		evalQueue.Submit(&ctrl.Event{
			Old:     wf,
			Updated: wf,
			Event: &fes.Event{
				Type:      "recover-workflow",
				Aggregate: &aggregate,
				Timestamp: ptypes.TimestampNow(),
			},
		})
	}
}
