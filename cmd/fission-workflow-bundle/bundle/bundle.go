package bundle

import (
	"os"

	"context"

	"net/http"

	"net"

	"fmt"

	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/api/invocation"
	"github.com/fission/fission-workflow/pkg/api/workflow"
	"github.com/fission/fission-workflow/pkg/api/workflow/parse"
	"github.com/fission/fission-workflow/pkg/apiserver"
	"github.com/fission/fission-workflow/pkg/controller"
	"github.com/fission/fission-workflow/pkg/controller/query"
	"github.com/fission/fission-workflow/pkg/fes"
	"github.com/fission/fission-workflow/pkg/fes/eventstore/nats"
	"github.com/fission/fission-workflow/pkg/fnenv/fission"
	"github.com/fission/fission-workflow/pkg/fnenv/native"
	"github.com/fission/fission-workflow/pkg/fnenv/native/builtin"
	"github.com/fission/fission-workflow/pkg/scheduler"
	"github.com/fission/fission-workflow/pkg/types/aggregates"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
	"github.com/fission/fission-workflow/pkg/util/labels"
	"github.com/fission/fission-workflow/pkg/util/pubsub"
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
	FISSION_PROXY_ADDRESS = ":8090"
)

type Options struct {
	Nats                  *NatsOptions
	Fission               *FissionOptions
	InternalRuntime       interface{}
	Controller            interface{}
	ApiAdmin              interface{}
	ApiWorkflow           interface{}
	ApiHttp               interface{}
	ApiWorkflowInvocation interface{}
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
		fmt.Println(es)
		panic("no event store provided")
	}

	// Caches
	wfiCache := getWorkflowInvocationCache(ctx, esPub)
	wfCache := getWorkflowCache(ctx, esPub)

	// Resolvers and runtimes
	resolvers := map[string]function.Resolver{}
	runtimes := map[string]function.Runtime{}
	if opts.InternalRuntime != nil {
		runtimes["internal"] = setupInternalFunctionRuntime()
		resolvers["internal"] = setupInternalFunctionRuntime()
	}
	if opts.Fission != nil {
		runtimes["fission"] = setupFissionFunctionRuntime(opts.Fission.PoolmgrAddr)
		resolvers["fission"] = setupFissionFunctionResolver(opts.Fission.ControllerAddr)
	}
	if len(runtimes) == 0 || len(resolvers) == 0 {
		panic("No runtimes or resolvers specified!")
	}

	// Controller
	if opts.Controller != nil {
		runController(ctx, wfiCache(), wfCache(), es, runtimes)
	}

	// Http servers
	if opts.Fission != nil {
		proxySrv := http.Server{Addr: FISSION_PROXY_ADDRESS}
		defer proxySrv.Shutdown(ctx)
		runFissionEnvironmentProxy(proxySrv, es, wfiCache(), wfCache(), resolvers)
	}

	if opts.ApiAdmin != nil {
		runAdminApiServer(grpcServer)
	}

	if opts.ApiWorkflow != nil {
		runWorkflowApiServer(grpcServer, es, resolvers, wfCache())
	}

	if opts.ApiWorkflowInvocation != nil {
		runWorkflowInvocationApiServer(grpcServer, es, wfiCache())
	}

	if opts.ApiHttp != nil {
		apiSrv := http.Server{Addr: API_GATEWAY_ADDRESS}
		defer apiSrv.Shutdown(ctx)
		runHttpGateway(ctx, apiSrv)
	}

	if opts.ApiAdmin != nil || opts.ApiWorkflow != nil || opts.ApiWorkflowInvocation != nil {
		lis, err := net.Listen("tcp", GRPC_ADDRESS)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		defer lis.Close()
		log.Info("Serving gRPC services at: ", lis.Addr())
		go grpcServer.Serve(lis)
	}

	<-ctx.Done()
	log.Info("Shutting down...")
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
		"if":    &builtin.FunctionIf{},
		"noop":  &builtin.FunctionNoop{},
		"sleep": &builtin.FunctionSleep{},
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
	conn, err := stan.Connect(cluster, clientId, stan.NatsURL(url))
	if err != nil {
		panic(err)
	}

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
	log.Info("Serving admin API.")
}

func runWorkflowApiServer(s *grpc.Server, es fes.EventStore, resolvers map[string]function.Resolver, wfCache fes.CacheReader) {
	workflowParser := parse.NewResolver(resolvers)
	workflowValidator := parse.NewValidator()
	workflowApi := workflow.NewApi(es, workflowParser)
	workflowServer := apiserver.NewGrpcWorkflowApiServer(workflowApi, workflowValidator, wfCache)
	apiserver.RegisterWorkflowAPIServer(s, workflowServer)
	log.Info("Serving workflow API.")
}

func runWorkflowInvocationApiServer(s *grpc.Server, es fes.EventStore, invocationCache fes.CacheReader) {
	invocationApi := invocation.NewApi(es, invocationCache)
	invocationServer := apiserver.NewGrpcInvocationApiServer(invocationApi)
	apiserver.RegisterWorkflowInvocationAPIServer(s, invocationServer)
	log.Info("Serving workflow invocation API.")
}

func runHttpGateway(ctx context.Context, gwSrv http.Server) {
	mux := grpcruntime.NewServeMux()
	grpcOpts := []grpc.DialOption{grpc.WithInsecure()}
	err := apiserver.RegisterWorkflowAPIHandlerFromEndpoint(ctx, mux, gwSrv.Addr, grpcOpts)
	if err != nil {
		panic(err)
	}
	err = apiserver.RegisterAdminAPIHandlerFromEndpoint(ctx, mux, gwSrv.Addr, grpcOpts)
	if err != nil {
		panic(err)
	}
	err = apiserver.RegisterWorkflowInvocationAPIHandlerFromEndpoint(ctx, mux, gwSrv.Addr, grpcOpts)
	if err != nil {
		panic(err)
	}

	gwSrv.Handler = mux //handlers.LoggingHandler(os.Stdout, mux)
	go gwSrv.ListenAndServe()
	log.Info("Serving HTTP API gateway at: ", gwSrv.Addr)
}

func runFissionEnvironmentProxy(proxySrv http.Server, es fes.EventStore, wfiCache fes.CacheReader,
	wfCache fes.CacheReader, resolvers map[string]function.Resolver) {

	workflowParser := parse.NewResolver(resolvers)
	workflowValidator := parse.NewValidator()
	workflowApi := workflow.NewApi(es, workflowParser)
	wfServer := apiserver.NewGrpcWorkflowApiServer(workflowApi, workflowValidator, wfCache)
	wfiApi := invocation.NewApi(es, wfiCache)
	wfiServer := apiserver.NewGrpcInvocationApiServer(wfiApi)
	proxyMux := http.NewServeMux()
	fissionProxyServer := fission.NewFissionProxyServer(wfiServer, wfServer)
	fissionProxyServer.RegisterServer(proxyMux)

	proxySrv.Handler = handlers.LoggingHandler(os.Stdout, proxyMux)
	go proxySrv.ListenAndServe()
	log.Info("Serving HTTP Fission Proxy at: ", proxySrv.Addr)
}

func runController(ctx context.Context, invocationCache fes.CacheReader, wfCache fes.CacheReader, es fes.EventStore,
	fnRuntimes map[string]function.Runtime) {

	functionApi := function.NewApi(fnRuntimes, es)
	invocationApi := invocation.NewApi(es, invocationCache)
	s := &scheduler.WorkflowScheduler{}
	pf := typedvalues.DefaultParserFormatter
	ep := query.NewJavascriptExpressionParser(pf)
	ctr := controller.NewController(invocationCache, wfCache, s, functionApi, invocationApi, ep)
	go ctr.Run(ctx)
	log.Info("Setup controller component.")

	// TODO properly shutdown
}
