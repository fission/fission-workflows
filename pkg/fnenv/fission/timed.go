package fission

import (
	"container/heap"
	"sync"
	"time"
)

// timedExecPool provides a data structure for scheduling executions based on a timestamp.
type timedExecPool struct {
	fnQueue *timedFnQueue
	cancel  chan struct{}
	fnsLock *sync.Mutex
}

func newTimedExecPool() *timedExecPool {
	return &timedExecPool{
		fnQueue: &timedFnQueue{},
		cancel:  make(chan struct{}),
		fnsLock: &sync.Mutex{},
	}
}

func (ds *timedExecPool) Submit(fn func(), execAt time.Time) {
	ds.fnsLock.Lock()
	defer ds.fnsLock.Unlock()
	ds.fnQueue.Push(timedFn{
		execAt: execAt,
		fn:     fn,
	})
	heap.Init(ds.fnQueue)
	ds.eval()
}

func (ds *timedExecPool) Eval() {
	ds.fnsLock.Lock()
	defer ds.fnsLock.Unlock()
	ds.eval()
}

func (ds *timedExecPool) eval() {
	// Get head
	t := ds.fnQueue.Peek()
	if t == nil {
		return
	}

	// Check if it can be executed now
	if time.Now().After(t.execAt) {
		it := ds.fnQueue.Pop()
		if it != nil {
			fn := it.(*timedFn).fn
			if fn != nil {
				fn()
			}
		}
		return
	}

	// Cancel
	select {
	case ds.cancel <- struct{}{}:
	default:
	}
	close(ds.cancel)

	// Wait for
	go func() {
		ds.cancel = make(chan struct{})
		// Either wait until the time has been reached or the wait is canceled
		select {
		case <-time.After(time.Until(t.execAt)):
			ds.Eval()
			return
		case <-ds.cancel:
			return
		}
	}()
}

type timedFn struct {
	execAt time.Time
	fn     func()
}

type timedFnQueue struct {
	fns []*timedFn
}

func (tf *timedFnQueue) Len() int { return len(tf.fns) }

func (tf *timedFnQueue) Less(i, j int) bool {
	return tf.fns[i].execAt.Before(tf.fns[j].execAt)
}

func (tf *timedFnQueue) Swap(i, j int) {
	tf.fns[i], tf.fns[j] = tf.fns[j], tf.fns[i]
}

func (tf *timedFnQueue) Push(x interface{}) {
	item := x.(*timedFn)
	tf.fns = append(tf.fns, item)
}

func (tf *timedFnQueue) Pop() interface{} {
	old := tf.fns
	n := len(old)
	item := old[n-1]
	tf.fns = old[0 : n-1]
	return item
}

func (tf *timedFnQueue) Peek() *timedFn {
	if len(tf.fns) == 0 {
		return nil
	}
	return tf.fns[0]
}
