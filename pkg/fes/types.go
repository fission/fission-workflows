package fes

import "github.com/fission/fission-workflows/pkg/util/pubsub"

// Aggregator is a entity that can be updated
// TODO we need to keep more event-related information (such as current index)
type Aggregator interface {

	// Entity-specific
	ApplyEvent(event *Event) error

	// Aggregate provides type information about the entity, such as the aggregate id and the aggregate type.
	//
	// This is implemented by AggregatorMixin
	Aggregate() Aggregate

	// UpdateState mutates the current entity to the provided target state
	//
	// This is implemented by AggregatorMixin, can be overridden for performance approach
	UpdateState(targetState Aggregator) error

	// Copy copies the actual wrapped object. This is useful to get a snapshot of the state.
	GenericCopy() Aggregator
}

type EventAppender interface {
	Append(event *Event) error
}

// Backend is a persistent store for events
type Backend interface {
	EventAppender

	// Get fetches all events that belong to a specific aggregate
	Get(aggregate Aggregate) ([]*Event, error)
	List(matcher StringMatcher) ([]Aggregate, error)
}

// Projector projects events into an entity
type Projector interface {
	Project(target Aggregator, events ...*Event) error
}

type CacheReader interface {
	Get(entity Aggregator) error
	List() []Aggregate
	GetAggregate(a Aggregate) (Aggregator, error)
}

type CacheWriter interface {
	Put(entity Aggregator) error
	Invalidate(entity *Aggregate)
}

type CacheReaderWriter interface {
	CacheReader
	CacheWriter
}

type StringMatcher func(target string) bool

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
