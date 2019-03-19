/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package workqueue

import (
	"sync"
)

const (
	DefaultMaxSize = 10000
)

type Interface interface {
	Add(item interface{}) (accepted bool)
	Len() int
	Get() (item interface{}, shutdown bool)
	Done(item interface{})
	ShutDown()
	ShuttingDown() bool
}

func NewNamed(maxSize int, _ string) *Type {
	return NewWorkQueue(maxSize, false)
}

func NewWorkQueue(maxSize int, replace bool) *Type {
	return &Type{
		MaxSize:    maxSize,
		Replace:    replace,
		dirty:      make(map[interface{}]interface{}),
		processing: make(map[interface{}]interface{}),
		cond:       sync.NewCond(&sync.Mutex{}),
	}
}

// New constructs a new work queue (see the package comment).
func New() *Type {
	return NewWorkQueue(DefaultMaxSize, false)
}

// Type is a work queue (see the package comment).
type Type struct {
	MaxSize int
	Replace bool

	// queue defines the order in which we will work on items. Every
	// element of queue should be in the dirty set and not in the
	// processing set.
	queue []interface{}

	// dirty defines all of the items that need to be processed.
	dirty map[interface{}]interface{}

	// Things that are currently being processed are in the processing set.
	// These things may be simultaneously in the dirty set. When we finish
	// processing something and remove it from this set, we'll check if
	// it's in the dirty set, and if so, add it to the queue.
	processing map[interface{}]interface{}

	cond *sync.Cond

	shuttingDown bool
}

// Add marks item as needing processing.
func (q *Type) Add(item interface{}) (accepted bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	if q.shuttingDown {
		return false
	}

	key := getKey(item)
	if _, ok := q.dirty[key]; ok {
		if q.Replace {
			q.dirty[key] = item
		}
		return true
	}

	if len(q.queue) >= q.MaxSize {
		return false
	}

	q.dirty[key] = item
	if _, ok := q.processing[key]; ok {
		return true
	}

	q.queue = append(q.queue, key)
	q.cond.Signal()
	return true
}

// Len returns the current queue length, for informational purposes only. You
// shouldn't e.g. gate a call to Add() or Get() on Len() being a particular
// value, that can't be synchronized properly.
func (q *Type) Len() int {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	return len(q.queue)
}

// Get blocks until it can return an item to be processed. If shutdown = true,
// the caller should end their goroutine. You must call Done with item when you
// have finished processing it.
func (q *Type) Get() (item interface{}, shutdown bool) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	for len(q.queue) == 0 && !q.shuttingDown {
		q.cond.Wait()
	}
	if len(q.queue) == 0 {
		// We must be shutting down.
		return nil, true
	}

	var key interface{}
	key, q.queue = q.queue[0], q.queue[1:]
	item = q.dirty[key]
	q.processing[key] = item
	delete(q.dirty, key)

	return item, false
}

// Done marks item as done processing, and if it has been marked as dirty again
// while it was being processed, it will be re-added to the queue for
// re-processing.
func (q *Type) Done(item interface{}) {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	key := getKey(item)
	delete(q.processing, key)
	if _, ok := q.dirty[key]; ok {
		q.queue = append(q.queue, key)
		q.cond.Signal()
	}
}

// ShutDown will cause q to ignore all new items added to it. As soon as the
// worker goroutines have drained the existing items in the queue, they will be
// instructed to exit.
func (q *Type) ShutDown() {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()
	q.shuttingDown = true
	q.cond.Broadcast()
}

func (q *Type) ShuttingDown() bool {
	q.cond.L.Lock()
	defer q.cond.L.Unlock()

	return q.shuttingDown
}

type Identifier interface {
	ID() interface{}
}

func getKey(item interface{}) interface{} {
	if identifier, ok := item.(Identifier); ok && identifier.ID() != nil {
		return identifier.ID()
	} else {
		return item
	}
}
