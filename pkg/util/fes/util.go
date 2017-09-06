package fes

import (
	"errors"
	"strings"
)

func validateAggregate(aggregate Aggregate) error {
	if len(aggregate.Id) == 0 {
		return errors.New("Aggregate does not contain id")
	}

	if len(aggregate.Type) == 0 {
		return errors.New("Aggregate does not contain type")
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
