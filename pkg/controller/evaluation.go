package controller

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
)

const (
	DefaultPriority = 0
)

// EvalStore allows storing and retrieving EvalStates in a thread-safe way.
type EvalStore struct {
	mp sync.Map
}

func (e *EvalStore) LoadOrStore(id string) *EvalState {
	s, _ := e.mp.LoadOrStore(id, NewEvalState(id))
	return s.(*EvalState)
}

func (e *EvalStore) Load(id string) (*EvalState, bool) {
	s, ok := e.mp.Load(id)
	if !ok {
		return nil, false
	}
	return s.(*EvalState), true
}

func (e *EvalStore) Store(state *EvalState) {
	e.mp.Store(state.id, state)
}

func (e *EvalStore) Delete(id string) {
	e.mp.Delete(id)
}

func (e *EvalStore) List() map[string]*EvalState {
	results := map[string]*EvalState{}
	e.mp.Range(func(k, v interface{}) bool {
		results[k.(string)] = v.(*EvalState)
		return true
	})
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
	if e == nil {
		return ""
	}
	return e.id
}

func (e *EvalState) Len() int {
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

func (e EvalLog) Len() int {
	return len(e)
}

func (e EvalLog) Last() (EvalRecord, bool) {
	if e.Len() == 0 {
		return EvalRecord{}, false
	}
	return e[len(e)-1], true
}

func (e EvalLog) First() (EvalRecord, bool) {
	if e.Len() == 0 {
		return EvalRecord{}, false
	}
	return e[0], true
}

func (e *EvalLog) Record(record EvalRecord) {
	*e = append(*e, record)
}

type HeapItem struct {
	*EvalState
	Priority int
	index    int
}

func (h *HeapItem) GetEvalState() *EvalState {
	if h == nil {
		return nil
	}
	return h.EvalState
}

type EvalStateHeap struct {
	heap   []*HeapItem
	items  map[string]*HeapItem
	unique bool
}

func NewEvalStateHeap(unique bool) *EvalStateHeap {
	h := &EvalStateHeap{
		items:  map[string]*HeapItem{},
		unique: unique,
	}
	heap.Init(h)
	return h
}

func (h EvalStateHeap) Len() int {
	return len(h.heap)
}

func (h EvalStateHeap) Front() *HeapItem {
	if h.Len() == 0 {
		return nil
	}
	return h.heap[0]
}

func (h EvalStateHeap) Less(i, j int) bool {
	it := h.heap[i]
	jt := h.heap[j]

	// Check priorities (descending)
	if it.Priority > jt.Priority {
		return true
	} else if it.Priority < jt.Priority {
		return false
	}

	// If priorities are equal, compare timestamp (ascending)
	return ignoreOk(it.Last()).Timestamp.Before(ignoreOk(jt.Last()).Timestamp)
}

func (h EvalStateHeap) Swap(i, j int) {
	h.heap[i], h.heap[j] = h.heap[j], h.heap[i]
	h.heap[i].index = i
	h.heap[j].index = j
}

// Do not use directly, use heap.Push!
// The signature of push requests an interface{} to adhere to the sort.Interface interface, but will panic if a
// a type other than *EvalState is provided.
func (h *EvalStateHeap) Push(x interface{}) {
	switch t := x.(type) {
	case *EvalState:
		h.pushPriority(t, DefaultPriority)
	case *HeapItem:
		h.pushPriority(t.EvalState, t.Priority)
	default:
		panic(fmt.Sprintf("invalid entity submitted: %v", t))
	}
}

func (h *EvalStateHeap) pushPriority(state *EvalState, priority int) {
	if h.unique {
		if _, ok := h.items[state.id]; ok {
			h.UpdatePriority(state, priority)
			return
		}
	}
	el := &HeapItem{
		EvalState: state,
		Priority:  priority,
		index:     h.Len(),
	}
	h.heap = append(h.heap, el)
	h.items[state.id] = el
}

// Do not use directly, use heap.Pop!
func (h *EvalStateHeap) Pop() interface{} {
	popped := h.heap[h.Len()-1]
	delete(h.items, popped.id)
	h.heap = h.heap[:h.Len()-1]
	return popped.EvalState
}

func (h *EvalStateHeap) Update(es *EvalState) *HeapItem {
	if existing, ok := h.items[es.id]; ok {
		existing.EvalState = es
		heap.Fix(h, existing.index)
		return existing
	}
	return nil
}

func (h *EvalStateHeap) UpdatePriority(es *EvalState, priority int) *HeapItem {
	if existing, ok := h.items[es.id]; ok {
		existing.Priority = priority
		existing.EvalState = es
		heap.Fix(h, existing.index)
		return existing
	}
	return nil
}

func (h *EvalStateHeap) Get(key string) *HeapItem {
	item, _ := h.items[key]
	return item
}

func ignoreOk(r EvalRecord, _ bool) EvalRecord {
	return r
}
