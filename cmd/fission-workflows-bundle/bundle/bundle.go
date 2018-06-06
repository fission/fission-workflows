package bundle

import (
	"context"
	"net"
	"net/http"
	"os"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/controller/expr"
	wfictr "github.com/fission/fission-workflows/pkg/controller/invocation"
	wfctr "github.com/fission/fission-workflows/pkg/controller/workflow"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fes/backend/nats"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/fnenv/fission"
	"github.com/fission/fission-workflows/pkg/fnenv/native"
	"github.com/fission/fission-workflows/pkg/fnenv/native/builtin"
	"github.com/fission/fission-workflows/pkg/fnenv/workflows"
	"github.com/fission/fission-workflows/pkg/scheduler"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/fission/fission-workflows/pkg/util/labels"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/fission/fission-workflows/pkg/version"
	controllerc "github.com/fission/fission/controller/client"
	executor "github.com/fission/fission/executor/client"
	"github.com/gorilla/handlers"
	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
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
}

type FissionOptions struct {
	ExecutorAddress string
	ControllerAddr  string
	RouterAddr      string
}

// Run serves enabled components in a blocking way
func Run(ctx context.Context, opts *Options) error {
	log.WithField("version", version.VersionInfo()).Info("Starting bundle...")

	var es fes.Backend
	var esPub pubsub.Publisher

	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()

	// Event Stores
	if opts.Nats != nil {
		log.WithFields(log.Fields{
			"url":     "!redacted!",
			"cluster": opts.Nats.Cluster,
			"client":  opts.Nats.Client,
		}).Infof("Using event store: NATS")
		natsEs := setupNatsEventStoreClient(opts.Nats.URL, opts.Nats.Cluster, opts.Nats.Client)
		es = natsEs
		esPub = natsEs
	}

	// Caches
	wfiCache := getWorkflowInvocationCache(ctx, esPub)
	wfCache := getWorkflowCache(ctx, esPub)

	// Resolvers and runtimes
	invocationAPI := api.NewInvocationAPI(es)
	resolvers := map[string]fnenv.RuntimeResolver{}
	runtimes := map[string]fnenv.Runtime{}

	log.Infof("Using Task Runtime: Workflow")
	reflectiveRuntime := workflows.NewRuntime(invocationAPI, wfiCache())
	runtimes[workflows.Name] = reflectiveRuntime
	if opts.InternalRuntime {
		log.Infof("Using Task Runtime: Internal")
		runtimes["internal"] = setupInternalFunctionRuntime()
		resolvers["internal"] = setupInternalFunctionRuntime()
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

	// Controller
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

		log.Info("Running controllers.")
		runController(ctx, ctrls...)
	}

	// Http servers
	if opts.Fission != nil {
		proxySrv := &http.Server{Addr: fissionProxyAddress}
		defer proxySrv.Shutdown(ctx)
		runFissionEnvironmentProxy(proxySrv, es, wfiCache(), wfCache(), resolvers)
	}

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
		lis, err := net.Listen("tcp", gRPCAddress)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		defer lis.Close()
		log.Info("Serving gRPC services at: ", lis.Addr())
		go grpcServer.Serve(lis)
	}

	if opts.HTTPGateway {
		HTTPGatewaySrv := &http.Server{Addr: apiGatewayAddress}
		defer HTTPGatewaySrv.Shutdown(ctx)
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
		serveHTTPGateway(ctx, HTTPGatewaySrv, admin, wf, wfi)
	}

	log.Info("Bundle set up.")
	<-ctx.Done()
	log.Info("Shutting down...")
	// TODO properly shutdown components
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
		Selector: labels.Or(
			labels.In(fes.PubSubLabelAggregateType, "invocation"),
			labels.In("parent.type", "invocation")),
	})
	wi := func() fes.Aggregator {
		return aggregates.NewWorkflowInvocation("")
	}

	return fes.NewSubscribedCache(ctx, fes.NewMapCache(), wi, invokeSub)
}

func setupWorkflowCache(ctx context.Context, workflowEventPub pubsub.Publisher) *fes.SubscribedCache {
	wfSub := workflowEventPub.Subscribe(pubsub.SubscriptionOptions{
		Buffer:   10,
		Selector: labels.In(fes.PubSubLabelAggregateType, "workflow"),
	})
	wb := func() fes.Aggregator {
		return aggregates.NewWorkflow("")
	}
	return fes.NewSubscribedCache(ctx, fes.NewMapCache(), wb, wfSub)
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

func serveHTTPGateway(ctx context.Context, gwSrv *http.Server, adminAPIAddr string, workflowAPIAddr string, invocationAPIAddr string) {
	mux := grpcruntime.NewServeMux()
	grpcOpts := []grpc.DialOption{grpc.WithInsecure()}
	if adminAPIAddr != "" {
		err := apiserver.RegisterAdminAPIHandlerFromEndpoint(ctx, mux, adminAPIAddr, grpcOpts)
		if err != nil {
			panic(err)
		}
	}

	if workflowAPIAddr != "" {
		err := apiserver.RegisterWorkflowAPIHandlerFromEndpoint(ctx, mux, workflowAPIAddr, grpcOpts)
		if err != nil {
			panic(err)
		}
	}

	if invocationAPIAddr != "" {
		err := apiserver.RegisterWorkflowInvocationAPIHandlerFromEndpoint(ctx, mux, invocationAPIAddr, grpcOpts)
		if err != nil {
			panic(err)
		}
	}

	gwSrv.Handler = mux
	go func() {
		err := gwSrv.ListenAndServe()
		log.WithField("err", err).Info("HTTP Gateway exited")
	}()

	log.Info("Serving HTTP API gateway at: ", gwSrv.Addr)
}

func runFissionEnvironmentProxy(proxySrv *http.Server, es fes.Backend, wfiCache fes.CacheReader,
	wfCache fes.CacheReader, resolvers map[string]fnenv.RuntimeResolver) {

	workflowParser := fnenv.NewMetaResolver(resolvers)
	workflowAPI := api.NewWorkflowAPI(es, workflowParser)
	wfServer := apiserver.NewWorkflow(workflowAPI, wfCache)
	wfiAPI := api.NewInvocationAPI(es)
	wfiServer := apiserver.NewInvocation(wfiAPI, wfiCache)
	proxyMux := http.NewServeMux()
	fissionProxyServer := fission.NewFissionProxyServer(wfiServer, wfServer)
	fissionProxyServer.RegisterServer(proxyMux)

	proxySrv.Handler = handlers.LoggingHandler(os.Stdout, proxyMux)
	go proxySrv.ListenAndServe()
	log.Info("Serving HTTP Fission Proxy at: ", proxySrv.Addr)
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

func runController(ctx context.Context, ctrls ...controller.Controller) {
	ctrl := controller.NewMetaController(ctrls...)
	go ctrl.Run(ctx)
}
