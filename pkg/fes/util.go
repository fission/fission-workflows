package fes

import (
	"errors"
	"strings"

	"github.com/golang/protobuf/ptypes"
)

func validateAggregate(aggregate Aggregate) error {
	if len(aggregate.Id) == 0 {
		return errors.New("aggregate does not contain id")
	}

	if len(aggregate.Type) == 0 {
		return errors.New("aggregate does not contain type")
	}

	return nil
}

type DeepFoldMatcher struct {
	Expected string
}

func (df *DeepFoldMatcher) Match(target string) bool {
	return strings.EqualFold(df.Expected, target)
}

type ContainsMatcher struct {
	Substr string
}

func (cm *ContainsMatcher) Match(target string) bool {
	return strings.Contains(target, cm.Substr)
}

func NewAggregate(entityType string, entityID string) Aggregate {
	return Aggregate{
		Id:   entityID,
		Type: entityType,
	}
}

func NewEvent(aggregate Aggregate, data []byte) *Event {
	return &Event{
		Id:        aggregate.Id,
		Type:      aggregate.Type,
		Aggregate: &aggregate,
		Data:      data,
		Timestamp: ptypes.TimestampNow(),
		Parent:    nil,
		Hints:     nil,
	}
}
