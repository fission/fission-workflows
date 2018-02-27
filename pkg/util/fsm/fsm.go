package fsm

import (
	"errors"
	"sync"
)

// Fsm is a read-only FSM definition
type Fsm struct {
	Initial     interface{}
	Transitions map[interface{}][]Transition
}

type Transition struct {
	Event interface{}
	Src   interface{}
	Dst   interface{}
}

func New(initial interface{}, transitions []Transition) *Fsm {
	// TODO check if comparable types and consistent
	nodeTransitions := map[interface{}][]Transition{}
	for _, t := range transitions {
		existing, ok := nodeTransitions[t.Src]
		if ok {
			nodeTransitions[t.Src] = append(existing, t)
		} else {
			nodeTransitions[t.Src] = []Transition{t}
		}

		_, ok = nodeTransitions[t.Dst]
		if !ok {
			nodeTransitions[t.Dst] = []Transition{}
		}
	}

	f := &Fsm{
		Initial:     initial,
		Transitions: nodeTransitions,
	}

	// Check if initial is valid
	if !f.NodeExists(initial) {
		panic("Initial does not exist!")
	}

	return f
}

func (f *Fsm) NodeExists(i interface{}) bool {
	_, ok := f.Transitions[i]
	return ok
}

func (f *Fsm) GetTransition(src, dst interface{}) *Transition {
	for _, t := range f.Transitions[src] {
		if t.Dst == dst {
			return &t
		}
	}
	return nil
}

func (f *Fsm) NewInstance(overrideInitial ...interface{}) FsmInstance {
	initial := f.Initial
	if len(overrideInitial) > 0 {
		initial = overrideInitial[0]
	}

	return FsmInstance{
		Fsm:     f,
		mx:      sync.RWMutex{},
		current: initial,
	}
}

type FsmInstance struct {
	*Fsm
	mx      sync.RWMutex
	current interface{}
}

func (f *FsmInstance) CanTransitionTo(node interface{}) bool {
	f.mx.RLock()
	defer f.mx.RLock()

	if !f.Fsm.NodeExists(node) {
		return false
	}

	for _, t := range f.Fsm.Transitions[f.current] {
		if t.Dst == node {
			return true
		}
	}
	return false
}

func (f *FsmInstance) TransitionTo(node interface{}) error {
	if !f.CanTransitionTo(node) {
		return errors.New("cannot transition to next")
	}

	f.mx.Lock()
	defer f.mx.Unlock()
	f.current = node
	return nil
}

func (f *FsmInstance) Evaluate(input interface{}) error {
	f.mx.Lock()
	defer f.mx.Unlock()

	for _, t := range f.Transitions[f.current] {
		if t.Event == input {
			f.current = t.Dst
		}
	}
	return nil
}

func (f *FsmInstance) Current() interface{} {
	f.mx.RLock()
	defer f.mx.RLock()
	return f.current
}
