package controller

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/sirupsen/logrus"
)

const (
	NotificationBuffer = 100
	WorkQueueSize      = 50
	InvocationTimeout  = time.Duration(10) * time.Minute
	MaxErrorCount      = 0
)

var log = logrus.New().WithFields(logrus.Fields{
	"component": "controller",
})

type Controller interface {
	Init(ctx context.Context) error
	HandleTick() error
	HandleNotification(msg *fes.Notification) error
}

type Action interface {
	Id() string
	Apply() error
}

// TODO maybe an explanation func or argument?
type Filter interface {
	Filter(wfi *types.WorkflowInstance, status *ControlState) (actions []Action, err error)
}

type ControlState struct {
	ErrorCount  uint32
	RecentError error
	QueueSize   uint32
	lock        *sync.Mutex
	// TODO record timestamp/version of used model
}

func NewControlState() *ControlState {
	return &ControlState{
		lock: &sync.Mutex{},
	}
}

func (cs *ControlState) Lock() {
	if cs.lock == nil {
		cs.lock = &sync.Mutex{}
	}
	cs.lock.Lock()
}

func (cs *ControlState) Unlock() {
	if cs.lock == nil {
		cs.lock = &sync.Mutex{}
	} else {
		cs.lock.Unlock()
	}
}

func (cs *ControlState) AddError(err error) uint32 {
	cs.RecentError = err
	return atomic.AddUint32(&cs.ErrorCount, 1)
}

func (cs *ControlState) ResetError() {
	cs.RecentError = nil
	cs.ErrorCount = 0
}

func (cs *ControlState) IncrementQueueSize() uint32 {
	return atomic.AddUint32(&cs.QueueSize, 1)
}

func (cs *ControlState) DecrementQueueSize() uint32 {
	return atomic.AddUint32(&cs.QueueSize, ^uint32(0)) // TODO avoid overflow
}
