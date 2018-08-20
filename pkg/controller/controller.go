package controller

import (
	"context"
	"errors"
	"io"
	"reflect"
	"time"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	TickInterval = time.Duration(100) * time.Millisecond
)

var (
	metaLog = logrus.New().WithFields(logrus.Fields{"component": "controller"})

	// Controller-related metrics
	EvalJobs = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "workflows",
		Subsystem: "controller_workflow",
		Name:      "eval_job",
		Help:      "Count of the different statuses of evaluations (e.g. skipped, errored, action).",
	}, []string{"controller", "status"})

	EvalRecovered = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "workflows",
		Subsystem: "controller_workflow",
		Name:      "eval_recovered",
		Help:      "Count of the number of jobs that were lost and recovered",
	}, []string{"controller", "from"})

	EvalDuration = prometheus.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "workflows",
		Subsystem: "controller_workflow",
		Name:      "eval_duration",
		Help:      "Duration of an evaluation.",
	}, []string{"controller", "action"})

	EvalQueueSize = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "workflows",
		Subsystem: "controller_workflow",
		Name:      "eval_queue_size",
		Help:      "A gauge of the evaluation queue size",
	}, []string{"controller"})
)

func init() {
	prometheus.MustRegister(EvalJobs, EvalDuration, EvalQueueSize, EvalRecovered)
}

type Controller interface {
	Init(ctx context.Context) error
	Tick(tick uint64) error
	Notify(msg *fes.Notification) error
}

type Action interface {
	Apply() error
}

type Rule interface {
	Eval(cec EvalContext) Action
}

type EvalContext interface {
	EvalState() *EvalState
}

// MetaController is a 'controller for controllers', allowing for composition with controllers. It allows users to
// interface with the metacontroller, instead of needing to control the lifecycle of all underlying controllers.
type MetaController struct {
	// suspended indicates whether this metacontroller is suspended from running.
	suspended bool

	// ctrls contains the controllers that the metacontroller manages.
	ctrls []Controller
}

func NewMetaController(ctrls ...Controller) *MetaController {
	return &MetaController{ctrls: ctrls}
}

func (mc *MetaController) Init(ctx context.Context) error {
	for _, ctrl := range mc.ctrls {
		err := ctrl.Init(ctx)
		if err != nil {
			return err
		}
		metaLog.Debugf("Controller '%s' initialized.", reflect.TypeOf(ctrl))
	}

	return nil
}

// Run essentially is Init followed by Start
func (mc *MetaController) Run(ctx context.Context) error {
	err := mc.Init(ctx)
	if err != nil {
		return err
	}

	return mc.Start(ctx)
}

func (mc *MetaController) Start(ctx context.Context) error {
	// Control lane
	ticker := time.NewTicker(TickInterval)
	tick := uint64(0)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if !mc.suspended {
				mc.Tick(tick)
				tick++
			}
		}
	}
}

func (mc *MetaController) Tick(tick uint64) error {
	var err error
	for _, ctrl := range mc.ctrls {
		err = ctrl.Tick(tick)
	}
	return err
}

func (mc *MetaController) Notify(msg *fes.Notification) error {
	if msg == nil {
		return errors.New("cannot handle empty message")
	}

	// Future: Might need smarter event router, to avoid bothering controllers with notifications that don't concern them
	var err error
	for _, ctrl := range mc.ctrls {
		metaLog.WithField("msg", msg.EventType).Debugf("Routing msg to %v", ctrl)
		err = ctrl.Notify(msg)
	}
	return err
}

func (mc *MetaController) Close() error {
	var merr error
	for _, ctrl := range mc.ctrls {
		if closer, ok := ctrl.(io.Closer); ok {
			err := closer.Close()
			if err != nil {
				merr = err
				metaLog.Errorf("Failed to stop controller %s: %v", reflect.TypeOf(ctrl), err)
			} else {
				metaLog.Debugf("Controller '%s' stopped.", reflect.TypeOf(ctrl))
			}
		}
	}
	return merr
}
