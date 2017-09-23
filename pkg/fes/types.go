package fes

import "github.com/fission/fission-workflows/pkg/util/pubsub"

// Fast, minimal Event Sourcing
type Aggregator interface {
	// Entity-specific
	ApplyEvent(event *Event) error

	// The aggregate provides type information about the entity, such as the aggregate id and the aggregate type.
	//
	// Implemented by AggregatorMixin
	Aggregate() Aggregate

	// Update entity to the provided target state
	//
	// This is implemented by AggregatorMixin, can be overridden for performance approach
	UpdateState(targetState Aggregator) error
}

type EventHandler interface {
	HandleEvent(event *Event) error
}

// EventStore is a persistent store for events
type EventStore interface {
	EventHandler
	Get(aggregate *Aggregate) ([]*Event, error)
	List(matcher StringMatcher) ([]Aggregate, error)
}

// EventBus is the volatile reactive store that processes, stores events, and notifies subscribers
//type Dispatcher interface {
//	EventHandler
//	//pubsub.Publisher
//}

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
}

type CacheReaderWriter interface {
	CacheReader
	CacheWriter
}

type StringMatcher interface {
	Match(target string) bool
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
