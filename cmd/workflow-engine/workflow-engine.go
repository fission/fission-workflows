package main

import (
	"context"
	"github.com/fission/fission-workflow/cmd/workflow-engine/app"
	"github.com/sirupsen/logrus"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	app.Run(ctx)
}
