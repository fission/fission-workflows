package eventstore

import (
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
)

// Utility package for working with events
func NewEvent(eventId *EventID, eventType string, payload *any.Any) *Event {
	return &Event{
		EventId: eventId,
		Type:    eventType,
		Time:    ptypes.TimestampNow(),
		Data:    payload,
	}
}
