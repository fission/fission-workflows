package fes

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type MockAggregate struct {
	*BaseEntity
	Val int
}

func newMockAggregate(id string, atype string, val int) *MockAggregate {
	m := &MockAggregate{
		Val: val,
	}
	m.BaseEntity = NewBaseEntity(m, Aggregate{id, atype})
	return m
}

func (ma *MockAggregate) CopyEntity() Entity {
	panic("implement me")
}

func (ma *MockAggregate) ApplyEvent(event *Event) error {
	panic("Should not be relevant")
}

func TestNewBaseEntity(t *testing.T) {
	src := newMockAggregate("1", "foo", 1)
	updated := newMockAggregate("1", "foo", 2)

	src.UpdateState(updated)

	assert.Equal(t, src.Val, updated.Val)
}
