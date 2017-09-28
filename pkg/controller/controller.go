package controller

import (
	"context"

	"time"

	"reflect"

	"io"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/sirupsen/logrus"
)

type Controller interface {
	Init() error
	HandleTick()
	HandleNotification(msg *fes.Notification)
}

const (
	TICK_SPEED = time.Duration(1) * time.Second
)

type MetaController struct {
	ctrls []Controller
}

func NewMetaController(ctrls ...Controller) *MetaController {
	return &MetaController{ctrls: ctrls}
}

func (mc *MetaController) Init() error {
	logrus.Info("Running MetaController init.")
	for _, ctrl := range mc.ctrls {
		err := ctrl.Init()
		if err != nil {
			return err
		}
		logrus.Infof("'%s' controller init done.", reflect.TypeOf(ctrl))
	}

	logrus.Info("Finished MetaController init.")
	return nil
}

func (mc *MetaController) Run(ctx context.Context) error {
	logrus.Debug("Running controller init...")
	err := mc.Init()
	if err != nil {
		return err
	}

	// Control lane
	ticker := time.NewTicker(TICK_SPEED)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			mc.HandleTick()
		}
	}
}

func (mc *MetaController) HandleTick() {
	for _, ctrl := range mc.ctrls {
		ctrl.HandleTick()
	}
}

func (mc *MetaController) HandleNotification(msg *fes.Notification) {
	if msg == nil {
		return
	}

	// Might need smarter event router if used
	for _, ctrl := range mc.ctrls {
		ctrl.HandleNotification(msg)
	}
}

func (mc *MetaController) Close() error {
	var err error
	for _, ctrl := range mc.ctrls {
		if closer, ok := ctrl.(io.Closer); ok {
			err = closer.Close()
		}
	}
	logrus.Info("Closed MetaController")
	return err
}

//
// Actions
//

type Action interface {
	Id() string
	Apply() error
}
