package gopool

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"

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
