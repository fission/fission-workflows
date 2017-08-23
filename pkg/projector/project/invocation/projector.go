package invocation

import (
	"time"

	"strings"

	"github.com/fission/fission-workflow/pkg/cache"
	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/projector/project"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/events"
	"github.com/fission/fission-workflow/pkg/util/labels/kubelabels"
	"github.com/fission/fission-workflow/pkg/util/pubsub"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

type invocationProjector struct {
	pubsub.Publisher
	esClient   eventstore.Client
	cache      cache.Cache // TODO ensure concurrent
	sub        eventstore.Subscription
	updateChan chan *eventstore.Event
}

func NewInvocationProjector(esClient eventstore.Client, cache cache.Cache) project.InvocationProjector {
	p := &invocationProjector{
		Publisher:  pubsub.NewPublisher(),
		esClient:   esClient,
		cache:      cache,
		updateChan: make(chan *eventstore.Event),
	}
	go p.Run()
	return p
}

func (ip *invocationProjector) getCache(subject string) *types.WorkflowInvocation {
	raw, ok := ip.cache.Get(subject)
	if !ok {
		return nil
	}
	invocation, ok := raw.(*types.WorkflowInvocation)
	if !ok {
		logrus.Warnf("Cache contains invalid invocation '%v'. Invalidating key.", raw)
		ip.cache.Delete(subject)
	}
	return invocation
}

// Get projection from cache or attempt to replay it.
// Get should work without having to watch!
func (ip *invocationProjector) Get(subject string) (*types.WorkflowInvocation, error) {
	cached := ip.getCache(subject)
	if cached != nil {
		return cached, nil
	}

	events, err := ip.esClient.Get("invocation." + subject) // TODO Fix hardcode subject
	if err != nil {
		return nil, err
	}

	var resultState *types.WorkflowInvocation
	for _, event := range events {
		updatedState, err := ip.applyUpdate(event)
		if err != nil {
			logrus.Error(err)
		}
		resultState = updatedState
	}

	return resultState, nil
}

// Invalidate deletes any projection of the subject. A next get of the subject will require replaying of the events.
func (ip *invocationProjector) Invalidate(subject string) error {
	return ip.cache.Delete(subject)
}

func (ip *invocationProjector) Watch(subject string) error {
	_, err := ip.esClient.Subscribe(&eventstore.SubscriptionConfig{
		Subject: subject,
		EventCh: ip.updateChan, // TODO provide clean channel that multiplexes into actual one
	})
	return err
}

func (ip *invocationProjector) List(query string) ([]string, error) {
	subjects, err := ip.esClient.Subjects("invocation." + query) // TODO fix this hardcode
	if err != nil {
		return nil, err
	}
	results := make([]string, len(subjects))
	for key, subject := range subjects {
		results[key] = strings.SplitN(subject, ".", 2)[1] // TODO fix this hardcode
	}
	return results, nil
}

func (ip *invocationProjector) Cache() cache.Cache {
	return ip.cache
}

func (ip *invocationProjector) Close() error {
	// Note: subscribers are responsible for closing their own channels
	return ip.Publisher.Close()
}

func (ip *invocationProjector) Run() {
	defer logrus.Debug("InvocationProjector update loop has shutdown.")
	for event := range ip.updateChan {
		updatedState, err := ip.applyUpdate(event)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":          err,
				"updatedState": updatedState,
				"event":        event,
			}).Error("Failed to apply event to state")
		}

		timestamp, err := ptypes.Timestamp(event.Time)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"event": event,
			}).Warn("Could not parse timestamp, using fallback time.")
			timestamp = time.Now()
		}

		t, err := events.ParseInvocation(event.Type)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"event": event,
				"types": event.Type,
				"err":   err,
			}).Warn("Failed to parse event type")
			t = -1
		}

		// TODO should judge whether to send notification (old messages not)
		ip.Publisher.Publish(NewNotification(t, timestamp, updatedState))
	}
}

func (ip *invocationProjector) applyUpdate(event *eventstore.Event) (*types.WorkflowInvocation, error) {
	logrus.WithField("event", event).Debug("InvocationProjector handling event.")
	invocId := event.EventId.Subjects[1] // TODO fix hardcoded lookup

	currentState := ip.getCache(invocId)
	if currentState == nil {
		currentState = Initial()
	}

	newState, err := Apply(*currentState, event)
	if err != nil {
		// TODO improve error handling (e.g. retry / replay)
		return nil, err
	}

	err = ip.cache.Put(invocId, newState)
	if err != nil {
		return nil, err
	}
	return newState, nil
}

type Notification struct {
	*pubsub.EmptyMsg
	Payload *types.WorkflowInvocation
}

func (nf *Notification) Event() events.Invocation {
	return events.Invocation(events.Invocation_value[nf.Labels().Get("event")])
}

func NewNotification(event events.Invocation, timestamp time.Time, invoc *types.WorkflowInvocation) *Notification {
	return &Notification{
		EmptyMsg: pubsub.NewEmptyMsg(kubelabels.New(kubelabels.LabelSet{
			"event": event.String(),
		}), timestamp),
		Payload: invoc,
	}
}
