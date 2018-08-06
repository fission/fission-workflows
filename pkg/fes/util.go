package fes

import (
	"errors"
	"strings"
	"time"

	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	log "github.com/sirupsen/logrus"
)

// Project is convenience function to apply events to an entity.
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

type DeepFoldMatcher struct {
	Expected string
}

func (df *DeepFoldMatcher) Match(target string) bool {
	return strings.EqualFold(df.Expected, target)
}

func ContainsMatcher(Substr string) StringMatcher {
	return func(target string) bool {
		return strings.Contains(target, Substr)
	}
}

func NewAggregate(entityType string, entityID string) Aggregate {
	return Aggregate{
		Id:   entityID,
		Type: entityType,
	}
}

type EventOpts struct {
	Event
	Data      proto.Message
	Timestamp time.Time
}

func NewEvent(aggregate Aggregate, msg proto.Message) (*Event, error) {
	var data *any.Any
	if msg == nil {
		return nil, errors.New("event cannot have no message")
	}

	d, err := ptypes.MarshalAny(msg)
	if err != nil {
		return nil, err
	}
	data = d
	return &Event{
		Aggregate: &aggregate,
		Data:      data,
		Timestamp: ptypes.TimestampNow(),
		Type:      events.TypeOf(msg),
		Metadata:  map[string]string{},
	}, nil
}

func validateAggregate(aggregate Aggregate) error {
	if len(aggregate.Id) == 0 {
		return errors.New("aggregate does not contain id")
	}

	if len(aggregate.Type) == 0 {
		return errors.New("aggregate does not contain type")
	}

	return nil
}

func UnmarshalEventData(event *Event) (interface{}, error) {
	d := &ptypes.DynamicAny{}
	err := ptypes.UnmarshalAny(event.Data, d)
	if err != nil {
		return nil, err
	}
	return d.Message, nil
}
