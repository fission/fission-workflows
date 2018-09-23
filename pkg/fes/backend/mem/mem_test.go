package mem

import (
	"fmt"
	"math"
	"testing"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/util/labels"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/stretchr/testify/assert"
)

func newEvent(a fes.Aggregate, data []byte) *fes.Event {
	event, err := fes.NewEvent(a, &wrappers.BytesValue{
		Value: data,
	})
	if err != nil {
		panic(err)
	}
	return event
}

func setupBackend() *Backend {
	return NewBackend()
}

// 2018-09-20: 0.5 ms
func BenchmarkRoundtripSingleKey(b *testing.B) {
	mem := setupBackend()
	sub := mem.Subscribe()

	// Generate test data
	events := make([]*fes.Event, b.N)
	for i := 0; i < b.N; i++ {
		key := fes.Aggregate{Type: "type", Id: "id"}
		events[i] = newEvent(key, []byte(fmt.Sprintf("event-%d", i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := mem.Append(events[i])
		if err != nil {
			panic(err)
		}
		<-sub.Ch
	}
}

// 2018-09-20: 2 ms
func BenchmarkRoundtripManyKeys(b *testing.B) {
	mem := setupBackend()
	sub := mem.Subscribe()

	// Generate test data
	events := make([]*fes.Event, b.N)
	for i := 0; i < b.N; i++ {
		key := fes.Aggregate{Type: "type", Id: fmt.Sprintf("%d", i)}
		events[i] = newEvent(key, []byte(fmt.Sprintf("event-%d", i)))
	}

	// Run benchmark based on roundtrip
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := mem.Append(events[i])
		if err != nil {
			panic(err)
		}
		<-sub.Ch
	}
}

// 2018-09-20: 2-3 ms
func BenchmarkRoundtripManyEvictions(b *testing.B) {
	mem := setupBackend()
	sub := mem.Subscribe()

	mem.MaxKeys = int(math.Max(float64(b.N)/100, 1))

	// Generate test data
	events := make([]*fes.Event, b.N)
	for i := 0; i < b.N; i++ {
		key := fes.Aggregate{Type: "entity", Id: fmt.Sprintf("%d", i)}
		events[i] = newEvent(key, []byte(fmt.Sprintf("event-%d", i)))
		events[i].Hints = &fes.EventHints{
			Completed: true,
		}
	}

	// Run benchmark based on roundtrip
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := mem.Append(events[i])
		if err != nil {
			panic(err)
		}
		<-sub.Ch
	}
}

func TestBackendStoreFull(t *testing.T) {
	mem := setupBackend()
	mem.MaxKeys = 2
	n := 3
	events := make([]*fes.Event, n)
	for i := 0; i < n; i++ {
		key := fes.Aggregate{Type: "entity", Id: fmt.Sprintf("%d", i)}
		events[i] = newEvent(key, []byte(fmt.Sprintf("event-%d", i)))
	}
	for _, event := range events {
		err := mem.Append(event)
		if err != nil {
			assert.EqualError(t, err, fes.ErrEventStoreOverflow.WithAggregate(events[2].Aggregate).Error())
		}
	}
}

func TestBackendBufferEvict(t *testing.T) {
	mem := setupBackend()
	mem.MaxKeys = 3
	n := 10
	events := make([]*fes.Event, n)
	mem.Append(newEvent(fes.Aggregate{Type: "active", Id: "1"}, []byte("active stream")))
	for i := 0; i < n; i++ {
		key := fes.Aggregate{Type: "entity", Id: fmt.Sprintf("%d", i)}
		events[i] = newEvent(key, []byte(fmt.Sprintf("event-%d", i)))
		events[i].Hints = &fes.EventHints{
			Completed: true,
		}
	}
	for _, event := range events {
		err := mem.Append(event)
		assert.NoError(t, err)
	}
	assert.Equal(t, 2, mem.buf.Len())
	assert.Equal(t, 1, len(mem.store))
	assert.Equal(t, 3, mem.Len())
}

func TestBackend_Append(t *testing.T) {
	mem := setupBackend()

	event := newEvent(fes.Aggregate{Type: "type", Id: "id"}, []byte("event 1"))
	err := mem.Append(event)
	assert.NoError(t, err)
	assert.Equal(t, mem.Len(), 1)

	// Test if invalid event is rejected by the event store
	event2 := newEvent(fes.Aggregate{Type: "type", Id: "id"}, []byte("event 1"))
	event2.Aggregate = &fes.Aggregate{}
	err = mem.Append(event2)
	assert.EqualError(t, err, fes.ErrInvalidAggregate.Error())
	assert.Equal(t, mem.Len(), 1)

	// Event under existing aggregate
	event3, err := fes.NewEvent(fes.Aggregate{Type: "type", Id: "id"}, &wrappers.BytesValue{
		Value: []byte("event 2"),
	})
	assert.NoError(t, err)
	err = mem.Append(event3)
	assert.NoError(t, err)
	assert.Equal(t, mem.Len(), 1)
	assert.Equal(t, len(mem.mustGet(fes.Aggregate{Type: "type", Id: "id"})), 2)

	// Event under new aggregate
	event4, err := fes.NewEvent(fes.Aggregate{Type: "Type", Id: "other"}, &wrappers.BytesValue{
		Value: []byte("event 1"),
	})
	assert.NoError(t, err)
	err = mem.Append(event4)
	assert.NoError(t, err)
	assert.Equal(t, mem.Len(), 2)
	assert.Equal(t, len(mem.mustGet(fes.Aggregate{Type: "Type", Id: "other"})), 1)
	assert.Equal(t, len(mem.mustGet(fes.Aggregate{Type: "type", Id: "id"})), 2)
}

func TestBackend_GetMultiple(t *testing.T) {
	mem := setupBackend()
	key := fes.Aggregate{Type: "type", Id: "id"}
	events := []*fes.Event{
		newEvent(key, []byte("event 1")),
		newEvent(key, []byte("event 2")),
		newEvent(key, []byte("event 3")),
	}

	for k := range events {
		err := mem.Append(events[k])
		assert.NoError(t, err)
	}

	getEvents, err := mem.Get(key)
	assert.NoError(t, err)
	assert.EqualValues(t, events, getEvents)
}

func TestBackend_GetNonexistent(t *testing.T) {
	mem := setupBackend()
	key := fes.Aggregate{Type: "type", Id: "id"}
	getEvents, err := mem.Get(key)
	assert.NoError(t, err)
	assert.EqualValues(t, []*fes.Event{}, getEvents)
}

func TestBackend_Subscribe(t *testing.T) {
	mem := setupBackend()
	key := fes.Aggregate{Type: "type", Id: "id"}
	sub := mem.Subscribe(pubsub.SubscriptionOptions{
		LabelMatcher: labels.In(fes.PubSubLabelAggregateType, key.Type),
	})

	events := []*fes.Event{
		newEvent(key, []byte("event 1")),
		newEvent(key, []byte("event 2")),
		newEvent(key, []byte("event 3")),
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

func (b *Backend) mustGet(key fes.Aggregate) []*fes.Event {
	val, ok, _ := b.get(key)
	if !ok {
		panic(fmt.Sprintf("expected value present for key %s", key))
	}
	return val
}
