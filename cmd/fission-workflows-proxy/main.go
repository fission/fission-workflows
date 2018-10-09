package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/apiserver/fission"
	"github.com/fission/fission-workflows/pkg/version"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/gorilla/handlers"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

func main() {
	app := cli.NewApp()
	app.Author = "Erwin van Eyk"
	app.Email = "erwin@platform9.com"
	app.Version = version.Version
	app.EnableBashCompletion = true
	app.Usage = "A Fission environment that functions as a proxy between Fission and Fission Workflows."
	app.Description = app.Usage
	app.HideVersion = true
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name: "version, v",
		},
		cli.IntFlag{
			Name:  "verbosity",
			Value: 1,
			Usage: "CLI verbosity (0 is quiet, 1 is the default, 2 is verbose.)",
		},
		cli.StringFlag{
			Name:  "target, t",
			Value: "fission-workflows",
			Usage: "Address of Fission Workflows to proxy traffic to (do not add a scheme)",
		},
		cli.IntFlag{
			Name:  "port, p",
			Value: 80,
			Usage: "Port to serve the proxy at.",
		},
		cli.BoolFlag{
			Name:  "test",
			Usage: "Ensure that the proxy can reach the workflows before serving the proxy.",
		},
	}
	app.Action = commandContext(func(cliCtx Context) error {
		// Print version if asked
		if cliCtx.IsSet("version") {
			fmt.Println(version.VersionInfo().JSON())
			return nil
		}

		// Ensure that context is canceled upon termination signals
		ctx, cancelFn := context.WithCancel(cliCtx)
		c := make(chan os.Signal, 1)
		signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGTERM)
		go func() {
			for sig := range c {
				logrus.Debugf("Received signal: %v", sig)
				go func() {
					time.Sleep(30 * time.Second)
					fmt.Println("Deadline exceeded; forcing shutdown.")
					os.Exit(1)
				}()
				cancelFn()
				break
			}
		}()

		// Establish connection
		target := cliCtx.String("target")
		conn, err := grpc.DialContext(ctx, target, grpc.WithInsecure())
		if err != nil {
			logrus.Fatalf("Failed to establish connection to '%s': %v", target, err)
		}
		logrus.Infof("Established gRPC connection to '%s'", target)

		// Setup proxy
		// TODO we could also reuse the bundle here
		workflowClient := apiserver.NewWorkflowAPIClient(conn)
		invocationClient := apiserver.NewWorkflowInvocationAPIClient(conn)
		proxy := fission.NewEnvironmentProxyServer(invocationClient, workflowClient)

		// Test proxy
		if cliCtx.Bool("test") {
			adminClient := apiserver.NewAdminAPIClient(conn)
			statusCtx, cancelFn := context.WithTimeout(ctx, 10*time.Second)
			_, err := adminClient.Status(statusCtx, &empty.Empty{})
			if err != nil {
				logrus.Fatalf("Failed to reach workflows deployment: %v", err)
			}
			cancelFn()
		}

		// Setup proxy server
		proxyAddr := fmt.Sprintf(":%d", cliCtx.Int("port"))
		proxySrv := &http.Server{Addr: proxyAddr}
		mux := http.NewServeMux()
		proxy.RegisterServer(mux)
		proxySrv.Handler = handlers.LoggingHandler(os.Stdout, mux)

		// Serve...
		go func() {
			<-ctx.Done()
			logrus.Debugf("Shutting down server...")
			proxySrv.Shutdown(ctx)
		}()
		logrus.Infof("Serving proxy at: %s", proxyAddr)
		if err := proxySrv.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				logrus.Info(err)
			} else {
				logrus.Error(err)
			}
		}

		return nil
	})
	app.Run(os.Args)
}

func commandContext(fn func(c Context) error) func(c *cli.Context) error {
	return func(c *cli.Context) error {
		switch c.GlobalInt("verbosity") {
		case 0:
			logrus.SetLevel(logrus.ErrorLevel)
		case 1:
			logrus.SetLevel(logrus.InfoLevel)
		default:
			fallthrough
		case 2:
			logrus.SetLevel(logrus.DebugLevel)
		}
		return fn(Context{c})
	}
}

type Context struct {
	*cli.Context
}

func (c Context) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c Context) Done() <-chan struct{} {
	return nil
}

func (c Context) Err() error {
	return nil
}

func (c Context) Value(key interface{}) interface{} {
	if s, ok := key.(string); ok {
		return c.Generic(s)
	}
	return nil
}
