package mem

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/util/labels"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/stretchr/testify/assert"
)

func TestBackend_Append(t *testing.T) {
	mem := NewBackend()

	event := fes.NewEvent(fes.NewAggregate("type", "id"), []byte("event 1"))
	err := mem.Append(event)
	assert.NoError(t, err)
	assert.Len(t, mem.contents, 1)

	event2 := fes.NewEvent(fes.Aggregate{}, []byte("event 1"))
	err = mem.Append(event2)
	assert.EqualError(t, err, ErrInvalidAggregate.Error())
	assert.Len(t, mem.contents, 1)

	// Event under existing aggregate
	event3 := fes.NewEvent(fes.NewAggregate("type", "id"), []byte("event 2"))
	err = mem.Append(event3)
	assert.NoError(t, err)
	assert.Len(t, mem.contents, 1)
	assert.Len(t, mem.contents[fes.NewAggregate("type", "id")], 2)

	// Event under new aggregate
	event4 := fes.NewEvent(fes.NewAggregate("Type", "other"), []byte("event 1"))
	err = mem.Append(event4)
	assert.NoError(t, err)
	assert.Len(t, mem.contents, 2)
	assert.Len(t, mem.contents[fes.NewAggregate("Type", "other")], 1)
	assert.Len(t, mem.contents[fes.NewAggregate("type", "id")], 2)
}

func TestBackend_GetMultiple(t *testing.T) {
	mem := NewBackend()
	key := fes.NewAggregate("type", "id")
	events := []*fes.Event{
		fes.NewEvent(key, []byte("event 1")),
		fes.NewEvent(key, []byte("event 2")),
		fes.NewEvent(key, []byte("event 3")),
	}

	for k := range events {
		err := mem.Append(events[k])
		assert.NoError(t, err)
	}

	getEvents, err := mem.Get(&key)
	assert.NoError(t, err)
	assert.EqualValues(t, events, getEvents)
}

func TestBackend_GetNonexistent(t *testing.T) {
	mem := NewBackend()
	key := fes.NewAggregate("type", "id")
	getEvents, err := mem.Get(&key)
	assert.NoError(t, err)
	assert.EqualValues(t, []*fes.Event{}, getEvents)
}

func TestBackend_Subscribe(t *testing.T) {
	mem := NewBackend()
	key := fes.NewAggregate("type", "id")
	sub := mem.Subscribe(pubsub.SubscriptionOptions{
		LabelMatcher: labels.In(fes.PubSubLabelAggregateType, key.Type),
	})

	events := []*fes.Event{
		fes.NewEvent(key, []byte("event 1")),
		fes.NewEvent(key, []byte("event 2")),
		fes.NewEvent(key, []byte("event 3")),
	}
	for k := range events {
		err := mem.Append(events[k])
		assert.NoError(t, err)
	}
	mem.Unsubscribe(sub)

	var receivedEvents []*fes.Event
	for msg := range sub.Ch {
		event, ok := msg.(*fes.Event)
		assert.True(t, ok)
		receivedEvents = append(receivedEvents, event)
	}
	assert.EqualValues(t, events, receivedEvents)
}
