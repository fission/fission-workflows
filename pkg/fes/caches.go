package fes

import (
	"fmt"
	"sync"
	"time"

	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/golang/protobuf/ptypes"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	PubSubLabelEventID       = "event.id"
	PubSubLabelEventType     = "event.type"
	PubSubLabelAggregateType = "aggregate.type"
	PubSubLabelAggregateID   = "aggregate.id"
)

var (
	cacheCount = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "fes",
		Subsystem: "cache",
		Name:      "current_cache_counts",
		Help:      "The current number of entries in the caches",
	}, []string{"name"})
	// TODO add additional metrics once cache has been improved
)

func init() {
	prometheus.MustRegister(cacheCount)
}

// MapCache provides a simple non-preempting map-based CacheReaderWriter implementation.
type MapCache struct {
	Name     string
	contents map[string]map[string]Entity // Map: AggregateType -> AggregateId -> entity
	lock     *sync.RWMutex
}

func NewMapCache() *MapCache {
	c := &MapCache{
		contents: map[string]map[string]Entity{},
		lock:     &sync.RWMutex{},
	}
	c.Name = fmt.Sprintf("%p", c)
	return c
}

func NewNamedMapCache(name string) *MapCache {
	return &MapCache{
		contents: map[string]map[string]Entity{},
		lock:     &sync.RWMutex{},
		Name:     name,
	}
}

func (rc *MapCache) Get(entity Entity) error {
	if err := ValidateEntity(entity); err != nil {
		return err
	}

	ref := entity.Aggregate()
	cached, err := rc.GetAggregate(ref)
	if err != nil {
		return err
	}
	if cached == nil {
		return ErrEntityNotFound.WithEntity(entity)
	}

	e := entity.UpdateState(cached)

	return e
}

func (rc *MapCache) GetAggregate(aggregate Aggregate) (Entity, error) {
	if err := ValidateAggregate(&aggregate); err != nil {
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

func (rc *MapCache) Put(entity Entity) error {
	if err := ValidateEntity(entity); err != nil {
		return err
	}
	ref := entity.Aggregate()

	rc.lock.Lock()
	defer rc.lock.Unlock()
	if _, ok := rc.contents[ref.Type]; !ok {
		rc.contents[ref.Type] = map[string]Entity{}
	}
	rc.contents[ref.Type][ref.Id] = entity
	cacheCount.WithLabelValues(rc.Name).Inc()
	return nil
}

func (rc *MapCache) Invalidate(ref *Aggregate) {
	if err := ValidateAggregate(ref); err != nil {
		return
	}
	rc.lock.Lock()
	defer rc.lock.Unlock()
	delete(rc.contents[ref.Type], ref.Id)
	cacheCount.WithLabelValues(rc.Name).Dec()
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

// A SubscribedCache is subscribed to an event emitter
type SubscribedCache struct {
	pubsub.Publisher
	CacheReaderWriter
	createdAt     time.Time
	entityFactory func() Entity
	closeC        chan struct{}
}

func NewSubscribedCache(cache CacheReaderWriter, factory func() Entity, sub *pubsub.Subscription) *SubscribedCache {
	c := &SubscribedCache{
		Publisher:         pubsub.NewPublisher(),
		CacheReaderWriter: cache,
		entityFactory:     factory,
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
				event, ok := e.(*Event)
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
func (uc *SubscribedCache) applyEvent(event *Event) error {
	logrus.WithFields(logrus.Fields{
		PubSubLabelEventID:       event.Id,
		PubSubLabelEventType:     event.Type,
		PubSubLabelAggregateID:   event.Aggregate.Id,
		PubSubLabelAggregateType: event.Aggregate.Type,
	}).Debug("Applying event to subscribed cache.")

	if err := ValidateEvent(event); err != nil {
		return err
	}

	// Attempt to fetch the entity from the cache.
	entity, err := uc.getOrCreateAggregateForEvent(event)
	if err != nil {
		return err
	}

	// Clone the entity to avoid race conflicts with subscribers of the cache.
	// TODO maybe safer and more readable to replace implementation by a immutable projection?
	entityCopy := entity.CopyEntity()

	// Apply the event on to the new copy of the entity
	err = Project(entityCopy, event)
	if err != nil {
		return err
	}

	// Replace the old entity in the cache with the new (copied) entity
	err = uc.Put(entityCopy)
	if err != nil {
		return err
	}

	// Do not publish replayed events as notifications.
	// We assume that this includes all events with a timestamp of before the cache was created.
	ets, _ := ptypes.Timestamp(event.Timestamp)
	if ets.After(uc.createdAt) {
		// Publish the event (along with the updated entity) to subscribers
		n := newNotification(entityCopy, event)
		logrus.WithFields(logrus.Fields{
			"event.id":          event.Id,
			"aggregate.id":      event.Aggregate.Id,
			"aggregate.type":    event.Aggregate.Type,
			"notification.type": n.EventType,
		}).Debug("SubscribedCache: publishing notification of event.")
		return uc.Publisher.Publish(n)
	}
	return nil
}

func (uc *SubscribedCache) Close() error {
	close(uc.closeC)
	return nil
}

func (uc *SubscribedCache) getOrCreateAggregateForEvent(event *Event) (Entity, error) {
	cached, err := uc.GetAggregate(*event.Aggregate)
	if err != nil {
		return nil, err
	}
	if cached == nil {
		// Case: for task events the parent entity (the invocation) handles the events.
		// So we need to send the event there.
		if event.Parent != nil {
			c, err := uc.GetAggregate(*event.Parent)
			if err != nil {
				return nil, err
			}
			cached = c
		} else {
			cached = uc.entityFactory()
		}
		// We truly can't find a entity responsible for the event
		if cached == nil {
			return nil, ErrEntityNotFound.WithEvent(event)
		}
	}
	return cached, nil
}

// LoadingCache looks into a backing data store in case there is a cache miss
type LoadingCache struct {
	cache         CacheReaderWriter
	client        Backend
	entityFactory func() Entity
}

func NewLoadingCache(cache CacheReaderWriter, client Backend, entityFactory func() Entity) *LoadingCache {
	return &LoadingCache{
		cache:         cache,
		client:        client,
		entityFactory: entityFactory,
	}
}

// TODO provide option to force fallback or only do quick cache lookup.
// TODO sync cache with store while you are at it.
func (c *LoadingCache) List() []Aggregate {
	// First c
	esAggregates, err := c.client.List(nil)
	if err != nil {
		logrus.Errorf("Failed to list event store aggregates: %v", err)
	}
	return esAggregates
}

func (c *LoadingCache) GetAggregate(a Aggregate) (Entity, error) {
	if err := ValidateAggregate(&a); err != nil {
		return nil, err
	}

	entity := c.entityFactory()
	err := c.getFromEventStore(a, entity)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (c *LoadingCache) Get(entity Entity) error {
	if err := ValidateEntity(entity); err != nil {
		return err
	}

	return c.getFromEventStore(entity.Aggregate(), entity)
}

// getFromEventStore assumes that it can mutate target entity
func (c *LoadingCache) getFromEventStore(aggregate Aggregate, target Entity) error {
	err := c.cache.Get(target)
	if err != nil {
		if !ErrEntityNotFound.Is(err) {
			return err
		}

		// Look up relevant events in event store
		events, err := c.client.Get(aggregate)
		if err != nil {
			return err
		}
		if len(events) == 0 {
			return ErrEntityNotFound.WithAggregate(&aggregate).WithEntity(target)
		}

		// Reconstruct entity by replaying all events
		err = Project(target, events...)
		if err != nil {
			return err
		}

		// EvalCache entityFactory
		return c.cache.Put(target) // TODO ensure entity is a copy
	}

	return nil
}
