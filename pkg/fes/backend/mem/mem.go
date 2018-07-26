package mem

import (
	"errors"
	"sync"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	ErrInvalidAggregate = errors.New("invalid aggregate")

	eventsAppended = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "fes",
		Subsystem: "mem",
		Name:      "events_appended_total",
		Help:      "Count of appended events (excluding any internal events).",
	}, []string{"eventType"})
)

// An in-memory, fes backend for development and testing purposes
type Backend struct {
	pubsub.Publisher
	contents map[fes.Aggregate][]*fes.Event
	lock     sync.RWMutex
}

func NewBackend() *Backend {
	return &Backend{
		Publisher: pubsub.NewPublisher(),
		contents:  map[fes.Aggregate][]*fes.Event{},
		lock:      sync.RWMutex{},
	}
}

func (b *Backend) Append(event *fes.Event) error {
	if !fes.ValidateAggregate(event.Aggregate) {
		return ErrInvalidAggregate
	}

	key := *event.Aggregate
	b.lock.Lock()
	defer b.lock.Unlock()

	events, ok := b.contents[key]
	if !ok {
		events = []*fes.Event{}
	}
	b.contents[key] = append(events, event)

	eventsAppended.WithLabelValues(event.Type).Inc()
	return b.Publish(event)
}

func (b *Backend) Get(key fes.Aggregate) ([]*fes.Event, error) {
	if !fes.ValidateAggregate(&key) {
		return nil, ErrInvalidAggregate
	}
	b.lock.RLock()
	defer b.lock.RUnlock()
	events, ok := b.contents[key]
	if !ok {
		events = []*fes.Event{}
	}
	return events, nil
}

func (b *Backend) List(matchFn fes.StringMatcher) ([]fes.Aggregate, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	var results []fes.Aggregate
	for k := range b.contents {
		if matchFn(k.Type + k.Id) {
			results = append(results, k)
		}
	}
	return results, nil
}
