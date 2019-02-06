package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fission/fission-workflows/cmd/fission-workflows-bundle/bundle"
	"github.com/fission/fission-workflows/pkg/fes/backend/nats"
	"github.com/fission/fission-workflows/pkg/util"
	natsio "github.com/nats-io/go-nats"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
)

func main() {
	ctx, cancelFn := context.WithCancel(context.Background())
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func() {
		for sig := range c {
			fmt.Println("Received signal: ", sig)
			go func() {
				time.Sleep(30 * time.Second)
				fmt.Println("Deadline exceeded; forcing shutdown.")
				os.Exit(0)
			}()
			cancelFn()
			break
		}
	}()

	cliApp := createCli()
	cliApp.Action = func(c *cli.Context) error {
		setupLogging(c)

		return bundle.Run(ctx, &bundle.Options{
			Nats:                 parseNatsOptions(c),
			Fission:              parseFissionOptions(c),
			InternalRuntime:      c.Bool("internal"),
			InvocationController: c.Bool("controller") || c.Bool("invocation-controller"),
			WorkflowController:   c.Bool("controller") || c.Bool("workflow-controller"),
			AdminAPI:             c.Bool("api") || c.Bool("api-admin"),
			WorkflowAPI:          c.Bool("api") || c.Bool("api-workflow"),
			InvocationAPI:        c.Bool("api") || c.Bool("api-workflow-invocation"),
			HTTPGateway:          c.Bool("api") || c.Bool("api-http"),
			Metrics:              c.Bool("metrics"),
			Debug:                c.Bool("debug"),
			FissionProxy:         c.Bool("fission-proxy"),
			CheckFnenvStatus:     c.Bool("test"),
		})
	}
	cliApp.Run(os.Args)
}

func setupLogging(c *cli.Context) {
	if c.Bool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	} else {
		logrus.SetLevel(logrus.InfoLevel)
	}
}

func parseFissionOptions(c *cli.Context) *bundle.FissionOptions {
	if !c.Bool("fission") {
		return nil
	}

	return &bundle.FissionOptions{
		ExecutorAddress: c.String("fission-executor"),
		ControllerAddr:  c.String("fission-controller"),
		RouterAddr:      c.String("fission-router"),
	}
}

func parseNatsOptions(c *cli.Context) *nats.Config {
	if !c.Bool("nats") {
		return nil
	}

	client := c.String("nats-client")
	if client == "" {
		client = fmt.Sprintf("workflow-bundle-%s", util.UID())
	}

	return &nats.Config{
		URL:     c.String("nats-url"),
		Cluster: c.String("nats-cluster"),
		Client:  client,
	}
}

func createCli() *cli.App {

	cliApp := cli.NewApp()

	cliApp.Flags = []cli.Flag{
		// Generic
		cli.BoolFlag{
			Name:   "d, debug",
			EnvVar: "WORKFLOW_DEBUG",
		},

		// NATS
		cli.StringFlag{
			Name:   "nats-url",
			Usage:  "URL to the data store used by the NATS event store.",
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
			Usage:  "Client name used for the NATS event store. By default it will generate a unique clientID.",
			EnvVar: "ES_NATS_CLIENT",
		},
		cli.BoolFlag{
			Name:  "nats",
			Usage: "Use NATS as the event store",
		},

		// Fission
		cli.BoolFlag{
			Name:  "fission-proxy",
			Usage: "Serve a Fission environment as a proxy",
		},
		cli.BoolFlag{
			Name:  "fission",
			Usage: "Use Fission as a function environment",
		},
		cli.StringFlag{
			Name:   "fission-executor",
			Usage:  "Address of the Fission executor to optimize executions",
			Value:  "http://executor.fission",
			EnvVar: "FNENV_FISSION_EXECUTOR",
		},
		cli.StringFlag{
			Name:   "fission-controller",
			Usage:  "Address of the Fission controller for resolving functions",
			Value:  "http://controller.fission",
			EnvVar: "FNENV_FISSION_CONTROLLER",
		},
		cli.StringFlag{
			Name:   "fission-router",
			Usage:  "Address of the Fission router for executing functions",
			Value:  "http://router.fission",
			EnvVar: "FNENV_FISSION_ROUTER",
		},

		// Components
		cli.BoolFlag{
			Name:  "internal",
			Usage: "Use internal function runtime",
		},
		cli.BoolFlag{
			Name:  "controller",
			Usage: "Run the controller with all components",
		},
		cli.BoolFlag{
			Name:  "workflow-controller",
			Usage: "Run the workflow controller",
		},
		cli.BoolFlag{
			Name:  "invocation-controller",
			Usage: "Run the invocation controller",
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
		cli.BoolFlag{
			Name:  "metrics",
			Usage: "Serve prometheus metrics",
		},
		cli.BoolFlag{
			Name:  "api",
			Usage: "Shortcut for serving all APIs over both gRPC and HTTP",
		},
		cli.BoolFlag{
			Name:  "test",
			Usage: "Ensure that the proxy can reach the function runtimes (Fission) before serving the workflow engine..",
		},
	}

	return cliApp
}
