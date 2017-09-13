package main

import (
	"os"

	"context"

	"github.com/fission/fission-workflow/cmd/fission-workflow-bundle/bundle"
	natsio "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	cliApp := createCli()
	cliApp.Action = func(c *cli.Context) error {

		return bundle.Run(ctx, &bundle.Options{
			Nats:                  parseNatsOptions(c),
			Fission:               parseFissionOptions(c),
			InternalRuntime:       c.Bool("internal"),
			Controller:            c.Bool("controller"),
			ApiAdmin:              c.Bool("api-admin"),
			ApiWorkflow:           c.Bool("api-workflow"),
			ApiWorkflowInvocation: c.Bool("api-workflow-invocation"),
			ApiHttp:               c.Bool("api-http"),
		})
	}
	cliApp.Run(os.Args)
}

func parseFissionOptions(c *cli.Context) *bundle.FissionOptions {
	if !c.Bool("fission") {
		return nil
	}

	return &bundle.FissionOptions{
		PoolmgrAddr:    c.String("fission-poolmgr"),
		ControllerAddr: c.String("fission-controller"),
	}
}

func parseNatsOptions(c *cli.Context) *bundle.NatsOptions {
	if !c.Bool("nats") {
		return nil
	}

	return &bundle.NatsOptions{
		Url:     c.String("nats-url"),
		Cluster: c.String("nats-cluster"),
		Client:  c.String("nats-client"),
	}
}

func createCli() *cli.App {

	cliApp := cli.NewApp()

	cliApp.Flags = []cli.Flag{
		// NATS
		cli.StringFlag{
			Name:   "nats-url",
			Usage:  "Url to the data store used by the NATS event store.",
			Value:  natsio.DefaultURL, // http://nats-streaming.fission
			EnvVar: "ES_NATS_URL",
		},
		cli.StringFlag{
			Name:   "nats-cluster",
			Usage:  "Cluster name used for the NATS event store (if needed)",
			Value:  "test-cluster", // mqtrigger
			EnvVar: "ES_NATS_CLUSTER",
		},
		cli.StringFlag{
			Name:   "nats-client",
			Usage:  "Client name used for the NATS event store",
			Value:  "undefined-client",
			EnvVar: "ES_NATS_CLIENT",
		},
		cli.BoolFlag{
			Name:  "nats",
			Usage: "Use NATS as the event store",
		},

		// Fission
		cli.BoolFlag{
			Name:  "fission",
			Usage: "Use Fission as a function environment",
		},
		cli.StringFlag{
			Name:   "fission-poolmgr",
			Usage:  "Address of the poolmgr",
			Value:  "http://poolmgr.fission",
			EnvVar: "FNENV_FISSION_POOLMGR",
		},
		cli.StringFlag{
			Name:  "fission-controller",
			Usage: "Address of the controller for resolving functions",
			Value: "http://controller.fission",
			EnvVar: "FNENV_FISSION_CONTROLLER",
		},

		// Components
		cli.BoolFlag{
			Name: "internal",
			Usage: "Use internal function runtime",
		},
		cli.BoolFlag{
			Name:  "controller",
			Usage: "Run the controller",
		},
		cli.BoolFlag{
			Name:  "api-http",
			Usage: "Serve the http apis of the apis",
		},
		cli.BoolFlag{
			Name:  "api-workflow-invocation",
			Usage: "Serve the workflow invocation gRPC api",
		},
		cli.BoolFlag{
			Name:  "api-workflow",
			Usage: "Serve the workflow gRPC api",
		},
		cli.BoolFlag{
			Name:  "api-admin",
			Usage: "Serve the admin gRPC api",
		},
	}

	return cliApp
}
