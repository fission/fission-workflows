package integration

import (
	"context"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/fission/fission-workflows/cmd/fission-workflows-bundle/bundle"
	fesnats "github.com/fission/fission-workflows/pkg/fes/backend/nats"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/sirupsen/logrus"
)

// SetupBundle sets up and runs the workflows-bundle.
//
// By default the bundle runs with all components are enabled, setting up a NATS cluster as the
// backing event store, and internal fnenv and workflow runtime
func SetupBundle(ctx context.Context, opts ...bundle.Options) bundle.Options {
	nats := SetupNatsCluster(ctx)
	var bundleOpts bundle.Options
	if len(opts) > 0 {
		bundleOpts = opts[0]
	} else {
		bundleOpts = bundle.Options{
			InternalRuntime:       true,
			InvocationController:  true,
			WorkflowController:    true,
			ApiHttp:               true,
			ApiWorkflowInvocation: true,
			ApiWorkflow:           true,
			ApiAdmin:              true,
			Nats:                  &nats,
		}
	}
	go bundle.Run(ctx, &bundleOpts)
	return bundleOpts
}

// TODO check if there is a nats instance already is running
func SetupNatsCluster(ctx context.Context) fesnats.Config {
	id := util.Uid()
	clusterId := fmt.Sprintf("fission-workflows-tests-%s", id)
	port, err := findFreePort()
	if err != nil {
		panic(err)
	}
	address := "127.0.0.1"
	flags := strings.Split(fmt.Sprintf("-cid %s -p %d -a %s", clusterId, port, address), " ")
	cmd := exec.CommandContext(ctx, "nats-streaming-server", flags...)
	stdOut, _ := cmd.StdoutPipe()
	stdErr, _ := cmd.StderrPipe()
	go io.Copy(os.Stdout, stdOut)
	go io.Copy(os.Stdout, stdErr)
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	cfg := fesnats.Config{
		Cluster: clusterId,
		Client:  fmt.Sprintf("client-%s", id),
		Url:     fmt.Sprintf("nats://%s:%d", address, port),
	}

	logrus.WithField("config", cfg).Info("Setting up NATS server")

	// wait for a bit to set it up
	awaitCtx, cancel := context.WithTimeout(ctx, time.Duration(10)*time.Second)
	defer cancel()
	err = waitForNats(awaitCtx, cfg.Url, cfg.Cluster)
	if err != nil {
		logrus.Error(err)
	}
	logrus.WithField("config", cfg).Info("NATS Server running")

	return cfg
}

func findFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer listener.Close()
	tcpAddr := listener.Addr().(*net.TCPAddr)
	return tcpAddr.Port, nil
}

// Wait for NATS to come online, ignoring ErrNoServer as it could mean that NATS is still being setup
func waitForNats(ctx context.Context, url string, cluster string) error {
	conn, err := stan.Connect(cluster, "setupEventStore-alive-test", stan.NatsURL(url),
		stan.ConnectWait(time.Duration(10)*time.Second))
	if err == nats.ErrNoServers {
		logrus.WithFields(logrus.Fields{
			"cluster": cluster,
			"url":     url,
		}).Warnf("retrying due to err: %v", err)
		select {
		case <-time.After(time.Duration(1) * time.Second):
			return waitForNats(ctx, url, cluster)
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	if err != nil {
		return err
	}
	defer conn.Close()
	return nil
}
