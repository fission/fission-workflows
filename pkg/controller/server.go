package controller

import (
	"fmt"
	"time"

	"github.com/fission/fission-workflow/pkg/api"
	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/eventstore/eventids"
	"github.com/fission/fission-workflow/pkg/eventstore/events"
	"github.com/fission/fission-workflow/pkg/projector/project"
	"github.com/fission/fission-workflow/pkg/scheduler"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/util"
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
	functionApi         function.Api
	invocationApi       *api.InvocationApi
	esClient            eventstore.Client
	scheduler           *scheduler.WorkflowScheduler
	notifyChan          chan *project.InvocationNotification // TODO more complex => discard notifications of the same invocation
}

// Does not deal with Workflows (notifications)
func NewController(iproject project.InvocationProjector, wfproject project.WorkflowProjector,
	workflowScheduler *scheduler.WorkflowScheduler, functionApi function.Api, invocationApi *api.InvocationApi,
	esClient eventstore.Client) *InvocationController {
	return &InvocationController{
		invocationProjector: iproject,
		workflowProjector:   wfproject,
		scheduler:           workflowScheduler,
		functionApi:         functionApi,
		esClient:            esClient,
		invocationApi:       invocationApi,
	}
}

// Blocking control loop
func (cr *InvocationController) Run() {

	logrus.Debug("Running controller init...")

	// Subscribe to invocation creations
	cr.notifyChan = make(chan *project.InvocationNotification, NOTIFICATION_BUFFER)

	err := cr.invocationProjector.Subscribe(cr.notifyChan) // TODO provide clean channel that multiplexes into actual one
	if err != nil {
		panic(err)
	}

	// Invocation Notification lane
	go func() {
		for {
			notification := <-cr.notifyChan
			logrus.WithField("notification", notification).Info("Handling invocation notification.")
			cr.handleNotification(notification)
		}
	}()

	// Control lane
	logrus.Debug("Init done. Entering control loop.")
	ticker := time.NewTicker(TICK_SPEED)
	for {
		<-ticker.C
		cr.handleControlLoopTick()
	}
}

func (cr *InvocationController) handleNotification(notification *project.InvocationNotification) {
	println("controller event trigger!")
	switch notification.Type {
	case types.InvocationEvent_INVOCATION_CREATED:
		fallthrough
	case types.InvocationEvent_TASK_SUCCEEDED:
		fallthrough
	case types.InvocationEvent_TASK_FAILED:
		// Check if done

		// Decide which task to execute next
		wfId := notification.Data.Spec.WorkflowId
		wf, err := cr.workflowProjector.Get(wfId)
		if err != nil {
			logrus.Errorf("Failed to get workflow for invocation '%s': %v", wfId, err)
			return
		}

		schedule, err := cr.scheduler.Evaluate(&scheduler.ScheduleRequest{
			Invocation: notification.Data,
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
			err := cr.invocationApi.Success(notification.Data.Metadata.Id)
			if err != nil {
				panic(err)
			}
		}

		for _, action := range schedule.Actions {

			switch action.Type {
			case scheduler.ActionType_ABORT:
				logrus.Infof("aborting: '%v'", action)
				cr.invocationApi.Cancel(notification.Data.Metadata.Id)
			case scheduler.ActionType_INVOKE_TASK:
				invokeAction := &scheduler.InvokeTaskAction{}
				ptypes.UnmarshalAny(action.Payload, invokeAction)

				//go func() { // TODO move to functionproxy/execution/manager/api
				task, _ := wf.Status.ResolvedTasks[invokeAction.Id]
				fnSpec := &types.FunctionInvocationSpec{
					TaskId:       invokeAction.Id,
					FunctionId:   task.Resolved,
					FunctionName: task.Src,
				}
				fn := &types.FunctionInvocation{
					Spec: fnSpec,
					Metadata: &types.ObjectMetadata{
						Id:        util.Uid(),
						CreatedAt: ptypes.TimestampNow(),
					},
				}

				fnAny, err := ptypes.MarshalAny(fn)
				if err != nil {
					panic(err)
				}

				eventid := eventids.NewSubject(types.SUBJECT_INVOCATION, notification.Id)

				startEvent := events.New(eventid, types.InvocationEvent_TASK_STARTED.String(), fnAny)

				err = cr.esClient.Append(startEvent)
				if err != nil {
					panic(err)
				}

				fnInvoke, err := cr.functionApi.InvokeSync(fnSpec)

				if err != nil {
					fnFailedAny, err := ptypes.MarshalAny(&types.FunctionInvocation{
						Spec: fnSpec,
					})
					if err != nil {
						panic(err)
					}
					logrus.WithFields(logrus.Fields{
						"err":  err,
						"spec": fnSpec,
					}).Errorf("Failed to execute function")
					failedEvent := events.New(eventid, types.InvocationEvent_TASK_FAILED.String(), fnFailedAny)
					err = cr.esClient.Append(failedEvent)
					if err != nil {
						panic(err)
					}
					return
				}

				fnInvokeAny, err := ptypes.MarshalAny(fnInvoke)
				if err != nil {
					panic(err)
				}

				succeededEvent := events.New(eventid, types.InvocationEvent_TASK_SUCCEEDED.String(), fnInvokeAny)
				err = cr.esClient.Append(succeededEvent)
				if err != nil {
					panic(err)
				}
				logrus.WithField("fnInvoke", fnInvoke).Info("Function execution succeeded")
				//}()
			}
		}
	default:
		fmt.Printf("I do not know what to do on this event: %s\n", notification.Type)
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
