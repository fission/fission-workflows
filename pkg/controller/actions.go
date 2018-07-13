package controller

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/fission/fission-workflows/pkg/fes"
)

// TODO remove from EvalCache actions

type ActionWait struct {
	EvalState *EvalState
	Wait      time.Duration
}

func (a *ActionWait) Apply() error {
	// TODO set flag in evalstate
	panic("not implemented")
}

func (a *ActionWait) Eval(rule EvalContext) Action {
	return a
}

type ActionSkip struct{}

func (a *ActionSkip) Apply() error {
	return nil
}

func (a *ActionSkip) Eval(rule EvalContext) Action {
	return a
}

type ActionRemoveFromEvalCache struct {
	EvalCache *EvalCache
	ID        string
}

func (a *ActionRemoveFromEvalCache) Apply() error {
	a.EvalCache.Del(a.ID)
	return nil
}

func (a *ActionRemoveFromEvalCache) Eval(rule EvalContext) Action {
	return a
}

type ActionRemoveFromFesCache struct {
	Aggregate fes.Aggregate
	Cache     fes.CacheWriter
}

func (a *ActionRemoveFromFesCache) Apply() error {
	a.Cache.Invalidate(&a.Aggregate)
	return nil
}

func (a *ActionRemoveFromFesCache) Eval(rule EvalContext) Action {
	return a
}

type ActionError struct {
	Err error
}

func (a *ActionError) Apply() error {
	return a.Err
}

func (a *ActionError) Eval(rule EvalContext) Action {
	return a
}

type MultiAction struct {
	Actions []Action
}

func (a *MultiAction) Apply() error {
	var wg sync.WaitGroup
	var multiErr atomic.Value
	wg.Add(len(a.Actions))
	for _, action := range a.Actions {
		go func(action Action) {
			err := action.Apply()
			if err != nil {
				multiErr.Store(err)
			}
			wg.Done()
		}(action)
	}
	wg.Wait()
	err := multiErr.Load()
	if err == nil {
		return nil
	}
	return err.(error)
}
