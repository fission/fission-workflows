package main

import (
	"os"

	"context"

	"net/http"

	"net"

	"github.com/fission/fission-workflow/cmd/workflow-engine/app"
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
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	natsio "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli"
	"google.golang.org/grpc"
)

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

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
			Name: "fission",
			Usage: "Use Fission as a function environment",
		},
		cli.StringFlag{
			Name: "fission-poolmgr",
			Usage: "Address of the poolmgr",
			Value: "http://poolmgr.fission",
			EnvVar: "FNENV_FISSION_POOLMGR",
		},
		cli.StringFlag{
			Name: "fission-controller",
			Usage: "Address of the controller for resolving functions",
			Value: "http://controller.fission",
		},

		// Components
		cli.BoolFlag{
			Name: "controller",
			Usage: "Run the controller",
		},
		cli.BoolFlag{
			Name: "api-http",
			Usage: "Serve the http apis of the apis",
		},
		cli.BoolFlag{
			Name: "api-workflow-invocation",
			Usage: "Serve the workflow invocation gRPC api",
		},
		cli.BoolFlag{
			Name: "api-workflow",
			Usage: "Serve the workflow gRPC api",
		},
		cli.BoolFlag{
			Name: "api-admin",
			Usage: "Serve the admin gRPC api",
		},

	}

	cliApp.Action = run
	cliApp.Run(os.Args)
}

// blocking
func run(c *cli.Context) error {
	ctx := context.Background()

	var es fes.EventStore
	var esPub pubsub.Publisher

	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()

	// Event Stores
	if c.Bool("nats") {
		natsEs := setupNatsEventStore(c.String("nats-url"), c.String("nats-cluster"),
			c.String("nats-client"))
		es = natsEs
		esPub = natsEs
	}
	if es != nil {
		panic("no event store provided")
	}

	// Caches
	wfiCache := getWorkflowInvocationCache(ctx, esPub)
	wfCache := getWorkflowCache(ctx, esPub)

	// Resolvers and runtimes
	resolvers := map[string]function.Resolver{}
	runtimes := map[string]function.Runtime{}
	if c.Bool("internal") {
		runtime["internal"] = setupInternalFunctionRuntime()
		resolvers["internal"] = setupInternalFunctionRuntime()
	}
	if c.Bool("fission") {
		runtime["fission"] = setupFissionFunctionRuntime(c.String("fission-poolmgr"))
		resolvers["fission"] = setupFissionFunctionResolver(c.String("fission-controller"))
	}
	if len(runtimes) == 0 || len(resolvers) == 0 {
		panic("No runtimes or resolvers specified!")
	}

	// Controller
	if c.Bool("controller") {
		runController(ctx, wfiCache(), wfCache(), es, nil)
	}

	// Http servers
	if c.Bool("fission") {
		proxySrv := http.Server{Addr: app.FISSION_PROXY_ADDRESS}
		defer proxySrv.Shutdown(ctx)
		runFissionEnvironmentProxy(proxySrv, es, wfiCache())
	}

	if c.Bool("api-admin") {
		runAdminApiServer(grpcServer)
	}

	if c.Bool("api-workflow") {
		runWorkflowApiServer(grpcServer, es, nil, wfCache())
	}

	if c.Bool("api-workflow-invocation") {
		runWorkflowInvocationApiServer(grpcServer, es, wfiCache())
	}

	if c.Bool("api-http") {
		apiSrv := http.Server{Addr: app.API_GATEWAY_ADDRESS}
		defer apiSrv.Shutdown(ctx)
		runHttpGateway(ctx, apiSrv)
	}

	if c.Bool("api-admin") || c.Bool("api-workflow") || c.Bool("api-workflow-invocation") {
		lis, err := net.Listen("tcp", app.GRPC_ADDRESS)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
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

func setupNatsEventStore(url string, cluster string, clientId string) *nats.EventStore {
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
}

func runWorkflowApiServer(s *grpc.Server, es fes.EventStore, resolvers map[string]function.Resolver, wfCache fes.CacheReader) {
	workflowParser := parse.NewResolver(resolvers)
	workflowValidator := parse.NewValidator()
	workflowApi := workflow.NewApi(es, workflowParser)
	workflowServer := apiserver.NewGrpcWorkflowApiServer(workflowApi, workflowValidator, wfCache)
	apiserver.RegisterWorkflowAPIServer(s, workflowServer)
}

func runWorkflowInvocationApiServer(s *grpc.Server, es fes.EventStore, invocationCache fes.CacheReader) {
	invocationApi := invocation.NewApi(es, invocationCache)
	invocationServer := apiserver.NewGrpcInvocationApiServer(invocationApi)
	apiserver.RegisterWorkflowInvocationAPIServer(s, invocationServer)
}

func runHttpGateway(ctx context.Context, gwSrv http.Server) {
	mux := runtime.NewServeMux()
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

	gwSrv.Handler = handlers.LoggingHandler(os.Stdout, mux)
	go gwSrv.ListenAndServe()
	log.Info("Serving HTTP API gateway at: ", gwSrv.Addr)
}

func runFissionEnvironmentProxy(proxySrv http.Server, es fes.EventStore, wfiCache fes.CacheReader) {
	wfiApi := invocation.NewApi(es, wfiCache)
	wfiServer := apiserver.NewGrpcInvocationApiServer(wfiApi)
	proxyMux := http.NewServeMux()
	fissionProxyServer := fission.NewFissionProxyServer(wfiServer)
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
	defer ctr.Close()
	go ctr.Run(ctx)
}
