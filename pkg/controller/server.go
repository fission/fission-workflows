package controller

import (
	"time"

	"context"

	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/api/invocation"
	"github.com/fission/fission-workflow/pkg/controller/query"
	"github.com/fission/fission-workflow/pkg/projector/project"
	invocproject "github.com/fission/fission-workflow/pkg/projector/project/invocation"
	"github.com/fission/fission-workflow/pkg/scheduler"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/events"
	"github.com/fission/fission-workflow/pkg/util/labels/kubelabels"
	"github.com/fission/fission-workflow/pkg/util/pubsub"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

const (
	NOTIFICATION_BUFFER = 100
	TICK_SPEED          = time.Duration(10) * time.Second
)

type InvocationController struct {
	invocationProjector project.InvocationProjector
	workflowProjector   project.WorkflowProjector
	functionApi         *function.Api
	invocationApi       *invocation.Api
	scheduler           *scheduler.WorkflowScheduler
	invocSub            *pubsub.Subscription
	exprParser          query.ExpressionParser
}

// Does not deal with Workflows (notifications)
func NewController(iproject project.InvocationProjector, wfproject project.WorkflowProjector,
	workflowScheduler *scheduler.WorkflowScheduler, functionApi *function.Api,
	invocationApi *invocation.Api, exprParser query.ExpressionParser) *InvocationController {
	return &InvocationController{
		invocationProjector: iproject,
		workflowProjector:   wfproject,
		scheduler:           workflowScheduler,
		functionApi:         functionApi,
		invocationApi:       invocationApi,
		exprParser:          exprParser,
	}
}

// Blocking control loop
func (cr *InvocationController) Run(ctx context.Context) error {

	logrus.Debug("Running controller init...")

	// Subscribe to invocation creations and task events.
	req, err := labels.NewRequirement("event", selection.In, []string{
		events.Invocation_TASK_SUCCEEDED.String(),
		events.Invocation_TASK_FAILED.String(),
		events.Invocation_INVOCATION_CREATED.String(),
	})
	if err != nil {
		return err
	}
	selector := kubelabels.NewSelector(labels.NewSelector().Add(*req))

	cr.invocSub = cr.invocationProjector.Subscribe(pubsub.SubscriptionOptions{
		Buf:           NOTIFICATION_BUFFER,
		LabelSelector: selector,
	})
	if err != nil {
		panic(err)
	}

	// Invocation Notification lane
	go func(ctx context.Context) {
		for {
			select {
			case notification := <-cr.invocSub.Ch:
				logrus.WithField("notification", notification).Info("Handling invocation notification.")
				switch n := notification.(type) {
				case *invocproject.Notification:
					cr.handleNotification(n)
				default:
					logrus.WithField("notification", n).Warn("Ignoring unknown notification type")
				}
			case <-ctx.Done():
				logrus.WithField("ctx.err", ctx.Err()).Debug("Notification listener closed.")
				return
			}
		}
	}(ctx)

	// Control lane
	logrus.Debug("Init done. Entering control loop.")
	ticker := time.NewTicker(TICK_SPEED)
	for {
		select {
		case <-ctx.Done():
			return err
		case <-ticker.C:
			cr.handleControlLoopTick()
		}
	}
}

func (cr *InvocationController) handleNotification(msg *invocproject.Notification) {
	logrus.WithField("notification", msg).Debug("controller event trigger!")

	switch msg.Event() {
	case events.Invocation_INVOCATION_CREATED:
		fallthrough
	case events.Invocation_TASK_SUCCEEDED:
		fallthrough
	case events.Invocation_TASK_FAILED:
		// Decide which task to execute next
		invoc := msg.Payload

		wfId := invoc.Spec.WorkflowId
		wf, err := cr.workflowProjector.Get(wfId)
		if err != nil {
			logrus.Errorf("Failed to get workflow for invocation '%s': %v", wfId, err)
			return
		}

		schedule, err := cr.scheduler.Evaluate(&scheduler.ScheduleRequest{
			Invocation: invoc,
			Workflow:   wf,
		})
		if err != nil {
			logrus.Errorf("Failed to schedule workflow invocation '%s': %v", invoc.Metadata.Id, err)
			return
		}
		logrus.WithFields(logrus.Fields{
			"schedule": schedule,
		}).Info("Scheduler decided on schedule")

		if len(schedule.Actions) == 0 { // TODO controller should verify (it is an invariant)
			output := invoc.Status.Tasks[wf.Spec.OutputTask].Status.GetOutput()
			err := cr.invocationApi.Success(invoc.Metadata.Id, output)
			if err != nil {
				panic(err)
			}
		}

		for _, action := range schedule.Actions {

			switch action.Type {
			case scheduler.ActionType_ABORT:
				logrus.Infof("aborting: '%v'", action)
				cr.invocationApi.Cancel(invoc.Metadata.Id)
			case scheduler.ActionType_INVOKE_TASK:
				invokeAction := &scheduler.InvokeTaskAction{}
				ptypes.UnmarshalAny(action.Payload, invokeAction)

				// Resolve type of the task
				taskDef, _ := wf.Status.ResolvedTasks[invokeAction.Id]

				// Resolve the inputs
				inputs := map[string]*types.TypedValue{}
				queryScope := query.NewScope(wf, invoc)
				for inputKey, val := range invokeAction.Inputs {
					resolvedInput, err := cr.exprParser.Resolve(queryScope, queryScope.Tasks[invokeAction.Id], val)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"val":      val,
							"inputKey": inputKey,
						}).Warnf("Failed to parse input: %v", err)
						continue
					}

					inputs[inputKey] = resolvedInput
					logrus.WithFields(logrus.Fields{
						"val":      val,
						"key":      inputKey,
						"resolved": resolvedInput,
					}).Infof("Resolved expression")
				}

				// Invoke
				fnSpec := &types.FunctionInvocationSpec{
					TaskId: invokeAction.Id,
					Type:   taskDef,
					Inputs: inputs,
				}
				go func() {
					_, err := cr.functionApi.Invoke(invoc.Metadata.Id, fnSpec)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"id":  invoc.Metadata.Id,
							"err": err,
						}).Errorf("Failed to execute task")
					}
				}()
			}
		}
	default:
		logrus.WithField("type", msg.Event().String()).Warn("Controller ignores unknown event.")
	}
}

func (cr *InvocationController) handleControlLoopTick() {
	logrus.Debug("Controller tick...")
	// Options: refresh projection, send ping, cancel invocation
}

func (cr *InvocationController) Close() error {
	logrus.Debug("Closing controller...")
	return cr.invocationProjector.Close()
}
