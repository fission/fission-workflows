package invocation

import (
	"time"

	"fmt"
	"github.com/fission/fission-workflow/pkg/cache"
	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/projector/project"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/invocationevent"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
	"strings"
)

type invocationProjector struct {
	esClient    eventstore.Client
	cache       cache.Cache // TODO ensure concurrent
	sub         eventstore.Subscription
	updateChan  chan *eventstore.Event
	subscribers []chan *project.InvocationNotification
}

func NewInvocationProjector(esClient eventstore.Client, cache cache.Cache) project.InvocationProjector {
	p := &invocationProjector{
		esClient:   esClient,
		cache:      cache,
		updateChan: make(chan *eventstore.Event),
	}
	go p.Run()
	return p
}

func (ip *invocationProjector) getCache(subject string) *types.WorkflowInvocationContainer {
	raw, ok := ip.cache.Get(subject)
	if !ok {
		return nil
	}
	invocation, ok := raw.(*types.WorkflowInvocationContainer)
	if !ok {
		logrus.Warnf("Cache contains invalid invocation '%v'. Invalidating key.", raw)
		ip.cache.Delete(subject)
	}
	return invocation
}

// Get projection from cache or attempt to replay it.
// Get should work without having to watch!
func (ip *invocationProjector) Get(subject string) (*types.WorkflowInvocationContainer, error) {
	cached := ip.getCache(subject)
	if cached != nil {
		return cached, nil
	}

	events, err := ip.esClient.Get("invocation." + subject) // TODO Fix hardcode subject
	if err != nil {
		return nil, err
	}

	var resultState *types.WorkflowInvocationContainer
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

// TODO Maybe add identifier per consumer
func (ip *invocationProjector) Subscribe(updateCh chan *project.InvocationNotification) error {
	ip.subscribers = append(ip.subscribers, updateCh)
	return nil
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
	// Close channel
	for _, ch := range ip.subscribers {
		close(ch)
	}
	return nil
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

		invocationEventType, err := invocationevent.Parse(event.Type)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"event": event,
				"types": event.Type,
				"err":   err,
			}).Warn("Failed to parse event type")
			invocationEventType = -1
		}

		// TODO should judge whether to send notification (old messages not)
		ip.notifySubscribers(&project.InvocationNotification{
			Id:   updatedState.GetMetadata().GetId(),
			Data: updatedState,
			Type: invocationEventType,
			Time: timestamp,
		})
	}
}

func (ip *invocationProjector) applyUpdate(event *eventstore.Event) (*types.WorkflowInvocationContainer, error) {
	logrus.WithField("event", event).Debug("InvocationProjector handling event.")
	invocationId := event.EventId.Subjects[1] // TODO fix hardcoded lookup

	currentState := ip.getCache(invocationId)
	if currentState == nil {
		fmt.Printf("New state! '%s'\n", invocationId)
		currentState = Initial()
	}

	newState, err := Apply(*currentState, event)
	if err != nil {
		// TODO improve error handling (e.g. retry / replay)
		return nil, err
	}
	fmt.Printf("newState: '%v' for event '%v'\n", newState, event)

	err = ip.cache.Put(invocationId, newState)
	if err != nil {
		return nil, err
	}
	return newState, nil
}

func (ip *invocationProjector) notifySubscribers(notification *project.InvocationNotification) {
	for _, c := range ip.subscribers {
		select {
		case c <- notification:
			logrus.WithField("notification", notification).Debug("Notified subscriber.")
		default:
			logrus.WithField("notification", notification).
				Debug("Failed to notify subscriber chan because of blocked channel.")
		}
	}
}
