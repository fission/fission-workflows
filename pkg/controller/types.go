package controller

import (
	"github.com/fission/fission-workflows/pkg/fes"
	"time"
	"github.com/sirupsen/logrus"
	"context"
)

const (
	NotificationBuffer = 100
	WorkQueueSize      = 50
	InvocationTimeout  = time.Duration(10) * time.Minute
	MaxErrorCount      = 3
	LogKeyController   = "ctrl"
)

var log = logrus.New().WithFields(logrus.Fields{
	"component":      "controller",
	LogKeyController: "?",
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
