package gopool

import (
	"context"
	"sync/atomic"

	"golang.org/x/sync/semaphore"
)

type GoPool struct {
	routines       *semaphore.Weighted
	activeRoutines int64
	maxRoutines    int64
}

func New(maxRoutines int64) *GoPool {
	return &GoPool{
		routines: semaphore.NewWeighted(maxRoutines),
	}
}

func (g *GoPool) Active() int64 {
	return g.activeRoutines
}

func (g *GoPool) Submit(ctx context.Context, fn func()) error {
	if err := g.routines.Acquire(ctx, 1); err != nil {
		return err
	}
	go g.wrapRoutine(fn)
	return nil
}

func (g *GoPool) wrapRoutine(fn func()) {
	atomic.AddInt64(&g.activeRoutines, 1)
	fn()
	atomic.AddInt64(&g.activeRoutines, -1)
	g.routines.Release(1)
}
