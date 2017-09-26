package bundle

import (
	"os"

	"context"

	"net/http"

	"net"

	"github.com/fission/fission-workflows/pkg/api/function"
	"github.com/fission/fission-workflows/pkg/api/invocation"
	"github.com/fission/fission-workflows/pkg/api/workflow"
	"github.com/fission/fission-workflows/pkg/api/workflow/parse"
	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/controller/expr"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fes/eventstore/nats"
	"github.com/fission/fission-workflows/pkg/fnenv/fission"
	"github.com/fission/fission-workflows/pkg/fnenv/native"
	"github.com/fission/fission-workflows/pkg/fnenv/native/builtin"
	"github.com/fission/fission-workflows/pkg/scheduler"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/fission/fission-workflows/pkg/util/labels"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/fission/fission-workflows/pkg/version"
	controllerc "github.com/fission/fission/controller/client"
	poolmgrc "github.com/fission/fission/poolmgr/client"
	"github.com/gorilla/handlers"
	grpcruntime "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/nats-io/go-nats-streaming"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	GRPC_ADDRESS          = ":5555"
	API_GATEWAY_ADDRESS   = ":8080"
	FISSION_PROXY_ADDRESS = ":8888"
)

type Options struct {
	Nats                  *NatsOptions
	Fission               *FissionOptions
	InternalRuntime       bool
	Controller            bool
	ApiAdmin              bool
	ApiWorkflow           bool
	ApiHttp               bool
	ApiWorkflowInvocation bool
}

type FissionOptions struct {
	PoolmgrAddr    string
	ControllerAddr string
}

type NatsOptions struct {
	Url     string
	Client  string
	Cluster string
}

// Run serves enabled components in a blocking way
func Run(ctx context.Context, opts *Options) error {
	log.WithField("version", version.VERSION).Info("Starting bundle")

	var es fes.EventStore
	var esPub pubsub.Publisher

	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()

	// Event Stores
	if opts.Nats != nil {
		natsEs := setupNatsEventStoreClient(opts.Nats.Url, opts.Nats.Cluster, opts.Nats.Client)
		es = natsEs
		esPub = natsEs
	}
	if es == nil {
		panic("no event store provided")
	}

	// Caches
	wfiCache := getWorkflowInvocationCache(ctx, esPub)
	wfCache := getWorkflowCache(ctx, esPub)

	// Resolvers and runtimes
	resolvers := map[string]function.Resolver{}
	runtimes := map[string]function.Runtime{}
	if opts.InternalRuntime {
		runtimes["internal"] = setupInternalFunctionRuntime()
		resolvers["internal"] = setupInternalFunctionRuntime()
	}
	if opts.Fission != nil {
		runtimes["fission"] = setupFissionFunctionRuntime(opts.Fission.PoolmgrAddr)
		resolvers["fission"] = setupFissionFunctionResolver(opts.Fission.ControllerAddr)
	}

	// Controller
	if opts.Controller {
		runController(ctx, wfiCache(), wfCache(), es, runtimes, resolvers)
	}

	// Http servers
	if opts.Fission != nil {
		proxySrv := http.Server{Addr: FISSION_PROXY_ADDRESS}
		defer proxySrv.Shutdown(ctx)
		runFissionEnvironmentProxy(proxySrv, es, wfiCache(), wfCache(), resolvers)
	}

	if opts.ApiAdmin {
		runAdminApiServer(grpcServer)
	}

	if opts.ApiWorkflow {
		runWorkflowApiServer(grpcServer, es, resolvers, wfCache())
	}

	if opts.ApiWorkflowInvocation {
		runWorkflowInvocationApiServer(grpcServer, es, wfiCache())
	}

	if opts.ApiAdmin || opts.ApiWorkflow || opts.ApiWorkflowInvocation {
		lis, err := net.Listen("tcp", GRPC_ADDRESS)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		defer lis.Close()
		log.Info("Serving gRPC services at: ", lis.Addr())
		go grpcServer.Serve(lis)
	}

	if opts.ApiHttp {
		apiSrv := http.Server{Addr: API_GATEWAY_ADDRESS}
		defer apiSrv.Shutdown(ctx)
		var admin, wf, wfi string
		if opts.ApiAdmin {
			admin = GRPC_ADDRESS
		}
		if opts.ApiWorkflow {
			wf = GRPC_ADDRESS
		}
		if opts.ApiWorkflowInvocation {
			wfi = GRPC_ADDRESS
		}
		runHttpGateway(ctx, apiSrv, admin, wf, wfi)
	}

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
	return native.NewFunctionEnv(map[string]native.InternalFunction{
		"if":      &builtin.FunctionIf{},
		"noop":    &builtin.FunctionNoop{},
		"compose": &builtin.FunctionCompose{},
		"sleep":   &builtin.FunctionSleep{},
	})
}

func setupFissionFunctionRuntime(poolmgrAddr string) *fission.FunctionEnv {
	poolmgrClient := poolmgrc.MakeClient(poolmgrAddr)
	return fission.NewFunctionEnv(poolmgrClient)
}

func setupFissionFunctionResolver(controllerAddr string) *fission.Resolver {
	controllerClient := controllerc.MakeClient(controllerAddr)
	return fission.NewResolver(controllerClient)
}

func setupNatsEventStoreClient(url string, cluster string, clientId string) *nats.EventStore {
	if clientId == "" {
		clientId = util.Uid()
	}

	conn, err := stan.Connect(cluster, clientId, stan.NatsURL(url))
	if err != nil {
		panic(err)
	}

	log.WithField("cluster", cluster).
		WithField("url", "!redacted!").
		WithField("client", clientId).
		Info("connected to NATS")
	es := nats.NewEventStore(nats.NewWildcardConn(conn))
	es.Watch(fes.Aggregate{Type: "invocation"})
	es.Watch(fes.Aggregate{Type: "workflow"})
	return es
}

func setupWorkflowInvocationCache(ctx context.Context, invocationEventPub pubsub.Publisher) *fes.SubscribedCache {
	invokeSub := invocationEventPub.Subscribe(pubsub.SubscriptionOptions{
		Buf: 50,
		LabelSelector: labels.OrSelector(
			labels.InSelector("aggregate.type", "invocation"),
			labels.InSelector("parent.type", "invocation")),
	})
	wi := func() fes.Aggregator {
		return aggregates.NewWorkflowInvocation("", nil)
	}

	return fes.NewSubscribedCache(ctx, fes.NewMapCache(), wi, invokeSub)
}

func setupWorkflowCache(ctx context.Context, workflowEventPub pubsub.Publisher) *fes.SubscribedCache {
	wfSub := workflowEventPub.Subscribe(pubsub.SubscriptionOptions{
		Buf:           10,
		LabelSelector: labels.InSelector("aggregate.type", "workflow"),
	})
	wb := func() fes.Aggregator {
		return aggregates.NewWorkflow("", nil)
	}
	return fes.NewSubscribedCache(ctx, fes.NewMapCache(), wb, wfSub)
}

func runAdminApiServer(s *grpc.Server) {
	adminServer := &apiserver.GrpcAdminApiServer{}
	apiserver.RegisterAdminAPIServer(s, adminServer)
	log.Infof("Serving admin gRPC API at %s.", GRPC_ADDRESS)
}

func runWorkflowApiServer(s *grpc.Server, es fes.EventStore, resolvers map[string]function.Resolver, wfCache fes.CacheReader) {
	workflowParser := parse.NewResolver(resolvers)
	workflowValidator := parse.NewValidator()
	workflowApi := workflow.NewApi(es, workflowParser)
	workflowServer := apiserver.NewGrpcWorkflowApiServer(workflowApi, workflowValidator, wfCache)
	apiserver.RegisterWorkflowAPIServer(s, workflowServer)
	log.Infof("Serving workflow gRPC API at %s.", GRPC_ADDRESS)
}

func runWorkflowInvocationApiServer(s *grpc.Server, es fes.EventStore, wfiCache fes.CacheReader) {
	invocationApi := invocation.NewApi(es)
	invocationServer := apiserver.NewGrpcInvocationApiServer(invocationApi, wfiCache)
	apiserver.RegisterWorkflowInvocationAPIServer(s, invocationServer)
	log.Infof("Serving workflow invocation gRPC API at %s.", GRPC_ADDRESS)
}

func runHttpGateway(ctx context.Context, gwSrv http.Server, adminApiAddr string, wfApiAddr string, wfiApiAddr string) {
	mux := grpcruntime.NewServeMux()
	grpcOpts := []grpc.DialOption{grpc.WithInsecure()}
	if adminApiAddr != "" {
		err := apiserver.RegisterWorkflowAPIHandlerFromEndpoint(ctx, mux, adminApiAddr, grpcOpts)
		if err != nil {
			panic(err)
		}
	}

	if wfApiAddr != "" {
		err := apiserver.RegisterAdminAPIHandlerFromEndpoint(ctx, mux, wfApiAddr, grpcOpts)
		if err != nil {
			panic(err)
		}
	}

	if wfiApiAddr != "" {
		err := apiserver.RegisterWorkflowInvocationAPIHandlerFromEndpoint(ctx, mux, wfiApiAddr, grpcOpts)
		if err != nil {
			panic(err)
		}
	}

	gwSrv.Handler = handlers.LoggingHandler(os.Stdout, mux)
	go func() {
		err := gwSrv.ListenAndServe()
		log.WithField("err", err).Info("HTTP Gateway exited")
	}()

	log.Info("Serving HTTP API gateway at: ", gwSrv.Addr)
}

func runFissionEnvironmentProxy(proxySrv http.Server, es fes.EventStore, wfiCache fes.CacheReader,
	wfCache fes.CacheReader, resolvers map[string]function.Resolver) {

	workflowParser := parse.NewResolver(resolvers)
	workflowValidator := parse.NewValidator()
	workflowApi := workflow.NewApi(es, workflowParser)
	wfServer := apiserver.NewGrpcWorkflowApiServer(workflowApi, workflowValidator, wfCache)
	wfiApi := invocation.NewApi(es)
	wfiServer := apiserver.NewGrpcInvocationApiServer(wfiApi, wfiCache)
	proxyMux := http.NewServeMux()
	fissionProxyServer := fission.NewFissionProxyServer(wfiServer, wfServer)
	fissionProxyServer.RegisterServer(proxyMux)

	proxySrv.Handler = handlers.LoggingHandler(os.Stdout, proxyMux)
	go proxySrv.ListenAndServe()
	log.Info("Serving HTTP Fission Proxy at: ", proxySrv.Addr)
}

func runController(ctx context.Context, invocationCache fes.CacheReader, wfCache fes.CacheReader, es fes.EventStore,
	fnRuntimes map[string]function.Runtime, fnResolvers map[string]function.Resolver) {

	workflowApi := workflow.NewApi(es, parse.NewResolver(fnResolvers))
	functionApi := function.NewApi(fnRuntimes, es)
	invocationApi := invocation.NewApi(es)
	s := &scheduler.WorkflowScheduler{}
	pf := typedvalues.DefaultParserFormatter
	ep := expr.NewJavascriptExpressionParser(pf)
	invocationCtrl := controller.NewInvocationController(invocationCache, wfCache, s, functionApi, invocationApi, ep)
	workflowCtrl := controller.NewWorkflowController(wfCache, workflowApi)

	ctrl := controller.NewMetaController(invocationCtrl, workflowCtrl)

	go ctrl.Run(ctx)
	log.Info("Setup controller component.")
}
