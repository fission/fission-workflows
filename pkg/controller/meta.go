package controller

import (
	"context"

	"time"

	"reflect"

	"io"

	"errors"
	"github.com/fission/fission-workflows/pkg/fes"
)

const (
	TickInterval = time.Duration(1) * time.Second
)

var (
	metaLog = log.WithField("controller", "controller-meta")
)

// MetaController is a 'controller for controllers', allowing for composition with controllers. It allows users to
// interface with the metacontroller, instead of needing to control the lifecycle of all underlying controllers.
type MetaController struct {
	ctrls []Controller
}

func NewMetaController(ctrls ...Controller) *MetaController {
	return &MetaController{ctrls: ctrls}
}

func (mc *MetaController) Init(ctx context.Context) error {
	metaLog.Info("Running MetaController init.")
	for _, ctrl := range mc.ctrls {
		err := ctrl.Init(ctx)
		if err != nil {
			return err
		}
		metaLog.Infof("'%s' controller init done.", reflect.TypeOf(ctrl))
	}

	metaLog.Info("Finished MetaController init.")
	return nil
}

func (mc *MetaController) Run(ctx context.Context) error {
	metaLog.Debug("Running controller init...")
	err := mc.Init(ctx)
	if err != nil {
		return err
	}

	// Control lane
	ticker := time.NewTicker(TickInterval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			mc.HandleTick()
		}
	}
}

func (mc *MetaController) HandleTick() error {
	var err error
	for _, ctrl := range mc.ctrls {
		err = ctrl.HandleTick()
	}
	return err
}

func (mc *MetaController) HandleNotification(msg *fes.Notification) error {
	if msg == nil {
		return errors.New("cannot handle empty message")
	}

	// Future: Might need smarter event router, to avoid bothering controllers with notifications that don't concern them
	var err error
	for _, ctrl := range mc.ctrls {
		metaLog.WithField("msg", msg.EventType).Debugf("Routing msg to %v", ctrl)
		err = ctrl.HandleNotification(msg)
	}
	return err
}

func (mc *MetaController) Close() error {
	metaLog.Info("Closing metacontroller and its controllers...")
	var err error
	for _, ctrl := range mc.ctrls {
		if closer, ok := ctrl.(io.Closer); ok {
			err = closer.Close()
		}
	}
	metaLog.Info("Closed MetaController")
	return err
}
