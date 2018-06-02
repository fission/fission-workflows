package controller

import (
	"sync"
	"time"
)

// EvalCache allows storing and retrieving EvalStates in a thread-safe way.
type EvalCache struct {
	states map[string]*EvalState
	lock   sync.RWMutex
}

func NewEvalCache() *EvalCache {
	return &EvalCache{
		states: map[string]*EvalState{},
	}
}

func (e *EvalCache) GetOrCreate(id string) *EvalState {
	s, ok := e.Get(id)
	if !ok {
		s = NewEvalState(id)
		e.Put(s)
	}
	return s
}

func (e *EvalCache) Get(id string) (*EvalState, bool) {
	e.lock.RLock()
	s, ok := e.states[id]
	e.lock.RUnlock()
	return s, ok
}

func (e *EvalCache) Put(state *EvalState) {
	e.lock.Lock()
	e.states[state.id] = state
	e.lock.Unlock()
}

func (e *EvalCache) Del(id string) {
	e.lock.Lock()
	delete(e.states, id)
	e.lock.Unlock()
}

func (e *EvalCache) List() map[string]*EvalState {
	results := map[string]*EvalState{}
	e.lock.RLock()
	for id, state := range e.states {
		results[id] = state
	}
	e.lock.RUnlock()
	return results
}

// EvalState is the state of a specific object that is evaluated in the controller.
//
// TODO add logger / or helper to log / context
// TODO add a time before next evaluation -> backoff
// TODO add current/in progress record
type EvalState struct {
	// id is the identifier of the evaluation. For example the invocation.
	id string

	// EvalLog keep track of previous evaluations of this resource
	log EvalLog

	// evalLock allows gaining exclusive access to this evaluation
	evalLock chan struct{}

	// dataLock ensures thread-safe read and writes to this state. For example appending and reading logs.
	dataLock sync.RWMutex
}

func NewEvalState(id string) *EvalState {
	e := &EvalState{
		log:      EvalLog{},
		id:       id,
		evalLock: make(chan struct{}, 1),
	}
	e.Free()
	return e
}

// Lock returns the single-buffer lock channel. A consumer has obtained exclusive access to this evaluation if it
// receives the element from the channel. Compared to native locking, this allows consumers to have option to implement
// backup logic in case an evaluation is locked.
//
// Example: `<- es.Lock()`
func (e *EvalState) Lock() chan struct{} {
	return e.evalLock
}

// Free releases the obtained exclusive access to this evaluation. In case the evaluation is already free, this function
// is a nop.
func (e *EvalState) Free() {
	select {
	case e.evalLock <- struct{}{}:
	default:
		// was already unlocked
	}
}

func (e *EvalState) ID() string {
	return e.id
}

func (e *EvalState) Count() int {
	e.dataLock.RLock()
	defer e.dataLock.RUnlock()
	return len(e.log)
}

func (e *EvalState) Get(i int) (EvalRecord, bool) {
	e.dataLock.RLock()
	defer e.dataLock.RUnlock()
	if i >= len(e.log) {
		return EvalRecord{}, false
	}
	return e.log[i], true
}

func (e *EvalState) Last() (EvalRecord, bool) {
	e.dataLock.RLock()
	defer e.dataLock.RUnlock()
	return e.log.Last()
}

func (e *EvalState) First() (EvalRecord, bool) {
	e.dataLock.RLock()
	defer e.dataLock.RUnlock()
	return e.log.First()
}

func (e *EvalState) Logs() EvalLog {
	e.dataLock.RLock()
	defer e.dataLock.RUnlock()
	logs := make(EvalLog, len(e.log))
	copy(logs, e.log)
	return logs
}

func (e *EvalState) Record(record EvalRecord) {
	e.dataLock.Lock()
	e.log.Record(record)
	e.dataLock.Unlock()
}

// EvalRecord contains all metadata related to a single evaluation of a controller.
type EvalRecord struct {
	// Timestamp is the time at which the evaluation started. As an evaluation should not take any significant amount
	// of time the evaluation is assumed to have occurred at a point in time.
	Timestamp time.Time

	// Cause is the reason why this evaluation was triggered. For example: 'tick' or 'notification' (optional).
	Cause string

	// Action contains the action that the evaluation resulted in, if any.
	Action Action

	// Error contains the error that the evaluation resulted in, if any.
	Error error

	// RulePath contains all the rules that were evaluated in order to complete the evaluation.
	RulePath []string
}

func NewEvalRecord() EvalRecord {
	return EvalRecord{
		Timestamp: time.Now(),
	}
}

// EvalLog is a time-ordered log of evaluation records. Newer records are appended to the end of the log.
type EvalLog []EvalRecord

func (e EvalLog) Count() int {
	return len(e)
}

func (e EvalLog) Last() (EvalRecord, bool) {
	if e.Count() == 0 {
		return EvalRecord{}, false
	}
	return e[len(e)-1], true
}

func (e EvalLog) First() (EvalRecord, bool) {
	if e.Count() == 0 {
		return EvalRecord{}, false
	}
	return e[0], true
}

func (e *EvalLog) Record(record EvalRecord) {
	*e = append(*e, record)
}
