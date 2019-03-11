package controller

import (
	"time"

	"github.com/fission/fission-workflows/pkg/fes"
)

// TODO remove from EvalStore actions

type ActionWait struct {
	EvalState *EvalState
	Wait      time.Duration
}

func (a *ActionWait) Apply() error {
	// TODO set flag in evalstate
	panic("not implemented")
}

func (a *ActionWait) Eval(rule EvalContext) []Action {
	return []Action{a}
}

type ActionSkip struct{}

func (a *ActionSkip) Apply() error {
	return nil
}

func (a *ActionSkip) Eval(rule EvalContext) []Action {
	return []Action{a}
}

type ActionRemoveFromEvalCache struct {
	EvalCache *EvalStore
	ID        string
}

func (a *ActionRemoveFromEvalCache) Apply() error {
	a.EvalCache.Delete(a.ID)
	return nil
}

func (a *ActionRemoveFromEvalCache) Eval(rule EvalContext) []Action {
	return []Action{a}
}

type ActionRemoveFromFesCache struct {
	Aggregate fes.Aggregate
	Cache     fes.CacheWriter
}

func (a *ActionRemoveFromFesCache) Apply() error {
	a.Cache.Invalidate(a.Aggregate)
	return nil
}

func (a *ActionRemoveFromFesCache) Eval(rule EvalContext) []Action {
	return []Action{a}
}

type ActionError struct {
	Err error
}

func (a *ActionError) Apply() error {
	return a.Err
}

func (a *ActionError) Eval(rule EvalContext) []Action {
	return []Action{a}
}
