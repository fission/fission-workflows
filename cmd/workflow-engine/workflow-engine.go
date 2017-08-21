package main

import (
	"context"
	"os"

	"github.com/fission/fission-workflow/cmd/workflow-engine/app"
	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/fnenv/native"
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

	// Fission client
	//poolmgrClient := poolmgr.MakeClient("http://192.168.99.100:32101")
	//controllerClient := client.MakeClient("http://192.168.99.100:31313")
	//fissionApi := fission.NewFunctionEnv(poolmgrClient, controllerClient)
	//fissionResolver := fission.NewResolver(controllerClient)
	nativeEnv := native.NewFunctionEnv()

	cliApp.Action = func(c *cli.Context) error {
		options := &app.Options{
			EventStore: &app.EventStoreOptions{
				Url:     c.String("eventstore-url"),
				Type:    "nats-streaming",
				Cluster: c.String("eventstore-cluster"),
			},
			GrpcApiServerAddress: app.GRPC_ADDRESS,
			HttpApiServerAddress: app.API_GATEWAY_ADDRESS,
			FissionProxyAddress:  app.FISSION_PROXY_ADDRESS,
			FunctionRuntimeEnv: map[string]function.Runtime{
				//"fission":  fissionApi,
				"internal": nativeEnv,
			},
			FunctionRegistry: map[string]function.Resolver{
				//"fission":  fissionResolver,
				"internal": nativeEnv,
			},
		}

		return app.Run(ctx, options)
	}

	cliApp.Run(os.Args)
}
