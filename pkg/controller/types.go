package controller

import (
	"context"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	NotificationBuffer = 100
	WorkQueueSize      = 50
	InvocationTimeout  = time.Duration(10) * time.Minute
	MaxErrorCount      = 3
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
