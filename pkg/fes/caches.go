package fes

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

var (
	ErrNotFound = errors.New("could not find entity")
)

// MapCache provides a simple non-preempting map-based CacheReaderWriter implementation.
type MapCache struct {
	contents map[string]map[string]Aggregator // Map: AggregateType -> AggregateId -> entity
	lock     *sync.RWMutex
}

func NewMapCache() *MapCache {
	return &MapCache{
		contents: map[string]map[string]Aggregator{},
		lock:     &sync.RWMutex{},
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
		return ErrNotFound
	}

	e := entity.UpdateState(cached)

	return e
}

func (rc *MapCache) GetAggregate(aggregate Aggregate) (Aggregator, error) {
	err := validateAggregate(aggregate)
	if err != nil {
		return nil, err
	}

	rc.lock.RLock()
	defer rc.lock.RUnlock()
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

	rc.lock.Lock()
	defer rc.lock.Unlock()
	if _, ok := rc.contents[ref.Type]; !ok {
		rc.contents[ref.Type] = map[string]Aggregator{}
	}

	rc.contents[ref.Type][ref.Id] = entity
	return nil
}

func (rc *MapCache) Invalidate(ref *Aggregate) {
	rc.lock.Lock()
	defer rc.lock.Unlock()
	delete(rc.contents[ref.Type], ref.Id)
}

func (rc *MapCache) List() []Aggregate {
	rc.lock.RLock()
	defer rc.lock.RUnlock()
	var results []Aggregate
	for atype := range rc.contents {
		for _, entity := range rc.contents[atype] {
			results = append(results, entity.Aggregate())
		}
	}
	return results
}

// A SubscribedCache is subscribed to some event emitter, using a mutex to ensure thread safety
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
				logrus.Info("SubscribedCache worker was canceled.")
				return
			case msg := <-sub.Ch:
				// Discard invalid messages
				event, ok := msg.(*Event)
				if !ok {
					logrus.WithField("msg", msg).Error("Received a malformed message. Ignoring.")
					continue
				}
				logrus.WithField("msg", msg.Labels()).Debug("Cache received new event.")
				err := c.ApplyEvent(event)
				if err != nil {
					logrus.WithField("err", err).Error("Failed to handle event. Continuing.")
				}
			}
		}
	}()

	return c
}

func (uc *SubscribedCache) ApplyEvent(event *Event) error {
	logrus.WithFields(logrus.Fields{
		"event.id":       event.Id,
		"aggregate.id":   event.Aggregate.Id,
		"aggregate.type": event.Aggregate.Type,
		"event.type":     event.Type,
	}).Debug("Applying event to subscribed cache.")

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
			return ErrNotFound
		}
	}

	// clone to avoid race conflicts with subscribers of the cache.
	cached = cached.GenericCopy()

	err = Project(cached, event)
	if err != nil {
		return err
	}

	err = uc.Put(cached)
	if err != nil {
		return err
	}

	// Do not publish replayed events as notifications
	// TODO replace with a token or flag that cache has cached up.
	ets, _ := ptypes.Timestamp(event.Timestamp)
	if ets.After(uc.ts) {
		n := newNotification(cached, event)
		logrus.WithFields(logrus.Fields{
			"event.id":          event.Id,
			"aggregate.id":      event.Aggregate.Id,
			"aggregate.type":    event.Aggregate.Type,
			"notification.type": n.EventType,
		}).Debug("Cache handling done. Sending out Notification.")
		err = uc.Publisher.Publish(n)
	}
	return err
}

// FallbackCache looks into a backing data store in case there is a cache miss
type FallbackCache struct {
	cache  CacheReaderWriter
	client EventStore
	domain StringMatcher
	target func() Aggregator // TODO extract to a TypedSubscription
}

func (c *FallbackCache) List() []Aggregate {
	esAggregates, err := c.client.List(c.domain)
	if err != nil {
		logrus.WithField("matcher", c.domain).
			WithField("err", err).
			Error("Failed to list event store aggregates")
	}
	for _, aggregate := range esAggregates {
		entity, err := c.cache.GetAggregate(aggregate)
		if err != nil || entity == nil {
			events, err := c.client.Get(&aggregate)
			if err != nil {
				logrus.WithField("err", err).Error("failed to get missed entity from event store")
				continue
			}
			e := c.target()
			err = Project(e, events...)
			if err != nil {
				logrus.WithField("err", err).Error("failed to project missed entity")
				continue
			}
			err = c.cache.Put(e)
			if err != nil {
				logrus.WithField("err", err).Error("failed to store missed entity")
				continue
			}
		}
	}
	return esAggregates
}

func (c *FallbackCache) GetAggregate(a Aggregate) (Aggregator, error) {
	cached, err := c.cache.GetAggregate(a)
	if err != nil {
		if err == ErrNotFound {
			logrus.WithField("entity", a).Info("Cache miss! Checking backing event store.")
			entity := c.target()
			err := c.getFromEventStore(a, entity)
			if err != nil {
				return nil, err
			}
			return entity, nil
		}
	}
	return cached, nil
}

func (c *FallbackCache) Get(entity Aggregator) error {
	err := c.cache.Get(entity)
	if err != nil {
		if err == ErrNotFound {
			logrus.WithField("entity", entity.Aggregate()).Info("Cache miss! Checking backing event store.")
			return c.getFromEventStore(entity.Aggregate(), entity)
		} else {
			return err
		}
	}
	return nil
}

func (c *FallbackCache) getFromEventStore(aggregate Aggregate, target Aggregator) error {
	// Look up relevant events in event store
	events, err := c.client.Get(&aggregate)
	if err != nil {
		return err
	}
	if len(events) == 0 {
		return ErrNotFound
	}

	// Reconstruct target
	err = Project(target, events...)
	if err != nil {
		return err
	}

	// Cache target
	err = c.cache.Put(target) // TODO ensure target is a copy
	if err != nil {
		logrus.WithField("target", target).WithField("err", err).Warn("Failed to cache fetched target")
	}

	return nil
}
