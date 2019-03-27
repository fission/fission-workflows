package ctrl

import (
	"context"
	"io"
	"runtime/debug"
	"sync"
	"time"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/util/workqueue"
	log "github.com/sirupsen/logrus"
)

// Future: decouple from fes.
type Event = fes.Notification

type Controller interface {
	Eval(ctx context.Context, processValue *Event) Result
}

type Identifier interface {
	ID() string
}

type Sensor interface {
	io.Closer
	Start(evalQueue EvalQueue) error
}

type EvalQueue interface {
	Submit(event *Event) bool
}

type Result interface {
	Apply(s *System, event *Event)
}

type ControllerFactory func(event *Event) (ctrl Controller, err error)

// Err logs the controller error.
type Err struct {
	Err error
}

func (r Err) Error() string {
	return r.Err.Error()
}

func (r Err) Apply(s *System, event *Event) {
	s.LoggerFor(event.Aggregate.Id).Errorf(r.Error())
}

// Success is default evaluation result.
type Success struct {
	Msg string
}

func (r Success) Apply(s *System, event *Event) {
	if len(r.Msg) > 0 {
		s.LoggerFor(event.Aggregate.Id).Debugf("controller: %v", r.Msg)
	}
}

// Done removes the controller for this evaluation, which prevents any further evaluations
type Done struct {
	Msg string
}

func (r Done) Apply(s *System, event *Event) {
	if len(r.Msg) == 0 {
		s.LoggerFor(event.Aggregate.Id).Debug("Removing finished controller")
	} else {
		s.LoggerFor(event.Aggregate.Id).Debugf("Removing finished controller: %v", r.Msg)
	}
	s.DeleteController(event.Aggregate.Id)
}

type ControllerStats struct {
	LastEvaluatedAt time.Time
	EvalCount       int64
}

func (c ControllerStats) RecordEval() ControllerStats {
	c.LastEvaluatedAt = time.Now()
	c.EvalCount++
	return c
}

// Future: support parallel executions in evaluator
type System struct {
	ctrls       map[string]Controller
	ctrlsMu     *sync.RWMutex
	ctrlStats   map[string]ControllerStats
	ctrlStatsMu *sync.RWMutex
	factory     ControllerFactory
	evalQueue   workqueue.Interface
	close       func()
	runOnce     *sync.Once
	logger      *log.Logger
}

func NewSystem(factory ControllerFactory) *System {
	return &System{
		factory:     factory,
		ctrlsMu:     &sync.RWMutex{},
		ctrls:       make(map[string]Controller),
		evalQueue:   workqueue.NewWorkQueue(workqueue.DefaultMaxSize, true),
		runOnce:     &sync.Once{},
		logger:      log.StandardLogger(),
		ctrlStats:   make(map[string]ControllerStats),
		ctrlStatsMu: &sync.RWMutex{},
	}
}

func (s *System) DeleteController(key string) {
	s.ctrlsMu.Lock()
	delete(s.ctrls, key)
	s.ctrlsMu.Unlock()
}

func (s *System) AddController(key string, ctrl Controller) {
	s.ctrlsMu.Lock()
	s.ctrls[key] = ctrl
	s.ctrlsMu.Unlock()
}

func (s *System) GetController(key string) (ctrl Controller, ok bool) {
	s.ctrlsMu.RLock()
	ctrl, ok = s.ctrls[key]
	s.ctrlsMu.RUnlock()
	return ctrl, ok
}

func (s *System) RangeControllerStats(consumer func(k string, v ControllerStats) bool) {
	s.ctrlStatsMu.RLock()
	defer s.ctrlStatsMu.RUnlock()
	for k, v := range s.ctrlStats {
		if !consumer(k, v) {
			break
		}
	}
}

func (s *System) Logger() *log.Logger {
	return s.logger
}

func (s *System) LoggerFor(entityID string) *log.Entry {
	return s.logger.WithField("key", entityID)
}

func (s *System) Submit(event *Event) bool {
	return s.evalQueue.Add(event)
}

func (s *System) Run() {
	s.runOnce.Do(func() {
		go s.run()
	})
}

func (s *System) run() {
	ctx, cancel := context.WithCancel(context.Background())
	s.close = cancel
	for {
		item, shutdown := s.evalQueue.Get()
		if shutdown {
			return
		}

		event, ok := item.(*Event)
		if !ok {
			s.logger.Errorf("Ignoring workqueue item. Expected an Event but got a %T", item)
			s.evalQueue.Done(item)
			continue
		}
		ctrlKey := event.Aggregate.Id
		s.LoggerFor(ctrlKey).Debugf("starting evaluation (reason: %v)", event.Event.GetType())

		// Get or create controller for item
		ctrl, ok := s.GetController(ctrlKey)
		if !ok {
			var err error
			ctrl, err = s.factory(event)
			if err != nil {
				s.LoggerFor(ctrlKey).Error(err)
				s.evalQueue.Done(item)
				continue
			}
			s.LoggerFor(ctrlKey).Debug("created new controller")
			s.AddController(ctrlKey, ctrl)
		}

		s.eval(ctx, ctrlKey, ctrl, event)
		s.evalQueue.Done(item)
	}
}

func (s *System) eval(ctx context.Context, ctrlKey string, ctrl Controller, event *Event) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Errorf("Recovered from controller crash: %v", r)
			if log.IsLevelEnabled(log.DebugLevel) {
				debug.PrintStack()
			}
		}
	}()

	// Record the evaluation
	s.ctrlStatsMu.Lock()
	s.ctrlStats[ctrlKey] = s.ctrlStats[ctrlKey].RecordEval()
	s.ctrlStatsMu.Unlock()

	// Trigger the evaluation
	result := ctrl.Eval(ctx, event)
	result.Apply(s, event)
}

func (s *System) Close() error {
	s.evalQueue.ShutDown()
	if s.close != nil {
		s.close()
	}
	return nil
}

type PollSensor struct {
	interval time.Duration
	poll     func(evalQueue EvalQueue)

	done   func()
	closeC <-chan struct{}
}

func NewPollSensor(interval time.Duration, pollFn func(queue EvalQueue)) *PollSensor {
	ctx, done := context.WithCancel(context.Background())
	return &PollSensor{
		interval: interval,
		done:     done,
		closeC:   ctx.Done(),
		poll:     pollFn,
	}
}

func (s *PollSensor) Close() error {
	s.done()
	return nil
}

func (s *PollSensor) Start(evalQueue EvalQueue) error {
	go s.Run(evalQueue)
	return nil
}

func (s *PollSensor) Run(evalQueue EvalQueue) {
	ticker := time.NewTicker(s.interval)
	for {
		select {
		case <-s.closeC:
			ticker.Stop()
			return
		case <-ticker.C:
		}

		s.poll(evalQueue)
	}
}
