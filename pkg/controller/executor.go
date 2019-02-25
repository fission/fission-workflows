package controller

import (
	"errors"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/util/workqueue"
)

var (
	ErrTaskQueueOverflow = errors.New("task queue full")

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
	maxQueueSize   int

	//
	// State
	//
	queue     workqueue.Interface
	workers   []*worker
	activeSet *sync.Map
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
		maxQueueSize:   maxQueueSize,
		queue:          workqueue.New(),
		activeSet:      &sync.Map{},
	}
}

func (ex *LocalExecutor) Start() {
	if ex.maxParallelism <= 0 {
		panic("LocalExecutor: parallelism should be larger than 0")
	}
	// Add workers based on max parallelism
	for i := 0; i < ex.maxParallelism; i++ {
		worker := &worker{
			queue:     ex.queue,
			activeSet: ex.activeSet,
		}
		ex.workers = append(ex.workers, worker)
		go worker.Run()
	}
}

func (ex *LocalExecutor) Close() error {
	ex.queue.ShutDown()
	return nil
}

// TODO Check behaviour that we can update the task in the
func (ex *LocalExecutor) Accept(t Task) error {
	// ignore if it is in the active set
	if _, ok := ex.activeSet.Load(t); ok {
		return nil
	}

	// Check if we do not exceed the max task queue size
	if ex.queue.Len() >= ex.maxQueueSize {
		return ErrTaskQueueOverflow
	}

	// Add to the queue
	ex.queue.Add(t)
	return nil
}

type Executor interface {
	Accept(t Task) error
}

type Task interface {
	Apply() error // TODO change to Run at some point
	// TODO add priority
}

type worker struct {
	queue     workqueue.Interface
	activeSet *sync.Map
}

func (w *worker) Run() {
	for {
		item, shutdown := w.queue.Get()
		if shutdown {
			break
		}
		w.activeSet.Store(item, true)
		task := item.(Task)
		tasksActive.Inc()
		err := task.Apply()
		if err != nil {
			logrus.Errorf("Task %T failed: %v", task, err)
		}
		tasksActive.Dec()
		tasksProcessed.Inc()
		w.activeSet.Delete(item)
	}
}
