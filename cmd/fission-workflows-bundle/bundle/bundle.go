package bundle

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/api/aggregates"
	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/controller/expr"
	wfictr "github.com/fission/fission-workflows/pkg/controller/invocation"
	wfctr "github.com/fission/fission-workflows/pkg/controller/workflow"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fes/backend/mem"
	"github.com/fission/fission-workflows/pkg/fes/backend/nats"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/fnenv/fission"
	"github.com/fission/fission-workflows/pkg/fnenv/native"
	"github.com/fission/fission-workflows/pkg/fnenv/native/builtin"
	"github.com/fission/fission-workflows/pkg/fnenv/workflows"
	"github.com/fission/fission-workflows/pkg/scheduler"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/fission/fission-workflows/pkg/util/labels"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/fission/fission-workflows/pkg/version"
	controllerc "github.com/fission/fission/controller/client"
	executor "github.com/fission/fission/executor/client"
	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	gRPCAddress         = ":5555"
	apiGatewayAddress   = ":8080"
	fissionProxyAddress = ":8888"
)

type Options struct {
	Nats                 *nats.Config
	Fission              *FissionOptions
	InternalRuntime      bool
	InvocationController bool
	WorkflowController   bool
	AdminAPI             bool
	WorkflowAPI          bool
	HTTPGateway          bool
	InvocationAPI        bool
	Metrics              bool
}

type FissionOptions struct {
	ExecutorAddress string
	ControllerAddr  string
	RouterAddr      string
}

// Run serves enabled components in a blocking way
func Run(ctx context.Context, opts *Options) error {
	log.WithFields(log.Fields{
		"version": fmt.Sprintf("%+v", version.VersionInfo()),
		"config":  fmt.Sprintf("%+v", opts),
	}).Info("Starting bundle...")

	var es fes.Backend
	var esPub pubsub.Publisher

	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)

	//
	// Event Store
	//
	if opts.Nats != nil {
		log.WithFields(log.Fields{
			"url":     "!redacted!",
			"cluster": opts.Nats.Cluster,
			"client":  opts.Nats.Client,
		}).Infof("Using event store: NATS")
		natsEs := setupNatsEventStoreClient(opts.Nats.URL, opts.Nats.Cluster, opts.Nats.Client)
		es = natsEs
		esPub = natsEs
	} else {
		log.Warn("No event store provided; using the development, in-memory event store")
		backend := mem.NewBackend()
		es = backend
		esPub = backend
	}

	// Caches
	wfiCache := getWorkflowInvocationCache(ctx, esPub)
	wfCache := getWorkflowCache(ctx, esPub)

	//
	// Function Runtimes
	//
	invocationAPI := api.NewInvocationAPI(es)
	resolvers := map[string]fnenv.RuntimeResolver{}
	runtimes := map[string]fnenv.Runtime{}

	if opts.InternalRuntime || opts.Fission != nil {
		log.Infof("Using Task Runtime: Workflow")
		reflectiveRuntime := workflows.NewRuntime(invocationAPI, wfiCache())
		runtimes[workflows.Name] = reflectiveRuntime
	} else {
		log.Info("No function runtimes specified.")
	}
	if opts.InternalRuntime {
		log.Infof("Using Task Runtime: Internal")
		internalRuntime := setupInternalFunctionRuntime()
		runtimes["internal"] = internalRuntime
		resolvers["internal"] = internalRuntime
		log.Infof("Internal runtime functions: %v", internalRuntime.Installed())
	}
	if opts.Fission != nil {
		log.WithFields(log.Fields{
			"controller": opts.Fission.ControllerAddr,
			"router":     opts.Fission.RouterAddr,
			"executor":   opts.Fission.ExecutorAddress,
		}).Infof("Using Task Runtime: Fission")
		runtimes["fission"] = setupFissionFunctionRuntime(opts.Fission.ExecutorAddress, opts.Fission.RouterAddr)
		resolvers["fission"] = setupFissionFunctionResolver(opts.Fission.ControllerAddr)
	}

	//
	// Controllers
	//
	if opts.InvocationController || opts.WorkflowController {
		var ctrls []controller.Controller
		if opts.WorkflowController {
			log.Info("Using controller: workflow")
			ctrls = append(ctrls, setupWorkflowController(wfCache(), es, resolvers))
		}

		if opts.InvocationController {
			log.Info("Using controller: invocation")
			ctrls = append(ctrls, setupInvocationController(wfiCache(), wfCache(), es, runtimes, resolvers))
		}

		ctrl := controller.NewMetaController(ctrls...)
		go ctrl.Run(ctx)
		defer func() {
			err := ctrl.Close()
			if err != nil {
				log.Errorf("Failed to stop controllers: %v", err)
			} else {
				log.Info("Stopped controllers")
			}
		}()
	} else {
		log.Info("No controllers specified to run.")
	}

	//
	// Fission integration
	//
	if opts.Fission != nil {
		proxyMux := http.NewServeMux()
		runFissionEnvironmentProxy(proxyMux, es, wfiCache(), wfCache(), resolvers)
		fissionProxySrv := &http.Server{Addr: fissionProxyAddress}
		fissionProxySrv.Handler = handlers.LoggingHandler(os.Stdout, proxyMux)

		if opts.Metrics {
			setupMetricsEndpoint(proxyMux)
		}

		go func() {
			err := fissionProxySrv.ListenAndServe()
			log.WithField("err", err).Info("Fission Proxy server stopped")
		}()
		defer func() {
			err := fissionProxySrv.Shutdown(ctx)
			if err != nil {
				log.Errorf("Failed to stop Fission Proxy server: %v", err)
			} else {
				log.Info("Stopped Fission Proxy server")
			}
		}()
		log.Info("Serving HTTP Fission Proxy at: ", fissionProxySrv.Addr)
	}

	//
	// gRPC API
	//
	if opts.AdminAPI {
		serveAdminAPI(grpcServer)
	}

	if opts.WorkflowAPI {
		serveWorkflowAPI(grpcServer, es, resolvers, wfCache())
	}

	if opts.InvocationAPI {
		serveInvocationAPI(grpcServer, es, wfiCache())
	}

	if opts.AdminAPI || opts.WorkflowAPI || opts.InvocationAPI {
		if opts.Metrics {
			log.Debug("Instrumenting gRPC server with Prometheus metrics")
			grpc_prometheus.Register(grpcServer)
		}

		lis, err := net.Listen("tcp", gRPCAddress)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		go grpcServer.Serve(lis)
		defer func() {
			grpcServer.GracefulStop()
			lis.Close()
			log.Info("Stopped gRPC server")
		}()
		log.Info("Serving gRPC services at: ", lis.Addr())
	}

	//
	// HTTP API
	//
	if opts.HTTPGateway || opts.Metrics {
		grpcMux := grpcruntime.NewServeMux()
		httpMux := http.NewServeMux()

		if opts.HTTPGateway {

			var admin, wf, wfi string
			if opts.AdminAPI {
				admin = gRPCAddress
			}
			if opts.WorkflowAPI {
				wf = gRPCAddress
			}
			if opts.InvocationAPI {
				wfi = gRPCAddress
			}
			serveHTTPGateway(ctx, grpcMux, admin, wf, wfi)
		}

		if opts.Metrics {
			setupMetricsEndpoint(httpMux)
			log.Infof("Set up prometheus collector: %v/metrics", apiGatewayAddress)
		}

		httpApiSrv := &http.Server{Addr: apiGatewayAddress}
		httpMux.Handle("/", grpcMux)
		httpApiSrv.Handler = handlers.LoggingHandler(os.Stdout, httpMux)
		go func() {
			err := httpApiSrv.ListenAndServe()
			log.WithField("err", err).Info("HTTP Gateway stopped")
		}()
		defer func() {
			err := httpApiSrv.Shutdown(ctx)
			if err != nil {
				log.Errorf("Failed to stop HTTP API server: %v", err)
			} else {
				log.Info("Stopped HTTP API server")
			}
		}()

		log.Info("Serving HTTP API gateway at: ", httpApiSrv.Addr)
	}

	log.Info("Setup completed.")
	<-ctx.Done()
	log.WithField("reason", ctx.Err()).Info("Shutting down...")
	time.Sleep(5 * time.Second)
	return nil
}

func getWorkflowCache(ctx context.Context, eventPub pubsub.Publisher) func() fes.CacheReaderWriter {
	var wfCache fes.CacheReaderWriter
	return func() fes.CacheReaderWriter {
		if wfCache != nil {
			return wfCache
		}

		wfCache = setupWorkflowCache(ctx, eventPub)
		return wfCache
	}
}

func getWorkflowInvocationCache(ctx context.Context, eventPub pubsub.Publisher) func() fes.CacheReaderWriter {
	var wfiCache fes.CacheReaderWriter
	return func() fes.CacheReaderWriter {
		if wfiCache != nil {
			return wfiCache
		}

		wfiCache = setupWorkflowInvocationCache(ctx, eventPub)
		return wfiCache
	}
}

func setupInternalFunctionRuntime() *native.FunctionEnv {
	return native.NewFunctionEnv(builtin.DefaultBuiltinFunctions)
}

func setupFissionFunctionRuntime(executorAddr string, routerAddr string) *fission.FunctionEnv {
	client := executor.MakeClient(executorAddr)
	return fission.NewFunctionEnv(client, routerAddr)
}

func setupFissionFunctionResolver(controllerAddr string) *fission.Resolver {
	controllerClient := controllerc.MakeClient(controllerAddr)
	return fission.NewResolver(controllerClient)
}

func setupNatsEventStoreClient(url string, cluster string, clientID string) *nats.EventStore {
	if clientID == "" {
		clientID = util.UID()
	}

	es, err := nats.Connect(nats.Config{
		Cluster: cluster,
		URL:     url,
		Client:  clientID,
	})
	if err != nil {
		panic(err)
	}

	err = es.Watch(fes.Aggregate{Type: "invocation"})
	if err != nil {
		panic(err)
	}
	err = es.Watch(fes.Aggregate{Type: "workflow"})
	if err != nil {
		panic(err)
	}
	return es
}

func setupWorkflowInvocationCache(ctx context.Context, invocationEventPub pubsub.Publisher) *fes.SubscribedCache {
	invokeSub := invocationEventPub.Subscribe(pubsub.SubscriptionOptions{
		Buffer: 50,
		LabelMatcher: labels.Or(
			labels.In(fes.PubSubLabelAggregateType, "invocation"),
			labels.In("parent.type", "invocation")),
	})
	wi := func() fes.Entity {
		return aggregates.NewWorkflowInvocation("")
	}

	return fes.NewSubscribedCache(ctx, fes.NewNamedMapCache("invocation"), wi, invokeSub)
}

func setupWorkflowCache(ctx context.Context, workflowEventPub pubsub.Publisher) *fes.SubscribedCache {
	wfSub := workflowEventPub.Subscribe(pubsub.SubscriptionOptions{
		Buffer:       10,
		LabelMatcher: labels.In(fes.PubSubLabelAggregateType, "workflow"),
	})
	wb := func() fes.Entity {
		return aggregates.NewWorkflow("")
	}
	return fes.NewSubscribedCache(ctx, fes.NewNamedMapCache("workflow"), wb, wfSub)
}

func serveAdminAPI(s *grpc.Server) {
	adminServer := &apiserver.Admin{}
	apiserver.RegisterAdminAPIServer(s, adminServer)
	log.Infof("Serving admin gRPC API at %s.", gRPCAddress)
}

func serveWorkflowAPI(s *grpc.Server, es fes.Backend, resolvers map[string]fnenv.RuntimeResolver,
	wfCache fes.CacheReader) {
	workflowParser := fnenv.NewMetaResolver(resolvers)
	workflowAPI := api.NewWorkflowAPI(es, workflowParser)
	workflowServer := apiserver.NewWorkflow(workflowAPI, wfCache)
	apiserver.RegisterWorkflowAPIServer(s, workflowServer)
	log.Infof("Serving workflow gRPC API at %s.", gRPCAddress)
}

func serveInvocationAPI(s *grpc.Server, es fes.Backend, wfiCache fes.CacheReader) {
	invocationAPI := api.NewInvocationAPI(es)
	invocationServer := apiserver.NewInvocation(invocationAPI, wfiCache)
	apiserver.RegisterWorkflowInvocationAPIServer(s, invocationServer)
	log.Infof("Serving workflow invocation gRPC API at %s.", gRPCAddress)
}

func serveHTTPGateway(ctx context.Context, mux *grpcruntime.ServeMux, adminAPIAddr string, workflowAPIAddr string,
	invocationAPIAddr string) {
	opts := []grpc.DialOption{grpc.WithInsecure()}
	if adminAPIAddr != "" {
		err := apiserver.RegisterAdminAPIHandlerFromEndpoint(ctx, mux, adminAPIAddr, opts)
		if err != nil {
			panic(err)
		}
		log.Info("Registered Workflow API HTTP Endpoint")
	}

	if workflowAPIAddr != "" {
		err := apiserver.RegisterWorkflowAPIHandlerFromEndpoint(ctx, mux, workflowAPIAddr, opts)
		if err != nil {
			panic(err)
		}
		log.Info("Registered Admin API HTTP Endpoint")
	}

	if invocationAPIAddr != "" {
		err := apiserver.RegisterWorkflowInvocationAPIHandlerFromEndpoint(ctx, mux, invocationAPIAddr, opts)
		if err != nil {
			panic(err)
		}
		log.Info("Registered Workflow Invocation API HTTP Endpoint")
	}
}

func runFissionEnvironmentProxy(proxyMux *http.ServeMux, es fes.Backend, wfiCache fes.CacheReader,
	wfCache fes.CacheReader, resolvers map[string]fnenv.RuntimeResolver) {

	workflowParser := fnenv.NewMetaResolver(resolvers)
	workflowAPI := api.NewWorkflowAPI(es, workflowParser)
	wfServer := apiserver.NewWorkflow(workflowAPI, wfCache)
	wfiAPI := api.NewInvocationAPI(es)
	wfiServer := apiserver.NewInvocation(wfiAPI, wfiCache)
	fissionProxyServer := fission.NewFissionProxyServer(wfiServer, wfServer)
	fissionProxyServer.RegisterServer(proxyMux)
}

func setupInvocationController(invocationCache fes.CacheReader, wfCache fes.CacheReader, es fes.Backend,
	fnRuntimes map[string]fnenv.Runtime, fnResolvers map[string]fnenv.RuntimeResolver) *wfictr.Controller {
	workflowAPI := api.NewWorkflowAPI(es, fnenv.NewMetaResolver(fnResolvers))
	invocationAPI := api.NewInvocationAPI(es)
	dynamicAPI := api.NewDynamicApi(workflowAPI, invocationAPI)
	taskAPI := api.NewTaskAPI(fnRuntimes, es, dynamicAPI)
	s := &scheduler.WorkflowScheduler{}
	stateStore := expr.NewStore()
	return wfictr.NewController(invocationCache, wfCache, s, taskAPI, invocationAPI, stateStore)
}

func setupWorkflowController(wfCache fes.CacheReader, es fes.Backend,
	fnResolvers map[string]fnenv.RuntimeResolver) *wfctr.Controller {
	workflowAPI := api.NewWorkflowAPI(es, fnenv.NewMetaResolver(fnResolvers))
	return wfctr.NewController(wfCache, workflowAPI)
}

func setupMetricsEndpoint(apiMux *http.ServeMux) {
	apiMux.Handle("/metrics", promhttp.Handler())
}
