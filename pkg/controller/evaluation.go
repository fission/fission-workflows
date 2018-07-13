package controller

import (
	"container/heap"
	"fmt"
	"sync"
	"time"
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

type heapCmdType string

const (
	heapCmdPush     heapCmdType = "push"
	heapCmdFront    heapCmdType = "front"
	heapCmdPop      heapCmdType = "pop"
	heapCmdUpdate   heapCmdType = "update"
	heapCmdGet      heapCmdType = "get"
	heapCmdLength   heapCmdType = "len"
	DefaultPriority             = 0
)

type heapCmd struct {
	cmd    heapCmdType
	input  interface{}
	result chan<- interface{}
}

type ConcurrentEvalStateHeap struct {
	heap             *EvalStateHeap
	cmdChan          chan heapCmd
	closeChan        chan bool
	queueChan        chan *EvalState
	init             sync.Once
	activeGoRoutines sync.WaitGroup
}

func NewConcurrentEvalStateHeap(unique bool) *ConcurrentEvalStateHeap {
	h := &ConcurrentEvalStateHeap{
		heap:      NewEvalStateHeap(unique),
		cmdChan:   make(chan heapCmd, 50),
		closeChan: make(chan bool),
		queueChan: make(chan *EvalState),
	}
	h.Init()
	heap.Init(h.heap)
	return h
}

func (h *ConcurrentEvalStateHeap) Init() {
	h.init.Do(func() {
		front := make(chan *EvalState, 1)
		updateFront := func() {
			front <- h.heap.Front().GetEvalState()
		}

		// Channel supplier
		go func() {
			h.activeGoRoutines.Add(1)
			var next *EvalState
			for {
				if next != nil {
					select {
					case h.queueChan <- next:
						// Queue item has been fetched; get next one.
						next = nil
						h.Pop()
					case u := <-front:
						// There has been an update to the queue item; replace it.
						next = u
					case <-h.closeChan:
						h.activeGoRoutines.Done()
						return
					}
				} else {
					// There is no current queue item, so we only listen to updates
					select {
					case u := <-front:
						next = u
					case <-h.closeChan:
						h.activeGoRoutines.Done()
						return
					}
				}
			}
		}()

		// Command handler
		go func() {
			h.activeGoRoutines.Add(1)
			updateFront()
			for {
				select {
				case cmd := <-h.cmdChan:
					switch cmd.cmd {
					case heapCmdFront:
						cmd.result <- h.heap.Front()
					case heapCmdLength:
						cmd.result <- h.heap.Len()
					case heapCmdPush:
						heap.Push(h.heap, cmd.input)
						updateFront()
					case heapCmdPop:
						cmd.result <- heap.Pop(h.heap)
						updateFront()
					case heapCmdGet:
						s, _ := h.heap.Get(cmd.input.(string))
						cmd.result <- s
					case heapCmdUpdate:
						i := cmd.input.(*HeapItem)
						if i.index < 0 {
							cmd.result <- h.heap.Update(i.EvalState)
						} else {
							cmd.result <- h.heap.UpdatePriority(i.EvalState, i.Priority)
						}
						updateFront()
					}
				case <-h.closeChan:
					h.activeGoRoutines.Done()
					return
				}
			}
		}()
	})
}

func (h *ConcurrentEvalStateHeap) Get(key string) *HeapItem {
	result := make(chan interface{})
	h.cmdChan <- heapCmd{
		cmd:    heapCmdGet,
		input:  key,
		result: result,
	}
	return (<-result).(*HeapItem)
}

func (h *ConcurrentEvalStateHeap) Chan() <-chan *EvalState {
	h.Len()
	return h.queueChan
}

func (h *ConcurrentEvalStateHeap) Len() int {
	result := make(chan interface{})
	h.cmdChan <- heapCmd{
		cmd:    heapCmdLength,
		result: result,
	}
	return (<-result).(int)
}

func (h *ConcurrentEvalStateHeap) Front() *HeapItem {
	result := make(chan interface{})
	h.cmdChan <- heapCmd{
		cmd:    heapCmdFront,
		result: result,
	}
	return (<-result).(*HeapItem)
}

func (h *ConcurrentEvalStateHeap) Update(s *EvalState) *HeapItem {
	if s == nil {
		return nil
	}
	result := make(chan interface{})
	h.cmdChan <- heapCmd{
		cmd:    heapCmdUpdate,
		result: result,
		input: &HeapItem{
			EvalState: s,
			index:     -1, // abuse index as a signal to not update priority
		},
	}
	return (<-result).(*HeapItem)
}

func (h *ConcurrentEvalStateHeap) UpdatePriority(s *EvalState, priority int) *HeapItem {
	if s == nil {
		return nil
	}
	result := make(chan interface{})
	h.cmdChan <- heapCmd{
		cmd:    heapCmdUpdate,
		result: result,
		input: &HeapItem{
			EvalState: s,
			Priority:  priority,
		},
	}
	return (<-result).(*HeapItem)
}

func (h *ConcurrentEvalStateHeap) Push(s *EvalState) {
	if s == nil {
		return
	}
	h.cmdChan <- heapCmd{
		cmd:   heapCmdPush,
		input: s,
	}
}

func (h *ConcurrentEvalStateHeap) PushPriority(s *EvalState, priority int) {
	if s == nil {
		return
	}
	h.cmdChan <- heapCmd{
		cmd: heapCmdPush,
		input: &HeapItem{
			EvalState: s,
			Priority:  priority,
		},
	}
}

func (h *ConcurrentEvalStateHeap) Pop() *EvalState {
	result := make(chan interface{})
	h.cmdChan <- heapCmd{
		cmd:    heapCmdPop,
		result: result,
	}
	res := <-result
	if res == nil {
		return nil
	}
	return (res).(*EvalState)
}

func (h *ConcurrentEvalStateHeap) Close() error {
	h.closeChan <- true
	close(h.closeChan)
	h.activeGoRoutines.Wait()
	close(h.cmdChan)
	close(h.queueChan)
	return nil
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
	return &EvalStateHeap{
		items:  map[string]*HeapItem{},
		unique: unique,
	}
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

// Use heap.Push
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

// Use heap.Pop
func (h *EvalStateHeap) Pop() interface{} {
	if h.Len() == 0 {
		return nil
	}
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

func (h *EvalStateHeap) Get(key string) (*HeapItem, bool) {
	item, ok := h.items[key]
	return item, ok
}

func ignoreOk(r EvalRecord, _ bool) EvalRecord {
	return r
}
