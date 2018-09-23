package testutil

import (
	"fmt"
	"sync"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/golang/protobuf/proto"
	"github.com/sirupsen/logrus"
)

// Entity is a stub implementation of a fes.Entity,
// which simply appends all contents of the DummyEvents it receives.
type Entity struct {
	*fes.BaseEntity
	S string
}

func (e *Entity) ApplyEvent(event *fes.Event) error {
	msg, err := fes.ParseEventData(event)
	if err != nil {
		return err
	}

	dummyEvent, ok := msg.(*DummyEvent)
	if !ok {
		return fmt.Errorf("entity expects DummyEvent, but received %T", msg)
	}

	e.S += dummyEvent.Msg
	logrus.Infof("Applied event to entity: %v", e.S)

	return nil
}

func (e *Entity) CopyEntity() fes.Entity {
	return &Entity{
		S: e.S,
	}
}

func NewStringAppendEntity(key fes.Aggregate) fes.Entity {
	e := &Entity{}
	e.BaseEntity = fes.NewBaseEntity(e, key)
	return e
}

func CreateDummyEvent(key fes.Aggregate, payload *DummyEvent) *fes.Event {
	event, err := fes.NewEvent(key, payload)
	if err != nil {
		panic(err)
	}
	return event
}

func ToDummyEvents(key fes.Aggregate, msg string) []*fes.Event {
	events := make([]*fes.Event, len(msg))
	for i, c := range msg {
		events[i] = CreateDummyEvent(key, &DummyEvent{
			Msg: string(c),
		})
	}
	return events
}

// Backend is a stub implementation of a fes.Backend
type Backend struct {
	events map[fes.Aggregate][]*fes.Event
	lock   sync.RWMutex
}

func NewBackend() *Backend {
	return &Backend{
		events: make(map[fes.Aggregate][]*fes.Event),
	}
}

func (b *Backend) Append(event *fes.Event) error {
	b.lock.Lock()
	defer b.lock.Unlock()
	eventCopy := proto.Clone(event).(*fes.Event)
	eventCopy.Id = fmt.Sprintf("%d", len(b.events[*event.Aggregate]))
	b.events[*event.Aggregate] = append(b.events[*event.Aggregate], eventCopy)
	return nil
}

func (b *Backend) Get(aggregate fes.Aggregate) ([]*fes.Event, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	events, ok := b.events[aggregate]
	if ok {
		return events, nil
	}
	return nil, fes.ErrEntityNotFound
}

func (b *Backend) List(matcher fes.AggregateMatcher) ([]fes.Aggregate, error) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	var keys []fes.Aggregate
	for k := range b.events {
		keys = append(keys, k)
	}
	return keys, nil
}

func (b *Backend) Reset() {
	b.lock.Lock()
	defer b.lock.Unlock()
	b.events = make(map[fes.Aggregate][]*fes.Event)
}

// Cache provides a thread-safe, memory-unrestricted map-based CacheReaderWriter implementation.
type Cache struct {
	Name     string
	contents map[string]map[string]fes.Entity // Map: AggregateType -> AggregateId -> entity
	lock     *sync.RWMutex
}

func NewCache() *Cache {
	c := &Cache{
		contents: map[string]map[string]fes.Entity{},
		lock:     &sync.RWMutex{},
	}
	c.Name = fmt.Sprintf("%p", c)
	return c
}

func (rc *Cache) Get(entity fes.Entity) error {
	if err := fes.ValidateEntity(entity); err != nil {
		return err
	}

	ref := entity.Aggregate()
	cached, err := rc.GetAggregate(ref)
	if err != nil {
		return err
	}
	if cached == nil {
		return fes.ErrEntityNotFound.WithEntity(entity)
	}

	e := entity.UpdateState(cached)

	return e
}

func (rc *Cache) GetAggregate(aggregate fes.Aggregate) (fes.Entity, error) {
	if err := fes.ValidateAggregate(&aggregate); err != nil {
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

func (rc *Cache) Put(entity fes.Entity) error {
	if err := fes.ValidateEntity(entity); err != nil {
		return err
	}
	ref := entity.Aggregate()

	rc.lock.Lock()
	defer rc.lock.Unlock()
	if _, ok := rc.contents[ref.Type]; !ok {
		rc.contents[ref.Type] = map[string]fes.Entity{}
	}
	rc.contents[ref.Type][ref.Id] = entity
	return nil
}

func (rc *Cache) Invalidate(ref fes.Aggregate) {
	if err := fes.ValidateAggregate(&ref); err != nil {
		logrus.Warnf("Failed to invalidate entry in cache: %v", err)
		return
	}
	rc.lock.Lock()
	defer rc.lock.Unlock()
	delete(rc.contents[ref.Type], ref.Id)
}

func (rc *Cache) List() []fes.Aggregate {
	rc.lock.RLock()
	defer rc.lock.RUnlock()
	var results []fes.Aggregate
	for atype := range rc.contents {
		for _, entity := range rc.contents[atype] {
			results = append(results, entity.Aggregate())
		}
	}
	return results
}
