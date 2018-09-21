package fes

import (
	"errors"
	"reflect"
	"strings"
	"time"

	"github.com/fission/fission-workflows/pkg/api/events"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
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
	if msg == nil {
		return nil, errors.New("msg cannot have no message")
	}

	data, err := ptypes.MarshalAny(msg)
	if err != nil {
		return nil, err
	}

	var t string
	if e, ok := msg.(events.Event); ok {
		t = e.Type()
	} else {
		reflect.Indirect(reflect.ValueOf(msg)).Type().Name()
	}

	return &Event{
		Aggregate: &aggregate,
		Data:      data,
		Timestamp: ptypes.TimestampNow(),
		Type:      t,
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
