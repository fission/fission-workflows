package gopool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGoPool(t *testing.T) {
	size := 100
	ctx := context.Background()
	pool := New(5)
	// Overflow pool
	wg := sync.WaitGroup{}
	var sum int32
	wg.Add(size)
	assert.Equal(t, int64(0), pool.Active())
	for i := 1; i <= size; i++ {
		a := i
		err := pool.Submit(ctx, func() {
			atomic.AddInt32(&sum, int32(a))
			wg.Done()
		})
		assert.NoError(t, err)
	}
	wg.Wait()
	assert.Equal(t, int32(size*(size+1)/2), sum)
	assert.Equal(t, int64(0), pool.Active())
}

func TestGoPoolClose(t *testing.T) {
	var counter int32 = 0
	pool := New(5)
	err := pool.Submit(context.Background(), func() {
		time.Sleep(100 * time.Millisecond)
		atomic.AddInt32(&counter, 1)
	})
	assert.NoError(t, err)
	err = pool.Close()
	assert.NoError(t, err)
	err = pool.Submit(context.Background(), func() {
		atomic.AddInt32(&counter, 1)
	})
	assert.Error(t, err, ErrPoolClosed.Error())
	assert.Equal(t, int64(0), pool.Active())
	assert.Equal(t, int32(1), counter)
}
