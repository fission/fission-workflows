package controller

import (
	"time"

	"context"

	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/api/invocation"
	"github.com/fission/fission-workflow/pkg/projector/project"
	"github.com/fission/fission-workflow/pkg/scheduler"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/events"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
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
	notifyChan          chan *project.InvocationNotification // TODO more complex => discard notifications of the same invocation
}

// Does not deal with Workflows (notifications)
func NewController(iproject project.InvocationProjector, wfproject project.WorkflowProjector,
	workflowScheduler *scheduler.WorkflowScheduler, functionApi *function.Api,
	invocationApi *invocation.Api) *InvocationController {
	return &InvocationController{
		invocationProjector: iproject,
		workflowProjector:   wfproject,
		scheduler:           workflowScheduler,
		functionApi:         functionApi,
		invocationApi:       invocationApi,
	}
}

// Blocking control loop
func (cr *InvocationController) Run(ctx context.Context) error {

	logrus.Debug("Running controller init...")

	// Subscribe to invocation creations
	cr.notifyChan = make(chan *project.InvocationNotification, NOTIFICATION_BUFFER)

	err := cr.invocationProjector.Subscribe(cr.notifyChan) // TODO provide clean channel that multiplexes into actual one
	if err != nil {
		panic(err)
	}

	// Invocation Notification lane
	go func(ctx context.Context) {
		for {
			select {
			case notification := <-cr.notifyChan:
				logrus.WithField("notification", notification).Info("Handling invocation notification.")
				cr.handleNotification(notification)
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

func (cr *InvocationController) handleNotification(notification *project.InvocationNotification) {
	logrus.WithField("notification", notification).Debug("controller event trigger!")
	switch notification.Type {
	case events.Invocation_INVOCATION_CREATED:
		fallthrough
	case events.Invocation_TASK_SUCCEEDED:
		fallthrough
	case events.Invocation_TASK_FAILED:
		// Decide which task to execute next
		invoc := notification.Data
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
			logrus.Errorf("Failed to schedule workflow invocation '%s': %v", notification.Id, err)
			return
		}
		logrus.WithFields(logrus.Fields{
			"schedule": schedule,
		}).Info("Scheduler decided on schedule")

		if len(schedule.Actions) == 0 { // TODO controller should verify (it is an invariant)
			output := invoc.Status.Tasks[wf.Spec.Src.OutputTask].Status.GetOutput()
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
				inputs := invokeAction.Inputs
				//inputs := map[string]string{}
				//cwd := fmt.Sprintf("$.Tasks.%s", invokeAction.Id)
				//
				//for inputKey, val := range invokeAction.Inputs {
				//
				//	// e.g. $.foo.bar
				//	// TODO make notation more consise, abstracting away the spec/metadata/status things
				//	resolvedInput, err := query.Resolve(invoc, val, cwd)
				//	if err != nil {
				//		logrus.WithFields(logrus.Fields{
				//			"val":      val,
				//			"inputKey": inputKey,
				//			"cwd":      cwd,
				//		}).Warnf("Failed to resolve input: %v", err)
				//		continue
				//	}
				//
				//	inputs[inputKey] = fmt.Sprintf("%v", resolvedInput) // TODO allow passing of actual json to functions
				//}

				// Invoke
				fnSpec := &types.FunctionInvocationSpec{
					TaskId: invokeAction.Id,
					Type:   taskDef,
					Inputs: inputs,
				}
				go func() {
					_, err := cr.functionApi.Invoke(notification.Id, fnSpec)
					if err != nil {
						logrus.WithFields(logrus.Fields{
							"id":  notification.Id,
							"err": err,
						}).Errorf("Failed to execute task")
					}
				}()
			}
		}
	default:
		logrus.WithField("type", notification.Type).Warn("Controller ignores event.")
	}
}

func (cr *InvocationController) handleControlLoopTick() {
	logrus.Debug("Controller tick...")
	// Options: refresh projection, send ping, cancel invocation
}

func (cr *InvocationController) Close() {
	logrus.Debug("Closing controller...")
	cr.invocationProjector.Close()
	close(cr.notifyChan)
}
