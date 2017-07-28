package controller

import (
	"fmt"
	"time"

	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/projector/project"
	"github.com/fission/fission-workflow/pkg/scheduler"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/invocationevent"
	"github.com/sirupsen/logrus"
)

const (
	NOTIFICATION_BUFFER = 100
	TICK_SPEED          = time.Duration(10) * time.Second
	INVOCATION_SUBJECT  = "invocation"
)

type InvocationController struct {
	invocationProjector project.InvocationProjector
	scheduler           *scheduler.WorkflowScheduler
	notifyChan          chan *project.InvocationNotification // TODO more complex => discard notifications of the same invocation
}

// Does not deal with Workflows (notifications)
func NewController(projector project.InvocationProjector,
	workflowScheduler *scheduler.WorkflowScheduler) *InvocationController {
	return &InvocationController{
		invocationProjector: projector,
		scheduler:           workflowScheduler,
	}
}

// Blocking control loop
func (cr *InvocationController) Run() {

	logrus.Debug("Running controller init...")

	// Subscribe to invocation creations
	cr.notifyChan = make(chan *project.InvocationNotification, NOTIFICATION_BUFFER)
	err := cr.invocationProjector.Watch(fmt.Sprintf("%s.>", INVOCATION_SUBJECT))
	if err != nil {
		panic(err)
	}
	err = cr.invocationProjector.Subscribe(cr.notifyChan) // TODO provide clean channel that multiplexes into actual one
	if err != nil {
		panic(err)
	}

	logrus.Debug("Init done. Entering control loop.")

	ticker := time.NewTicker(TICK_SPEED)
	go func() { // Notification lane
		for {
			notification := <-cr.notifyChan
			logrus.WithField("notification", notification).Info("Handling notification.")
			cr.handleNotification(notification)
		}
	}()
	for { // Control lane
		<-ticker.C
		cr.handleControlLoopTick()
	}
}

func (cr *InvocationController) onInvocationEvent(container *types.WorkflowInvocationContainer, cause *eventstore.Event) {
	println("controller event trigger!")
	flag, _ := invocationevent.Parse(cause.GetType())
	switch flag {
	case types.InvocationEvent_INVOCATION_CREATED:
		fallthrough
	case types.InvocationEvent_TASK_FAILED:
		fallthrough
	case types.InvocationEvent_TASK_SUCCEEDED:
		fallthrough
	case types.InvocationEvent_TASK_HEARTBEAT:
		fmt.Printf("Should evaluate now!\n")
	default:
		fmt.Printf("Doing nothing! (%v) (%v)\n", container, cause)
	}
}

func (cr *InvocationController) handleNotification(notification *project.InvocationNotification) {
	println("controller event trigger!")
	switch notification.Type {
	case types.InvocationEvent_INVOCATION_CREATED:
		fallthrough
	case types.InvocationEvent_TASK_FAILED:
		fallthrough
	case types.InvocationEvent_TASK_SUCCEEDED:
		fallthrough
	case types.InvocationEvent_TASK_HEARTBEAT:
		fmt.Printf("Should evaluate now!\n")
	default:
		fmt.Printf("Doing nothing! (%v)\n", notification)
	}
}

func (cr *InvocationController) handleControlLoopTick() {
	logrus.Debug("Running controller control loop tick")
	// Options: refresh projection, send ping, cancel invocation
	logrus.Debug("cache: %v", cr.invocationProjector.Cache().List())
}

func (cr *InvocationController) Close() {
	logrus.Debug("Closing controller...")
	cr.invocationProjector.Close()
	close(cr.notifyChan)
}
