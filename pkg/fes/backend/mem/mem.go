package mem

import (
	"errors"
	"sync"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
)

var (
	ErrInvalidAggregate = errors.New("invalid aggregate")
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

	return b.Publish(event)
}

func (b *Backend) Get(aggregate *fes.Aggregate) ([]*fes.Event, error) {
	if !fes.ValidateAggregate(aggregate) {
		return nil, ErrInvalidAggregate
	}

	key := *aggregate
	b.lock.RLock()
	defer b.lock.RUnlock()
	events, ok := b.contents[key]
	if !ok {
		events = []*fes.Event{}
	}
	return events, nil
}

func (b *Backend) List(matcher fes.StringMatcher) ([]fes.Aggregate, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()

	var results []fes.Aggregate
	for k := range b.contents {
		// TODO change matcher to fes.AggregateMatcher instead
		if matcher.Match(k.Type + k.Id) {
			results = append(results, k)
		}
	}
	return results, nil
}
