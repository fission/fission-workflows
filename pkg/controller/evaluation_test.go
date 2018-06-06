package controller

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var dummyRecord = EvalRecord{
	Timestamp: time.Now(),
	Error:     errors.New("other error"),
	RulePath:  []string{},
}

func TestEvalLog_Append(t *testing.T) {
	var log EvalLog
	record1 := dummyRecord
	log.Record(record1)
	record2 := EvalRecord{
		Timestamp: time.Now(),
		Error:     errors.New("stub error"),
		RulePath:  []string{"a", "b", "c"},
	}
	log.Record(record2)

	assert.Equal(t, 2, log.Len())
	last, ok := log.Last()
	assert.True(t, ok)
	assert.EqualValues(t, record2, last)

}

func TestEvalState_Lock(t *testing.T) {
	es := NewEvalState("id")

	<-es.Lock()
	select {
	case <-es.Lock():
		assert.Fail(t, "test was able to lock EvalState twice.")
	default:
		// ok
	}
}

func TestEvalState_Free(t *testing.T) {
	es := NewEvalState("id")

	es.Free() // Idempotent

	select {
	case <-es.Lock():
		// ok
	default:
		assert.Fail(t, "test failed to lock free EvalState twice.")
	}
	es.Free()
	select {
	case <-es.Lock():
		// ok
	default:
		assert.Fail(t, "test failed to lock free EvalState twice.")
	}
}

func TestEvalState_First(t *testing.T) {
	es := NewEvalState("id")
	assert.Equal(t, "id", es.ID())

	// Test non-existent
	r, ok := es.First()
	assert.False(t, ok)
	assert.Equal(t, EvalRecord{}, r)

	// Test existent
	es.Record(dummyRecord)
	r, ok = es.First()
	assert.True(t, ok)
	assert.Equal(t, dummyRecord, r)
}

func TestEvalState_Last(t *testing.T) {
	es := NewEvalState("id")
	assert.Equal(t, "id", es.ID())

	// Test non-existent
	r, ok := es.Last()
	assert.False(t, ok)
	assert.Equal(t, EvalRecord{}, r)

	// Test existent
	es.Record(dummyRecord)
	r, ok = es.Last()
	assert.True(t, ok)
	assert.Equal(t, dummyRecord, r)
}

func TestEvalState_Count(t *testing.T) {
	es := NewEvalState("id")
	assert.Equal(t, "id", es.ID())

	c := es.Len()
	assert.Equal(t, 0, c)

	es.Record(dummyRecord)
	c = es.Len()
	assert.Equal(t, 1, c)
}

func TestEvalState_Logs(t *testing.T) {
	es := NewEvalState("id")
	assert.Equal(t, "id", es.ID())

	// Test non-existent
	el := es.Logs()
	assert.Equal(t, EvalLog{}, el)

	// Test existent
	es.Record(dummyRecord)
	el = es.Logs()
	assert.Equal(t, EvalLog{dummyRecord}, el)
}

func TestEvalCache_GetOrCreate(t *testing.T) {
	ec := EvalStore{}
	id := "foo"
	es, ok := ec.Load(id)
	assert.False(t, ok)
	assert.Empty(t, es)

	es = ec.LoadOrStore(id)
	assert.Equal(t, id, es.ID())

	es, ok = ec.Load(id)
	assert.True(t, ok)
	assert.Equal(t, id, es.ID())
}

func TestEvalCache_Invalidate(t *testing.T) {
	ec := EvalStore{}
	id := "completedId"

	ec.Store(NewEvalState(id))
	es, ok := ec.Load(id)
	assert.True(t, ok)
	assert.Equal(t, id, es.ID())

	ec.Delete(id)
	es, ok = ec.Load(id)
	assert.False(t, ok)
	assert.Empty(t, es)
}

func TestConcurrentEvalStateHeap(t *testing.T) {
	h := NewConcurrentEvalStateHeap(true)
	defer h.Close()
	es1 := NewEvalState("id1")
	es2 := NewEvalState("id2")
	es3 := NewEvalState("id3")
	es1.Record(EvalRecord{
		Timestamp: time.Now().Add(2 * time.Minute),
	})
	es2.Record(EvalRecord{
		Timestamp: time.Now().Add(5 * time.Minute),
	})
	es3.Record(EvalRecord{
		Timestamp: time.Now().Add(1 * time.Minute),
	})
	h.Push(es1)
	h.Push(es2)
	h.Push(es3)

	assert.Equal(t, 3, h.Len())
	assert.Equal(t, es3, h.Pop())
	assert.Equal(t, es1, h.Pop())
	assert.Equal(t, es2, h.Pop())
}

func TestConcurrentEvalStateHeap_priorities(t *testing.T) {
	h := NewConcurrentEvalStateHeap(true)
	defer h.Close()
	es1 := NewEvalState("id1")
	es2 := NewEvalState("id2")
	es3 := NewEvalState("id3")
	es1.Record(EvalRecord{
		Timestamp: time.Now().Add(2 * time.Minute),
	})
	es2.Record(EvalRecord{
		Timestamp: time.Now().Add(5 * time.Minute),
	})
	es3.Record(EvalRecord{
		Timestamp: time.Now().Add(1 * time.Minute),
	})
	h.Push(es1)
	h.PushPriority(es2, 1)
	h.Push(es3)

	assert.Equal(t, es2, h.Pop())
	assert.Equal(t, es3, h.Pop())
	assert.Equal(t, es1, h.Pop())
}

func TestConcurrentEvalStateHeap_updateTimestamp(t *testing.T) {
	h := NewConcurrentEvalStateHeap(true)
	defer h.Close()
	es1 := NewEvalState("id1")
	es1.Record(EvalRecord{
		Timestamp: time.Now().Add(2 * time.Minute),
	})
	es2 := NewEvalState("id2")
	es2.Record(EvalRecord{
		Timestamp: time.Now().Add(5 * time.Minute),
	})
	h.Push(es1)
	h.Push(es2)

	// New record for es1
	es1.Record(EvalRecord{
		Timestamp: time.Now().Add(10 * time.Minute),
	})
	h.Update(es1)
	assert.Equal(t, es2, h.Pop())
	assert.Equal(t, es1, h.Pop())
}

func TestConcurrentEvalStateHeap_updatePriority(t *testing.T) {
	h := NewConcurrentEvalStateHeap(true)
	defer h.Close()
	es1 := NewEvalState("id1")
	es1.Record(EvalRecord{
		Timestamp: time.Now().Add(2 * time.Minute),
	})
	es2 := NewEvalState("id2")
	es2.Record(EvalRecord{
		Timestamp: time.Now().Add(5 * time.Minute),
	})
	h.Push(es1)
	h.Push(es2)
	assert.Equal(t, 2, h.Len())

	// Priority changes for es2
	h.UpdatePriority(es2, 1)
	assert.Equal(t, es2, h.Pop())
	assert.Equal(t, es1, h.Pop())
}

func TestConcurrentEvalStateHeap_chan(t *testing.T) {
	h := NewConcurrentEvalStateHeap(true)
	defer h.Close()
	es1 := NewEvalState("id1")
	es2 := NewEvalState("id2")
	es3 := NewEvalState("id3")
	es1.Record(EvalRecord{
		Timestamp: time.Now().Add(2 * time.Minute),
	})
	es2.Record(EvalRecord{
		Timestamp: time.Now().Add(5 * time.Minute),
	})
	es3.Record(EvalRecord{
		Timestamp: time.Now().Add(1 * time.Minute),
	})
	h.Push(es1)
	h.Push(es2)
	h.Push(es3)

	c := h.Chan()
	item := h.Front()
	assert.Equal(t, es3.id, item.id)
	assert.Equal(t, es3, <-c)
	assert.Equal(t, es1, <-c)
	assert.Equal(t, es2, <-c)
	select {
	case <-c:
		assert.Fail(t, "unexpected item")
	default:
	}
}
