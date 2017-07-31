package app

import (
	"context"
	"net"
	"net/http"

	"github.com/fission/fission-workflow/pkg/api"
	"github.com/fission/fission-workflow/pkg/apiserver"
	"github.com/fission/fission-workflow/pkg/cache"
	"github.com/fission/fission-workflow/pkg/controller"
	"github.com/fission/fission-workflow/pkg/eventstore/nats"
	"github.com/fission/fission-workflow/pkg/projector/project/invocation"
	"github.com/fission/fission-workflow/pkg/scheduler"
	"github.com/fission/fission/poolmgr/client"
	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/nats-io/go-nats-streaming"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

const (
	GRPC_ADDRESS        = ":5555"
	API_GATEWAY_ADDRESS = ":8080"
)

// Blocking
func Run(ctx context.Context) {
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
	natsConn, err := stan.Connect("test-cluster", "test-client")
	if err != nil {
		panic(err)
	}
	natsClient := nats.New(nats.NewConn(natsConn))
	cache := cache.NewMapCache()

	// Fission client
	poolmgrClient := client.MakeClient("http://192.168.99.100:32101")

	// Setup API
	workflowApi := api.NewWorkflowApi(natsClient)
	invocationApi := api.NewInvocationApi(natsClient)
	functionApi := api.NewFissionFunctionApi(poolmgrClient)

	// API gRPC Server
	workflowServer := apiserver.NewGrpcWorkflowApiServer(workflowApi)
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
	go http.ListenAndServe(API_GATEWAY_ADDRESS, mux)

	// Controller
	invocationProjector := invocation.NewInvocationProjector(natsClient, cache)
	s := &scheduler.WorkflowScheduler{}
	ctr := controller.NewController(invocationProjector, s)
	defer ctr.Close()
	go ctr.Run()

	// Serve gRPC server
	log.Info("Serving gRPC services at: ", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
