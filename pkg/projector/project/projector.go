package project

import (
	"time"

	"io"

	"github.com/fission/fission-workflow/pkg/cache"
	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/projector/project/invocation"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/invocationevent"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

// Per object type view only!!!
type InvocationProjector interface {
	io.Closer
	// Get projection from cache or attempt to replay it.
	Get(subject string) (*types.WorkflowInvocationContainer, error)

	Cache() cache.Cache

	// Replays events, if it already exists, it is invalidated and replayed
	// Populates cache
	Fetch(subject string) error

	// Suscribe to updates in this projector
	Subscribe(updateCh chan *InvocationNotification) error
}

// In order to avoid leaking eventstore details
type InvocationNotification struct {
	Id   string
	Data *types.WorkflowInvocationContainer
	Type types.InvocationEvent
	Time time.Time
}

type invocationProjector struct {
	esClient    eventstore.Client
	cache       cache.Cache // TODO ensure concurrent
	sub         eventstore.Subscription
	updateChan  chan *eventstore.Event
	subscribers []chan *InvocationNotification
}

func NewInvocationProjector(esClient eventstore.Client, cache cache.Cache) InvocationProjector {

	updateChan := make(chan *eventstore.Event)

	projector := &invocationProjector{
		esClient:   esClient,
		cache:      cache,
		updateChan: updateChan,
	}

	go func() {
		defer logrus.Debug("InvocationProjector update loop has shutdown.")
		for event := range projector.updateChan {
			invocationId := event.EventId.Subjects[1] // TODO fix hardcoded lookup

			currentState := projector.getCache(invocationId)
			if currentState == nil {
				currentState = invocation.Initial()
			}

			newState, err := invocation.Apply(*currentState, event)
			if err != nil {
				// TODO better error handling (e.g. retry / replay)
				logrus.WithFields(logrus.Fields{
					"err":          err,
					"currentState": currentState,
					"event":        event,
				}).Error("Failed to apply event to state")
				continue
			}

			err = projector.cache.Put(invocationId, newState)
			if err != nil {
				logrus.WithFields(logrus.Fields{
					"err":          err,
					"invocationId": invocationId,
					"state":        newState,
					"event":        event,
				}).Warn("Failed to put invocation into cache.")
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

			notification := &InvocationNotification{
				Id:   invocationId,
				Data: newState,
				Type: invocationEventType,
				Time: timestamp,
			}

			// TODO should judge whether to send notification (old messages not)
			for _, c := range projector.subscribers {
				select {
				case c <- notification:
					logrus.WithField("notification", notification).Debug("Notified subscriber.")
				default:
					logrus.WithField("notification", notification).
						Debug("Failed to notify subscriber chan because of blocked channel.")
				}
			}
		}
	}()

	return projector
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
func (ip *invocationProjector) Get(subject string) (*types.WorkflowInvocationContainer, error) {
	panic("not implemented")
}

// Invalidate deletes any projection of the subject. A next get of the subject will require replaying of the events.
func (ip *invocationProjector) Invalidate(subject string) error {
	return ip.cache.Delete(subject)
}

func (ip *invocationProjector) Fetch(subject string) error {
	_, err := ip.esClient.Subscribe(&eventstore.SubscriptionConfig{
		Subject: subject,
		EventCh: ip.updateChan, // TODO provide clean channel that multiplexes into actual one
	})
	return err
}

// TODO Maybe add identifier per consumer
func (ip *invocationProjector) Subscribe(updateCh chan *InvocationNotification) error {
	ip.subscribers = append(ip.subscribers, updateCh)
	return nil
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
