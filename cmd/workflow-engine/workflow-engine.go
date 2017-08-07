package main

import (
	"context"
	"os"

	"github.com/fission/fission-workflow/cmd/workflow-engine/app"
	"github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cliApp := cli.NewApp()

	cliApp.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   "eventstore-url",
			Usage:  "Url to the data store used by the event store.",
			Value:  nats.DefaultURL,
			EnvVar: "EVENT_STORE_URL",
		},
		cli.StringFlag{
			Name:   "eventstore-cluster",
			Usage:  "Cluster name used for eventstore (if needed)",
			Value:  "test-cluster",
			EnvVar: "EVENT_STORE_CLUSTER",
		},
	}

	cliApp.Action = func(c *cli.Context) error {
		options := &app.Options{
			EventStore: &app.EventStoreOptions{
				Url:     c.String("eventstore-url"),
				Type:    "nats-streaming",
				Cluster: c.String("eventstore-cluster"),
			},
		}

		return app.Run(ctx, options)
	}

	cliApp.Run(os.Args)
}
