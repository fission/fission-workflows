package fes

import (
	"errors"
	"reflect"

	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
)

// Project is a convenience function to apply events to an entity.
func Project(entity Entity, events ...*Event) error {
	if entity == nil {
		log.WithField("entity", entity).Warn("Empty entity")
		return nil
	}
	for _, event := range events {
		if event == nil {
			log.WithField("entity", entity).Warn("Empty event received")
			return nil
		}
		err := entity.ApplyEvent(event)
		if err != nil {
			return err
		}
	}
	return nil
}

// NewEvent returns a new event with the provided payload for the provided aggregate or an error if the input data
// was invalid.
//
// It returns one of the following errors:
// - ErrInvalidAggregate: provided aggregate is invalid
// - ErrCorruptedEventPayload: payload is empty or cannot be marshaled to bytes.
func NewEvent(aggregate Aggregate, payload proto.Message) (*Event, error) {
	if err := ValidateAggregate(&aggregate); err != nil {
		return nil, err
	}
	if payload == nil {
		return nil, ErrCorruptedEventPayload.WithAggregate(&aggregate).WithError(errors.New("payload cannot be empty"))
	}

	data, err := ptypes.MarshalAny(payload)
	if err != nil {
		return nil, ErrCorruptedEventPayload.WithAggregate(&aggregate).WithError(err)
	}

	// Hack: use the events.Event interface to use a non-reflection based name. Otherwise fallback to reflection-based.
	var t string
	if e, ok := payload.(events.Event); ok {
		t = e.Type()
	} else {
		t = reflect.Indirect(reflect.ValueOf(payload)).Type().Name()
	}

	return &Event{
		Aggregate: &aggregate,
		Data:      data,
		Timestamp: ptypes.TimestampNow(),
		Type:      t,
		Metadata:  map[string]string{},
	}, nil
}

// ParseEventData parses the payload of the event, returning the generic proto.Message payload.
//
// In case it fails to parse the payload it returns an ErrCorruptedEventPayload
func ParseEventData(event *Event) (proto.Message, error) {
	d := &ptypes.DynamicAny{}
	err := ptypes.UnmarshalAny(event.Data, d)
	if err != nil {
		return nil, ErrCorruptedEventPayload.WithEvent(event).WithError(err)
	}
	return d.Message, nil
}
