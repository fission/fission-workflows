package events

import (
	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

// Utility package for working with events
func New(eventId *eventstore.EventID, eventType string, payload *any.Any) *eventstore.Event {
	return &eventstore.Event{
		EventId: eventId,
		Type:    eventType,
		Time:    ptypes.TimestampNow(),
		Data:    payload,
	}
}
