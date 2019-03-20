package bundle

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/apiserver/fission"
	"github.com/gorilla/handlers"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

type FissionProxyConfig struct {
	DefaultTimeout time.Duration
	ProxyAddr      string
	WorkflowsAddr  string
	ExposeMetrics  bool

	server *http.Server
}

func ParseFissionProxyConfig(ctx *cli.Context) (*FissionProxyConfig, error) {
	if ctx.Bool("fission.proxy") {
		return nil, nil
	}
	return &FissionProxyConfig{
		WorkflowsAddr:  gRPCAddress,
		ProxyAddr:      ctx.String("fission.proxy.addr"),
		DefaultTimeout: ctx.Duration("fission.proxy.timeout"),
		ExposeMetrics:  ctx.Bool("metrics"),
	}, nil
}

func (c *FissionProxyConfig) Run() error {
	if c == nil {
		return nil
	}

	conn, err := grpc.Dial(c.ProxyAddr, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}

	proxyMux := http.NewServeMux()
	fissionProxyServer := fission.NewEnvironmentProxyServer(apiserver.NewClient(conn), c.DefaultTimeout)
	fissionProxyServer.RegisterServer(proxyMux)
	fissionProxySrv := &http.Server{
		Addr:    c.ProxyAddr,
		Handler: handlers.LoggingHandler(os.Stdout, proxyMux),
	}

	log.Infof("Serving HTTP Fission Proxy at: %s", fissionProxySrv.Addr)
	err = fissionProxySrv.ListenAndServe()
	log.Infof("Fission Proxy server stopped: %v", err)
	return err
}

func (c *FissionProxyConfig) Close() error {
	if c == nil {
		return nil
	}

	err := c.server.Shutdown(context.Background())
	if err != nil {
		log.Errorf("Failed to stop Fission Proxy server: %v", err)
		return err
	}
	log.Info("Stopped Fission Proxy server")
	return nil
}
