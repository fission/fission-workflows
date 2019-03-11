package controller

import (
	"sync"

	"github.com/fission/fission-workflows/pkg/util/workqueue"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

var (
	tasksProcessed = prometheus.NewCounter(prometheus.CounterOpts{
		Namespace: "workflows",
		Subsystem: "executor",
		Name:      "tasks_processed",
	})

	tasksActive = prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: "workflows",
		Subsystem: "executor",
		Name:      "tasks_active",
	})
)

type LocalExecutor struct {
	//
	// Config
	//
	maxParallelism int

	//
	// State
	//
	queue    workqueue.Interface
	workers  []*worker
	groups   map[interface{}]int
	groupsMu *sync.RWMutex
}

type DefaultTask struct {
	// TaskID is used to ensure that there is only one instance of this task
	TaskID interface{}

	// GroupID is used to group together tasks.
	GroupID interface{}

	Apply func() error
}

func (t *DefaultTask) ID() interface{} {
	return t.TaskID
}

func NewLocalExecutor(maxParallelism, maxQueueSize int) *LocalExecutor {
	if maxParallelism <= 0 {
		panic("LocalExecutor: parallelism should be larger than 0")
	}
	if maxQueueSize <= 0 {
		panic("LocalExecutor: queue size should be larger than 0")
	}
	wq := workqueue.New()
	wq.MaxSize = maxQueueSize
	return &LocalExecutor{
		maxParallelism: maxParallelism,
		queue:          wq,
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

func (ex *LocalExecutor) Submit(t *DefaultTask) bool {
	// Add to the queue
	accepted := ex.queue.Add(t)
	if !accepted {
		return false
	}

	// Increment the group
	if t.GroupID != nil {
		ex.groupsMu.Lock()
		ex.groups[t.GroupID]++
		ex.groupsMu.Unlock()
	}
	return true
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
		task := item.(*DefaultTask)
		tasksActive.Inc()

		err := task.Apply()
		if err != nil {
			logrus.Errorf("Task %T failed: %v", task, err)
		}
		tasksActive.Dec()
		tasksProcessed.Inc()
		w.queue.Done(task)

		if task.GroupID != nil {
			w.groupsMu.Lock()
			w.groups[task.GroupID]--
			w.groupsMu.Unlock()
		}
	}
}
