package app

import (
	"context"
	"net"
	"net/http"

	"os"

	"github.com/fission/fission-workflow/pkg/api"
	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/api/workflow"
	"github.com/fission/fission-workflow/pkg/apiserver"
	"github.com/fission/fission-workflow/pkg/cache"
	"github.com/fission/fission-workflow/pkg/controller"
	inats "github.com/fission/fission-workflow/pkg/eventstore/nats"
	"github.com/fission/fission-workflow/pkg/projector/project/invocation"
	"github.com/fission/fission-workflow/pkg/scheduler"
	"github.com/fission/fission/controller/client"
	poolmgr "github.com/fission/fission/poolmgr/client"
	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/nats-io/go-nats-streaming"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	GRPC_ADDRESS        = ":5555"
	API_GATEWAY_ADDRESS = ":8080"
)

type Options struct {
	EventStore *EventStoreOptions
}

type EventStoreOptions struct {
	Url     string
	Type    string
	Cluster string
}

// TODO scratch, should be cleaned up
// Blocking
func Run(ctx context.Context, options *Options) error {
	// (shared) gRPC server
	lis, err := net.Listen("tcp", GRPC_ADDRESS)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	conn, err := grpc.Dial(GRPC_ADDRESS, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	grpcServer := grpc.NewServer()
	defer grpcServer.GracefulStop()
	defer conn.Close()
	defer lis.Close()

	// EventStore
	stanConn, err := stan.Connect("fissionMQTrigger", "test-client", stan.NatsURL(options.EventStore.Url))
	if err != nil {
		panic(err)
	}
	natsClient := inats.New(inats.NewConn(stanConn))
	cache := cache.NewMapCache()

	// Fission client
	poolmgrClient := poolmgr.MakeClient("http://192.168.99.100:32101")
	controllerClient := client.MakeClient("http://192.168.99.100:31313")

	workflowParser := workflow.NewParser(controllerClient)
	workflowValidator := workflow.NewValidator()
	invocationProjector := invocation.NewInvocationProjector(natsClient, cache)
	err = invocationProjector.Watch("invocation.>")
	if err != nil {
		panic(err)
	}
	// Setup API
	workflowApi := workflow.NewApi(natsClient, workflowParser)
	invocationApi := api.NewInvocationApi(natsClient, invocationProjector)
	functionApi := function.NewFissionFunctionApi(poolmgrClient)
	err = workflowApi.Projector.Watch("workflows.>")
	if err != nil {
		log.Warnf("Failed to watch for workflows, because '%v'.", err)
	}

	// API gRPC Server
	workflowServer := apiserver.NewGrpcWorkflowApiServer(workflowApi, workflowValidator)
	adminServer := &apiserver.GrpcAdminApiServer{}
	invocationServer := apiserver.NewGrpcInvocationApiServer(invocationApi)
	functionServer := apiserver.NewGrpcFunctionApiServer(functionApi)

	apiserver.RegisterWorkflowAPIServer(grpcServer, workflowServer)
	apiserver.RegisterAdminAPIServer(grpcServer, adminServer)
	apiserver.RegisterWorkflowInvocationAPIServer(grpcServer, invocationServer)
	apiserver.RegisterFunctionEnvApiServer(grpcServer, functionServer)

	// API Gateway server
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}
	err = apiserver.RegisterWorkflowAPIHandlerFromEndpoint(ctx, mux, GRPC_ADDRESS, opts)
	if err != nil {
		panic(err)
	}
	err = apiserver.RegisterAdminAPIHandlerFromEndpoint(ctx, mux, GRPC_ADDRESS, opts)
	if err != nil {
		panic(err)
	}
	err = apiserver.RegisterWorkflowInvocationAPIHandlerFromEndpoint(ctx, mux, GRPC_ADDRESS, opts)
	if err != nil {
		panic(err)
	}
	err = apiserver.RegisterFunctionEnvApiHandlerFromEndpoint(ctx, mux, GRPC_ADDRESS, opts)
	if err != nil {
		panic(err)
	}

	log.Info("Serving HTTP API gateway at: ", API_GATEWAY_ADDRESS)
	go http.ListenAndServe(API_GATEWAY_ADDRESS, handlers.LoggingHandler(os.Stdout, mux))

	// Controller
	s := &scheduler.WorkflowScheduler{}
	ctr := controller.NewController(invocationProjector, workflowApi.Projector, s, functionApi, invocationApi, natsClient)
	defer ctr.Close()
	go ctr.Run()

	// Serve gRPC server
	log.Info("Serving gRPC services at: ", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

	return nil
}
