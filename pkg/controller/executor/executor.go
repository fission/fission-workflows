package executor

import (
	"fmt"
	"runtime/debug"
	"sync"
	"time"

	"github.com/fission/fission-workflows/pkg/util/workqueue"
	log "github.com/sirupsen/logrus"
)

type LocalExecutor struct {
	//
	// Config
	//
	maxParallelism int

	//
	// State
	//
	queue    workqueue.DelayingInterface
	workers  []*worker
	groups   map[interface{}]int
	groupsMu *sync.RWMutex
}

// Task is the unit of execution that the executor will execute.
type Task struct {
	// TaskID is used to ensure that there is only one instance of this task
	TaskID interface{}

	// GroupID is used to group together tasks.
	GroupID interface{}

	// Apply is the work that the task comprises.
	Apply func() error
}

func (t *Task) ID() interface{} {
	return t.TaskID
}

func NewLocalExecutor(maxParallelism, maxQueueSize int) *LocalExecutor {
	if maxParallelism <= 0 {
		panic("LocalExecutor: parallelism should be larger than 0")
	}
	if maxQueueSize <= 0 {
		panic("LocalExecutor: queue size should be larger than 0")
	}
	return &LocalExecutor{
		maxParallelism: maxParallelism,
		queue:          workqueue.NewDelayingQueue(maxQueueSize),
		groups:         make(map[interface{}]int),
		groupsMu:       &sync.RWMutex{},
	}
}

func (ex *LocalExecutor) Start() {
	if ex.maxParallelism <= 0 {
		panic("LocalExecutor: parallelism should be larger than 0")
	}
	// Add workers based on max parallelism
	for i := 0; i < ex.maxParallelism; i++ {
		worker := &worker{
			queue:    ex.queue,
			groups:   ex.groups,
			groupsMu: ex.groupsMu,
		}
		ex.workers = append(ex.workers, worker)
		go worker.Run()
	}
}

func (ex *LocalExecutor) Close() error {
	ex.queue.ShutDown()
	return nil
}

func (ex *LocalExecutor) GetGroupTasks(groupID interface{}) int {
	ex.groupsMu.RLock()
	count := ex.groups[groupID]
	ex.groupsMu.RUnlock()
	return count
}

func (ex *LocalExecutor) SubmitAfter(t *Task, after time.Duration) bool {
	// Add to the queue
	if after <= 0 {
		accepted := ex.queue.TryAddAfter(t, after)
		if !accepted {
			return false
		}
	} else {
		accepted := ex.queue.Add(t)
		if !accepted {
			return false
		}
	}

	// Increment the group
	if t.GroupID != nil {
		ex.groupsMu.Lock()
		ex.groups[t.GroupID]++
		ex.groupsMu.Unlock()
	}
	return true
}

func (ex *LocalExecutor) Submit(t *Task) bool {
	return ex.SubmitAfter(t, 0)
}

type worker struct {
	queue    workqueue.Interface
	groups   map[interface{}]int
	groupsMu *sync.RWMutex
}

func (w *worker) Run() {
	for {
		item, shutdown := w.queue.Get()
		if shutdown {
			return
		}
		task := item.(*Task)

		executeTask(task)

		w.queue.Done(task)
		if task.GroupID != nil {
			w.groupsMu.Lock()
			w.groups[task.GroupID]--
			w.groupsMu.Unlock()
		}
	}
}

func executeTask(task *Task) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Task %s/%s crashed: %v", task.GroupID, task.TaskID, r)
			fmt.Println(string(debug.Stack()))
		}
	}()
	err := task.Apply()
	if err != nil {
		log.Errorf("Task %s/%s failed: %v", task.GroupID, task.TaskID, err)
	}
}
