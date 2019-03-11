package cache

import (
	"time"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/golang/protobuf/ptypes"
	"github.com/hashicorp/golang-lru"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	cacheCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "fes",
		Subsystem: "cache",
		Name:      "current_cache_counts",
		Help:      "The current number of entries in the caches",
	}, []string{"name"})
)

func init() {
	prometheus.MustRegister(cacheCount)
}

type LRUCache struct {
	contents *lru.Cache
}

func NewLRUCache(size int) *LRUCache {
	c, err := lru.New(size)
	if err != nil {
		panic(err)
	}
	return &LRUCache{
		contents: c,
	}
}

func (c *LRUCache) GetAggregate(a fes.Aggregate) (fes.Entity, error) {
	if err := fes.ValidateAggregate(&a); err != nil {
		return nil, err
	}
	i, ok := c.contents.Get(a)
	if !ok {
		return nil, fes.ErrEntityNotFound.WithAggregate(&a)
	}
	return i.(fes.Entity), nil
}

func (c *LRUCache) Put(entity fes.Entity) error {
	if err := fes.ValidateEntity(entity); err != nil {
		return err
	}
	a := fes.GetAggregate(entity)
	c.contents.Add(a, entity)
	return nil
}

func (c *LRUCache) List() []fes.Aggregate {
	keys := c.contents.Keys()
	results := make([]fes.Aggregate, len(keys))
	for i, key := range keys {
		results[i] = key.(fes.Aggregate)
	}
	return results
}

func (c *LRUCache) Invalidate(a fes.Aggregate) {
	if err := fes.ValidateAggregate(&a); err != nil {
		logrus.Warnf("Failed to invalidate entry in cache: %v", err)
		return
	}
	c.contents.Remove(a)
}

// A SubscribedCache is subscribed to an event emitter
type SubscribedCache struct {
	pubsub.Publisher
	fes.CacheReaderWriter
	createdAt time.Time
	projector fes.Projector
	closeC    chan struct{}
}

func NewSubscribedCache(cache fes.CacheReaderWriter, projector fes.Projector,
	sub *pubsub.Subscription) *SubscribedCache {
	c := &SubscribedCache{
		Publisher:         pubsub.NewPublisher(),
		CacheReaderWriter: cache,
		projector:         projector,
		createdAt:         time.Now(),
	}

	c.closeC = make(chan struct{})
	go func() {
		for {
			select {
			case <-c.closeC:
				logrus.Debug("SubscribedCache: listener stopped.")
				return
			case e := <-sub.Ch:
				event, ok := e.(*fes.Event)
				if !ok {
					logrus.WithField("event", e).Error("Ignoring received malformed event.")
					continue
				}
				logrus.WithField("msg", e.Labels()).Debug("SubscribedCache: received event.")
				err := c.applyEvent(event)
				if err != nil {
					logrus.WithField("event", event).Errorf("Failed to handle event: %v", err)
				}
			}
		}
	}()

	return c
}

// applyEvent applies an event to the cache. It retrieves the corresponding entity from the cache, copies it,
// applies the event to it, and replaces the old with the new entity in the cache.
func (uc *SubscribedCache) applyEvent(event *fes.Event) error {
	logrus.WithFields(logrus.Fields{
		fes.PubSubLabelEventID:       event.Id,
		fes.PubSubLabelEventType:     event.Type,
		fes.PubSubLabelAggregateID:   event.Aggregate.Id,
		fes.PubSubLabelAggregateType: event.Aggregate.Type,
	}).Debug("Applying event to subscribed cache.")

	if err := fes.ValidateEvent(event); err != nil {
		return err
	}

	// Attempt to fetch the entity from the cache.
	old, err := uc.getOrCreateAggregateForEvent(event)
	if err != nil {
		return err
	}

	// Apply the event on to the new copy of the entity
	updated, err := uc.projector.Project(old, event)
	if err != nil {
		return err
	}

	// Replace the old entity in the cache with the new (copied) entity
	err = uc.Put(updated)
	if err != nil {
		return err
	}

	// Do not publish replayed events as notifications.
	// We assume that this includes all events with a timestamp of before the cache was created.
	ets, _ := ptypes.Timestamp(event.Timestamp)
	if ets.After(uc.createdAt) {
		// Publish the event (along with the updated entity) to subscribers
		n := fes.NewNotification(old, updated, event)
		logrus.WithFields(logrus.Fields{
			"event.id":       event.Id,
			"aggregate.id":   event.Aggregate.Id,
			"aggregate.type": event.Aggregate.Type,
			"event.type":     event.Type,
		}).Debug("SubscribedCache: publishing notification of event.")
		return uc.Publisher.Publish(n)
	}
	return nil
}

func (uc *SubscribedCache) Close() error {
	close(uc.closeC)
	return nil
}

func (uc *SubscribedCache) getOrCreateAggregateForEvent(event *fes.Event) (fes.Entity, error) {
	key := getKey(event)
	cached, err := uc.GetAggregate(key)
	if err != nil && err != fes.ErrEntityNotFound {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}
	// Case: for task events the parent entity (the invocation) handles the events.
	// So we need to send the event there.
	if event.Parent != nil {
		parentKey := *event.Parent
		c, err := uc.GetAggregate(parentKey)
		if err != nil {
			return nil, err
		}
		if c != nil {
			return c, nil
		}
		key = parentKey
	}

	return uc.projector.NewProjection(key)
}

// LoadingCache looks into a backing data store in case there is a cache miss
type LoadingCache struct {
	cache     fes.CacheReaderWriter
	client    fes.Backend
	projector fes.Projector
}

func NewLoadingCache(cache fes.CacheReaderWriter, client fes.Backend, projector fes.Projector) *LoadingCache {
	return &LoadingCache{
		cache:     cache,
		client:    client,
		projector: projector,
	}
}

// List for a LoadingCache returns the keys of all entities in the cache.
//
// TODO provide option to force fallback or only do quick cache lookup.
// TODO sync cache with store while you are at it.
func (c *LoadingCache) List() []fes.Aggregate {
	return c.cache.List()
}

func (c *LoadingCache) GetAggregate(key fes.Aggregate) (fes.Entity, error) {
	if err := fes.ValidateAggregate(&key); err != nil {
		return nil, err
	}

	// Check cache first
	cached, err := c.cache.GetAggregate(key)
	// Ensure that the error was regarding the entity not being available
	if err != nil && !fes.ErrEntityNotFound.Is(err) {
		return nil, err
	}
	if cached != nil {
		return cached, nil
	}

	// Otherwise check store
	entity, err := c.getFromEventStore(key)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (c *LoadingCache) Put(entity fes.Entity) error {
	return c.cache.Put(entity)
}

func (c *LoadingCache) Invalidate(key fes.Aggregate) {
	c.cache.Invalidate(key)
}

func (c *LoadingCache) Load(key fes.Aggregate) error {
	_, err := c.GetAggregate(key)
	return err
}

// getFromEventStore assumes that it can mutate target entity
func (c *LoadingCache) getFromEventStore(aggregate fes.Aggregate) (fes.Entity, error) {
	// Look up relevant events in event store
	events, err := c.client.Get(aggregate)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, fes.ErrEntityNotFound.WithAggregate(&aggregate)
	}

	base, err := c.projector.NewProjection(aggregate)
	if err != nil {
		return nil, err
	}

	// Reconstruct entity by replaying all events
	entity, err := c.projector.Project(base, events...)
	if err != nil {
		return nil, err
	}

	// Cache retrieved entity
	if err := c.Put(entity); err != nil {
		return nil, err
	}
	return entity, nil
}

// To ensure that tasks end up in invocation entities: if an event has a parent aggregate,
// this means that the event should be send to the parent aggregate instead.
func getKey(e *fes.Event) fes.Aggregate {
	if e.Parent != nil {
		return *e.Parent
	}
	return *e.Aggregate
}
