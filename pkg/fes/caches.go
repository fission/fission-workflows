package fes

import (
	"context"
	"errors"

	"fmt"

	"github.com/fission/fission-workflow/pkg/util/pubsub"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
	"time"
)

// MapCache provides a simple map-based CacheReaderWriter implementation
type MapCache struct {
	contents map[string]map[string]Aggregator // Map: AggregateType -> AggregateId -> entity
}

func NewMapCache() *MapCache {
	return &MapCache{
		contents: map[string]map[string]Aggregator{},
	}
}

func (rc *MapCache) Get(entity Aggregator) error {
	if entity == nil {
		return errors.New("entity is nil")
	}

	ref := entity.Aggregate()
	err := validateAggregate(ref)
	if err != nil {
		return err
	}

	cached, err := rc.GetAggregate(ref)
	if err != nil {
		return err
	}

	if cached == nil {
		//panic(fmt.Sprintf("%v", entity.Aggregate()))
		return fmt.Errorf("entity '%v' not found", entity.Aggregate())
	}

	e := entity.UpdateState(cached)

	return e
}

func (rc *MapCache) GetAggregate(aggregate Aggregate) (Aggregator, error) {
	err := validateAggregate(aggregate)
	if err != nil {
		return nil, err
	}

	aType, ok := rc.contents[aggregate.Type]
	if !ok {
		return nil, nil
	}

	cached, ok := aType[aggregate.Id]
	if !ok {
		return nil, nil
	}

	return cached, nil
}

func (rc *MapCache) Put(entity Aggregator) error {
	ref := entity.Aggregate()
	err := validateAggregate(ref)
	if err != nil {
		return err
	}

	if _, ok := rc.contents[ref.Type]; !ok {
		rc.contents[ref.Type] = map[string]Aggregator{}
	}

	rc.contents[ref.Type][ref.Id] = entity
	return nil
}

func (rc *MapCache) List() []Aggregate {
	results := []Aggregate{}
	for atype := range rc.contents {
		for _, entity := range rc.contents[atype] {
			results = append(results, entity.Aggregate())
		}
	}
	return results
}

// FallbackCache looks up missing entity in a backing event store in case of a cache miss.
//
// TODO implement GetAggregate, List
type FallbackCache struct {
	CacheReaderWriter
	es     EventStore
	target func() Aggregator
}

func NewFallbackCache(cache CacheReaderWriter, es EventStore, target func() Aggregator) *FallbackCache {
	return &FallbackCache{
		CacheReaderWriter: cache,
		es:                es,
		target:            target,
	}
}

func (c *FallbackCache) Get(entity Aggregator) error {
	if entity == nil {
		return errors.New("entity is nil")
	}

	ref := entity.Aggregate()
	err := validateAggregate(ref)

	cached, err := c.CacheReaderWriter.GetAggregate(ref)
	if err != nil {
		return err
	}
	if cached == nil {
		events, err := c.es.Get(&ref)
		if err != nil {
			return err
		}
		cached = c.target()
		err = Project(cached, events...)
		if err != nil {
			return err
		}
	}

	return entity.UpdateState(cached)
}

// A SubscribedCache is subscribed to some event emitter,
//
// TODO add fallback to query store in case of aggregate error or missing
type SubscribedCache struct {
	pubsub.Publisher
	CacheReaderWriter
	ts     time.Time
	target func() Aggregator // TODO extract to a TypedSubscription
}

func NewSubscribedCache(ctx context.Context, cache CacheReaderWriter, target func() Aggregator, sub *pubsub.Subscription) *SubscribedCache {
	c := &SubscribedCache{
		Publisher:         pubsub.NewPublisher(),
		CacheReaderWriter: cache,
		target:            target,
		ts:                time.Now(),
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case msg := <-sub.Ch:
				// Discard invalid messages
				event, ok := msg.(*Event)
				if !ok {
					logrus.WithField("msg", msg).Error("Received a malformed message")
					continue
				}
				err := c.HandleEvent(event)
				if err != nil {
					logrus.WithField("err", err).Error("Failed to handle event")
				}
			}
		}
	}()

	return c
}

func (uc *SubscribedCache) HandleEvent(event *Event) error {
	logrus.WithFields(logrus.Fields{
		"aggregate.id":   event.Aggregate.Id,
		"aggregate.type": event.Aggregate.Type,
		"event.type":     event.Type,
	}).Info("Handling event for subscribed cache.")

	cached, err := uc.GetAggregate(*event.GetAggregate())
	if err != nil {
		return err
	}

	if cached == nil {
		if event.Parent != nil {
			c, err := uc.GetAggregate(*event.Parent)
			if err != nil {
				return err
			}
			cached = c
		} else {
			cached = uc.target()
		}

		if cached == nil {
			return errors.New("could not find aggregate")
		}
	}

	err = Project(cached, event)
	if err != nil {
		return err
	}

	err = uc.Put(cached)
	if err != nil {
		return err
	}

	// Do not publish old events
	ets, _ := ptypes.Timestamp(event.Timestamp) // TODO replace with a token or flag that cache has cached up.
	if ets.After(uc.ts) {
		n := newNotification(cached, event)
		logrus.WithFields(logrus.Fields{
			"event.id":          event.Id,
			"aggregate.id":      event.Aggregate.Id,
			"aggregate.type":    event.Aggregate.Type,
			"notification.type": n.EventType,
		}).Info("Cache handling done. Sending out Notification.")
		return uc.Publisher.Publish(n)
	} else {
		return nil
	}
}

type Notification struct {
	*pubsub.EmptyMsg
	Payload   Aggregator
	EventType string
}

func newNotification(entity Aggregator, event *Event) *Notification {
	return &Notification{
		EmptyMsg:  pubsub.NewEmptyMsg(event.Labels(), event.CreatedAt()),
		Payload:   entity,
		EventType: event.Type,
	}
}
