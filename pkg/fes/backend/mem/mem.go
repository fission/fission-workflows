// package mem contains an implementation of the fes backend using an in-memory cache.
//
// This implementation is typically used for development and test purposes. However,
// if you are targeting pure performance, you can use this backend to effectively trade
// in persistence-related guarantees (e.g. fault-tolerance) to avoid overhead introduced
// by other event stores, such as the NATS implementation.
package mem

import (
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fes/backend"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/golang/protobuf/ptypes"
	"github.com/hashicorp/golang-lru"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	ErrEventLimitExceeded = &fes.EventStoreErr{
		S: "event limit exceeded",
	}

	cacheKeys = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "fes",
		Subsystem: "mem",
		Name:      "keys",
		Help:      "Number of keys in the store by entity type.",
	}, []string{"type"})

	cacheEvents = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "fes",
		Subsystem: "mem",
		Name:      "events",
		Help:      "Number of events in the store by entity type.",
	}, []string{"type"})
)

func init() {
	prometheus.MustRegister(cacheKeys, cacheEvents)
}

// Config contains the user-configurable options of the in-memory backend.
//
// To limit the memory consumption of the backend, you can make use of the MaxKeys and
// MaxEventsPerKey. The absolute maximum memory usage is the product of the two limits.
type Config struct {
	// MaxKeys specifies the limit of keys (or aggregates) in the backend.
	// If set to 0, MaxInt32 will be used as the limit.
	MaxKeys int

	// MaxEventsPerKey specifies a limit on the number
	MaxEventsPerKey int
}

// Backend is an in-memory, fes-compatible backend using a map for active entities with a LRU cache to store completed
// event streams, evicting oldest ones if it runs out of space. Active entities will never be deleted.
type Backend struct {
	pubsub.Publisher
	Config
	buf       *lru.Cache // map[fes.Aggregate][]*fes.Event
	store     map[fes.Aggregate][]*fes.Event
	storeLock sync.RWMutex
	entries   *int32
}

func NewBackend(cfgs ...Config) *Backend {
	cfg := Config{
		MaxKeys: math.MaxInt32,
	}
	if len(cfgs) > 0 {
		providedCfg := cfgs[0]
		if providedCfg.MaxKeys <= 0 {
			cfg.MaxKeys = math.MaxInt32
		} else {
			cfg.MaxKeys = providedCfg.MaxKeys
		}
		if providedCfg.MaxEventsPerKey > 0 {
			cfg.MaxEventsPerKey = providedCfg.MaxEventsPerKey
		}
	}

	e := int32(0)
	b := &Backend{
		Publisher: pubsub.NewPublisher(),
		Config:    cfg,
		store:     map[fes.Aggregate][]*fes.Event{},
		entries:   &e,
	}

	cache, err := lru.NewWithEvict(cfg.MaxKeys, b.evict)
	if err != nil {
		panic(err)
	}
	b.buf = cache
	return b
}

func (b *Backend) Append(event *fes.Event) error {
	if err := fes.ValidateEvent(event); err != nil {
		return err
	}
	key := *event.Aggregate
	if event.Parent != nil {
		key = *event.Parent
	}
	var newEntry bool

	b.storeLock.Lock()
	defer b.storeLock.Unlock()

	events, ok, fromStore := b.get(key)
	if !ok {
		events = []*fes.Event{}
		newEntry = true

		// Verify that there is space for the new event
		if !b.fitBuffer() {
			return fes.ErrEventStoreOverflow.WithAggregate(&key)
		}
	}

	// Check if event stream is not out of limit
	if b.MaxEventsPerKey > 0 && len(events) > b.MaxEventsPerKey {
		return ErrEventLimitExceeded.WithAggregate(&key)
	}

	if !fromStore {
		b.promote(key)
	}

	b.store[key] = append(events, event)
	logrus.Infof("Event appended: %s - %v", event.Aggregate.Format(), event.Type)

	if event.GetHints().GetCompleted() {
		b.demote(key)
	}
	err := b.Publish(event)

	if newEntry {
		atomic.AddInt32(b.entries, 1)
		cacheKeys.WithLabelValues(key.Type).Inc()
	}
	// Record the time it took for the event to be propagated from publisher to subscriber.
	ts, _ := ptypes.Timestamp(event.Timestamp)
	backend.EventDelay.Observe(float64(time.Now().Sub(ts).Nanoseconds()))
	backend.EventsAppended.WithLabelValues(event.Type).Inc()
	return err
}

func (b *Backend) Get(key fes.Aggregate) ([]*fes.Event, error) {
	if err := fes.ValidateAggregate(&key); err != nil {
		return nil, err
	}
	b.storeLock.RLock()
	events, ok, _ := b.get(key)
	b.storeLock.RUnlock()
	if !ok {
		events = []*fes.Event{}
	}
	return events, nil
}

func (b *Backend) Len() int {
	return int(atomic.LoadInt32(b.entries))
}

// Note: list does not return the cache contents
func (b *Backend) List(matcher fes.AggregateMatcher) ([]fes.Aggregate, error) {
	var results []fes.Aggregate
	b.storeLock.RLock()
	for key := range b.store {
		if matcher == nil || matcher(key) {
			results = append(results, key)
		}
	}
	b.storeLock.RUnlock()
	return results, nil
}

func (b *Backend) get(key fes.Aggregate) (events []*fes.Event, ok bool, fromStore bool) {
	// First check the store
	i, ok := b.store[key]
	if ok {
		return assertEventList(i), ok, true
	}

	// Fallback: check the buf buffer
	e, ok := b.buf.Get(key)
	if ok {
		return assertEventList(e), ok, false
	}
	return nil, ok, false
}

// promote moves a buf buffer entry to the store
func (b *Backend) promote(key fes.Aggregate) {
	events, ok, fromStore := b.get(key)
	if !ok || fromStore {
		return
	}
	b.buf.Remove(key)
	b.store[key] = events
}

// demote moves a store entry to the buf buffer
func (b *Backend) demote(key fes.Aggregate) {
	events, ok, fromStore := b.get(key)
	if !ok || !fromStore {
		return
	}
	delete(b.store, key)
	b.buf.Add(key, events)
}

func (b *Backend) fitBuffer() bool {
	last := -1
	size := b.Len()
	for size != last && size >= b.MaxKeys {
		b.buf.RemoveOldest()
		last = size
		size = b.Len()
	}
	return size < b.MaxKeys
}

func (b *Backend) evict(k, v interface{}) {
	logrus.Debugf("Evicted: %v", k)

	// Update gauges
	t := assertAggregate(k).Type
	events := assertEventList(v)
	cacheKeys.WithLabelValues(t).Dec()
	cacheEvents.WithLabelValues(t).Add(-1 * float64(len(events)))

	// Decrement entries counter
	atomic.AddInt32(b.entries, -1)
}

func assertEventList(i interface{}) []*fes.Event {
	events, typeOk := i.([]*fes.Event)
	if !typeOk {
		panic(fmt.Sprintf("found unexpected value type in the cache: %T", i))
	}
	return events
}

func assertAggregate(i interface{}) fes.Aggregate {
	key, ok := i.(fes.Aggregate)
	if !ok {
		panic(fmt.Sprintf("found unexpected key type in the cache: %T", i))
	}
	return key
}
