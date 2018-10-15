// package gopool provides functionality for bounded parallelism with goroutines
package gopool

import (
	"context"
	"errors"
	"fmt"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"
)

const (
	CloseGracePeriod = 5 * time.Second
)

var (
	ErrPoolClosed = errors.New("gopool is closed")
)

// GoPool is a structure to provide bounded parallelism with goroutines.
type GoPool struct {
	routines       *semaphore.Weighted
	activeRoutines int64
	maxRoutines    int64
	stopped        uint32
}

// New creates a new GoPool with the given size, where size > 0.
// The size indicates the maximum number of workers that can be executed in parallel at any moment.
func New(size int64) *GoPool {
	if size <= 0 {
		panic(fmt.Sprintf("invalid GoPool size: %d", size))
	}
	return &GoPool{
		maxRoutines: size,
		routines:    semaphore.NewWeighted(size),
	}
}

// Max returns the max workers executing in parallel in this pool.
func (g *GoPool) Max() int64 {
	return g.maxRoutines
}

// Active returns the currently active workers executing in this pool.
func (g *GoPool) Active() int64 {
	return atomic.LoadInt64(&g.activeRoutines)
}

// Submit submits a function to be executed in the pool.
// If the pool has been closed, this function will return ErrPoolClosed.
// If the pool is active, this function will block until either of the following is true:
// (1) there is space in this pool to execute the function, returning nil.
// (2) the context has completed, resulting in a return value of ctx.Err().
func (g *GoPool) Submit(ctx context.Context, fn func()) error {
	if atomic.LoadUint32(&g.stopped) != 0 {
		return ErrPoolClosed
	}
	if err := g.routines.Acquire(ctx, 1); err != nil {
		return err
	}
	go g.wrapRoutine(fn)
	return nil
}

// GracefulStop closes the pool for any new work, and waits for the current functions to finish, or
// until the context completes.
func (g *GoPool) GracefulStop(ctx context.Context) error {
	atomic.StoreUint32(&g.stopped, 1)
	return g.routines.Acquire(ctx, g.maxRoutines)
}

// Close closes using GracefulStop with the default CloseGracePeriod, to conform with the io.Closer interface.
func (g *GoPool) Close() error {
	ctx, cancelFn := context.WithTimeout(context.Background(), CloseGracePeriod)
	defer cancelFn()
	return g.GracefulStop(ctx)
}

func (g *GoPool) wrapRoutine(fn func()) {
	atomic.AddInt64(&g.activeRoutines, 1)
	fn()
	atomic.AddInt64(&g.activeRoutines, -1)
	g.routines.Release(1)
}
