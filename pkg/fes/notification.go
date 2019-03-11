package fes

import (
	"time"

	"github.com/fission/fission-workflows/pkg/util/labels"
)

const (
	PubSubLabelEventID        = "event.id"
	PubSubLabelEventType      = "event.type"
	PubSubLabelAggregateType  = "aggregate.type"
	PubSubLabelAggregateID    = "aggregate.id"
	DefaultNotificationBuffer = 64
)

// Notification is the message send to subscribers of the event store.
//
// It is an annotated fes.Event; it includes snapshots of the affected entity before (if applicable)
// and after the application of the event.
type Notification struct {
	// Old contains the snapshot of the entity before applying the event.
	//
	// This can be nil, if the event caused the creation of the entity.
	Old Entity

	// Updated is the snapshot of entity after applying the event
	Updated Entity

	// Event is the event that triggered the notification.
	Event *Event
}

func NewNotification(old Entity, new Entity, event *Event) *Notification {
	if event == nil {
		panic("event cannot be nil")
	}
	if new == nil {
		panic("new snapshot cannot be nil")
	}
	if new.Aggregate() == *event.Aggregate {
		panic("aggregate of event does not match aggregate of new entity snapshot")
	}
	if old != nil {
		if old.Aggregate() == new.Aggregate() {
			panic("aggregate of old entity snapshot does not match aggregate of the new entity snapshot")
		}
	}
	return &Notification{
		Old:     old,
		Updated: new,
		Event:   event,
	}
}

// Labels returns the labels of the event part of the notification.
//
// Necessary to conform with the pubsub.Msg interface.
func (n *Notification) Labels() labels.Labels {
	return n.Event.Labels()
}

// CreatedAt returns the timestamp of the event within the notification.
//
// Necessary to conform with the pubsub.Msg interface.
func (n *Notification) CreatedAt() time.Time {
	return n.Event.CreatedAt()
}

// Aggregate provides the aggregate for this notification. This is guaranteed to match for the included event/snapshots.
func (n *Notification) Aggregate() Aggregate {
	return n.Updated.Aggregate()
}
