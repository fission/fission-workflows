package app

import (
	"context"
	"net"
	"net/http"

	"os"

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
	"github.com/fission/fission-workflow/pkg/scheduler"
	"github.com/fission/fission-workflow/pkg/types/aggregates"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
	"github.com/fission/fission-workflow/pkg/util/labels"
	"github.com/fission/fission-workflow/pkg/util/pubsub"
	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
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
	FunctionRuntimeEnv   map[string]function.Runtime
	FunctionRegistry     map[string]function.Resolver
	EventStore           *EventStoreOptions
	FissionProxyAddress  string
	GrpcApiServerAddress string
	HttpApiServerAddress string
}

type EventStoreOptions struct {
	Url     string
	Type    string
	Cluster string
}

// TODO scratch, should be cleaned up
// Blocking
func Run(ctx context.Context, opts *Options) error {
	// (shared) gRPC server
	lis, err := net.Listen("tcp", opts.GrpcApiServerAddress)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	conn, err := grpc.Dial(opts.GrpcApiServerAddress, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()
	defer conn.Close()
	defer lis.Close()

	// EventStore
	stanConn, err := stan.Connect(opts.EventStore.Cluster, "test-client", stan.NatsURL(opts.EventStore.Url))
	if err != nil {
		panic(err)
	}
	//natsClient := inats.New(inats.NewConn(stanConn))
	//cache := cache.NewMapCache()

	workflowParser := parse.NewResolver(opts.FunctionRegistry)
	workflowValidator := parse.NewValidator()

	es := nats.NewEventStore(nats.NewWildcardConn(stanConn))
	es.Watch(fes.Aggregate{Type: "workflow"})
	es.Watch(fes.Aggregate{Type: "invocation"})
	defer es.Close()

	// Create updateble invocationCache
	invokeSub := es.Subscribe(pubsub.SubscriptionOptions{
		Buf: 50,
		LabelSelector: labels.OrSelector(
			labels.InSelector("aggregate.type", "invocation"),
			labels.InSelector("parent.type", "invocation")),
	})
	wi := func() fes.Aggregator {
		return aggregates.NewWorkflowInvocation("", nil)
	}
	invocationCache := fes.NewSubscribedCache(ctx, fes.NewMapCache(), wi, invokeSub)

	// Create updateble workflowCache
	wfSub := es.Subscribe(pubsub.SubscriptionOptions{
		Buf:           10,
		LabelSelector: labels.InSelector("aggregate.type", "workflow"),
	})
	wb := func() fes.Aggregator {
		return aggregates.NewWorkflow("", nil)
	}
	wfCache := fes.NewSubscribedCache(ctx, fes.NewMapCache(), wb, wfSub)

	// Setup API
	invocationApi := invocation.NewApi(es, invocationCache)

	workflowApi := workflow.NewApi(es, workflowParser)
	functionApi := function.NewFissionFunctionApi(opts.FunctionRuntimeEnv, es)
	//err = workflowProjector.Watch("workflows.>")
	//if err != nil {
	//	log.Warnf("Failed to watch for workflows, because '%v'.", err)
	//}

	// API gRPC Server
	workflowServer := apiserver.NewGrpcWorkflowApiServer(workflowApi, workflowValidator, wfCache)
	adminServer := &apiserver.GrpcAdminApiServer{}
	invocationServer := apiserver.NewGrpcInvocationApiServer(invocationApi)

	apiserver.RegisterWorkflowAPIServer(grpcServer, workflowServer)
	apiserver.RegisterAdminAPIServer(grpcServer, adminServer)
	apiserver.RegisterWorkflowInvocationAPIServer(grpcServer, invocationServer)

	// API Gateway server
	mux := runtime.NewServeMux()
	grpcOpts := []grpc.DialOption{grpc.WithInsecure()}
	err = apiserver.RegisterWorkflowAPIHandlerFromEndpoint(ctx, mux, opts.GrpcApiServerAddress, grpcOpts)
	if err != nil {
		panic(err)
	}
	err = apiserver.RegisterAdminAPIHandlerFromEndpoint(ctx, mux, opts.GrpcApiServerAddress, grpcOpts)
	if err != nil {
		panic(err)
	}
	err = apiserver.RegisterWorkflowInvocationAPIHandlerFromEndpoint(ctx, mux, opts.GrpcApiServerAddress, grpcOpts)
	if err != nil {
		panic(err)
	}

	// fission proxy
	proxyMux := http.NewServeMux()
	fissionProxyServer := fission.NewFissionProxyServer(invocationServer)
	fissionProxyServer.RegisterServer(proxyMux)

	proxySrv := http.Server{Addr: opts.FissionProxyAddress}
	proxySrv.Handler = handlers.LoggingHandler(os.Stdout, proxyMux)
	go proxySrv.ListenAndServe()
	log.Info("Serving HTTP Fission Proxy at: ", opts.FissionProxyAddress)

	apiSrv := http.Server{Addr: opts.HttpApiServerAddress}
	apiSrv.Handler = handlers.LoggingHandler(os.Stdout, mux)
	go apiSrv.ListenAndServe()
	log.Info("Serving HTTP API gateway at: ", opts.HttpApiServerAddress)

	// Controller
	s := &scheduler.WorkflowScheduler{}
	pf := typedvalues.DefaultParserFormatter
	ep := query.NewJavascriptExpressionParser(pf)
	ctr := controller.NewController(invocationCache, wfCache, s, functionApi, invocationApi, ep)
	defer ctr.Close()
	go ctr.Run(ctx)

	// Serve gRPC server
	log.Info("Serving gRPC services at: ", lis.Addr())
	go grpcServer.Serve(lis)

	<-ctx.Done()
	shutdownCtx := context.Background()
	log.Debug("Shutting down servers...")
	grpcServer.GracefulStop() // Close
	apiSrv.Shutdown(shutdownCtx)
	proxySrv.Shutdown(shutdownCtx)
	log.Debug("Servers shutdown successfully.")
	return shutdownCtx.Err()
}
