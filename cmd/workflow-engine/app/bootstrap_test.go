package app

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"os"
	"testing"

	"strings"

	"io"
	"net"

	"reflect"

	"github.com/fission/fission-workflow/cmd/workflow-engine/app"
	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/apiserver"
	"github.com/fission/fission-workflow/pkg/fnenv/test"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/test/"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var env *app.Options

const (
	UID_FUNC_ECHO = "FuncUid1"
)

var mockFuncResolves = map[string]string{
	"echo": UID_FUNC_ECHO,
}

var mockFuncs = map[string]test.MockFunc{
	UID_FUNC_ECHO: echo,
}

func echo(spec *types.FunctionInvocationSpec) ([]byte, error) {
	val, ok := spec.Input[types.INPUT_MAIN]
	if !ok {
		return []byte{}, nil
	}
	return []byte(val), nil
}

func TestMain(m *testing.M) {

	ctx, cancelFn := context.WithCancel(context.Background())
	env = setup(ctx)
	exitCode := m.Run()
	defer os.Exit(exitCode)
	// Teardown
	cancelFn()
	<-time.After(time.Duration(2) * time.Second) // Needed in order to let context cancel propagate
}

// Tests the submission of a workflow
func TestWorkflowCreate(t *testing.T) {
	ctx := context.Background()
	conn, _ := grpc.Dial(env.GrpcApiServerAddress, grpc.WithInsecure())
	cl := apiserver.NewWorkflowAPIClient(conn)

	// Test workflow creation
	spec := &types.WorkflowSpec{
		Name:    "TestWorkflowCreate_WF",
		Version: "unknown",
		Src: &types.WorkflowDefinition{
			ApiVersion: "v1",
			OutputTask: "fakeFinalTask",
			Tasks: map[string]*types.Task{
				"fakeFinalTask": {
					Type: types.TaskType_FUNCTION.String(),
					Name: "echo",
				},
			},
		},
	}
	wfId, err := cl.Create(ctx, spec)
	if err != nil {
		t.Fatal(err)
	}
	if wfId == nil || len(wfId.GetId()) == 0 {
		t.Errorf("Invalid ID returned '%v'", wfId)
	}

	// Test workflow list
	l, err := cl.List(ctx, &empty.Empty{})
	if err != nil {
		t.Error(err)
	}
	if len(l.Workflows) != 1 || l.Workflows[0] != wfId.Id {
		t.Errorf("Listed workflows '%v' did not match expected workflow '%s'", l.Workflows, wfId.Id)
	}

	// Test workflow get
	wf, err := cl.Get(ctx, wfId)
	if err != nil {
		t.Error(err)
	}

	if !reflect.DeepEqual(wf.Spec, spec) {
		t.Error("Specs of created and fetched do not match!")
	}

	if wf.Status.Status != types.WorkflowStatus_READY {
		t.Errorf("Workflow status is not ready, but '%v'", wf.Status.Status)
	}
}

func TestWorkflowInvocation(t *testing.T) {
	ctx := context.Background()
	conn, _ := grpc.Dial(env.GrpcApiServerAddress, grpc.WithInsecure())
	cl := apiserver.NewWorkflowAPIClient(conn)
	wi := apiserver.NewWorkflowInvocationAPIClient(conn)

	// Test workflow creation
	wf := "TestWorkflowInvocation_WF"
	wfSpec := &types.WorkflowSpec{
		Name:    wf,
		Version: "unknown",
		Src: &types.WorkflowDefinition{
			ApiVersion: "v1",
			OutputTask: "fakeFinalTask",
			Tasks: map[string]*types.Task{
				"fakeFinalTask": {
					Type: types.TaskType_FUNCTION.String(),
					Name: "echo",
					// TODO dependency input
					Dependencies: map[string]*types.TaskDependencyParameters{
						"FirstTask": {},
					},
				},
				"FirstTask": {
					Type: types.TaskType_FUNCTION.String(),
					// TODO input from workflow
					Name: "echo",
				},
			},
		},
	}
	wfResp, err := cl.Create(ctx, wfSpec)
	if err != nil {
		t.Fatal(err)
	}
	if wfResp == nil || len(wfResp.GetId()) == 0 {
		t.Errorf("Invalid ID returned '%v'", wfResp)
	}

	// Create invocation
	expectedOutput := "Hello world!"
	wiSpec := &types.WorkflowInvocationSpec{
		WorkflowId: wfResp.Id,
		Inputs: map[string]string{
			types.INPUT_MAIN: expectedOutput,
		},
	}
	wiId, err := wi.Invoke(ctx, wiSpec)
	if err != nil {
		t.Fatal(err)
	}

	// Test invocation list
	l, err := wi.List(ctx, &empty.Empty{})
	if err != nil {
		t.Error(err)
	}
	if len(l.Invocations) != 1 || l.Invocations[0] != wiId.Id {
		t.Errorf("Listed invocations '%v' did not match expected invocation '%s'", l.Invocations, wiId.Id)
	}

	// Test invocation get, give some slack to actually invoke it
	var invocation *types.WorkflowInvocation
	deadline := time.Now().Add(time.Duration(1) * time.Second)
	tick := time.NewTicker(time.Duration(100) * time.Millisecond)
	for ti := range tick.C {
		invoc, err := wi.Get(ctx, wiId)
		if err != nil {
			t.Error(err)
		}
		if invoc.Status.Status.Finished() || ti.After(deadline) {
			invocation = invoc
			tick.Stop()
			break
		}
	}

	if !reflect.DeepEqual(invocation.Spec, wiSpec) {
		t.Error("Specs of created and fetched do not match!")
	}

	if !strings.EqualFold(invocation.Status.Output, expectedOutput) {
		t.Errorf("Output '%s' does not match expected output '%s'", invocation.Status.Output, expectedOutput)
	}

	if !invocation.Status.Status.Successful() {
		t.Errorf("Invocation status is not succesfull,s but '%v", invocation.Status.Status)
	}
}

func setup(ctx context.Context) *app.Options {
	// TODO Maybe replace with actual Fission deployment
	mockFunctionResolver := &test.MockFunctionResolver{mockFuncResolves}
	mockFunctionRuntime := &test.MockRuntimeEnv{Functions: mockFuncs, Results: map[string]*types.FunctionInvocation{}}

	esOpts := setupEventStore(ctx)
	opts := &app.Options{
		FunctionRegistry: map[string]function.Resolver{
			"mock": mockFunctionResolver,
		},
		FunctionRuntimeEnv: map[string]function.Runtime{
			"mock": mockFunctionRuntime,
		},
		EventStore:           esOpts,
		GrpcApiServerAddress: app.GRPC_ADDRESS,
		HttpApiServerAddress: app.API_GATEWAY_ADDRESS,
		FissionProxyAddress:  app.FISSION_PROXY_ADDRESS,
	}
	go app.Run(ctx, opts)

	return opts
}

func setupEventStore(ctx context.Context) *app.EventStoreOptions {
	clusterId := fmt.Sprintf("fission-workflow-e2e-%d", time.Now().UnixNano())
	port, err := findFreePort()
	if err != nil {
		panic(err)
	}
	address := "0.0.0.0"
	flags := strings.Split(fmt.Sprintf("-cid %s -p %d -a %s", clusterId, port, address), " ")
	logrus.Info(flags)
	cmd := exec.CommandContext(ctx, "nats-streaming-server", flags...)
	stdOut, _ := cmd.StdoutPipe()
	stdErr, _ := cmd.StderrPipe()
	go io.Copy(os.Stdout, stdOut)
	go io.Copy(os.Stdout, stdErr)
	err = cmd.Start()
	if err != nil {
		panic(err)
	}
	esOpts := &app.EventStoreOptions{
		Cluster: clusterId,
		Type:    "NATS",
		Url:     fmt.Sprintf("nats://%s:%d", address, port),
	}

	logrus.WithField("config", esOpts).Info("Setting up NATS server")

	// wait for a bit to set it up
	awaitCtx, _ := context.WithTimeout(ctx, time.Duration(10)*time.Second)
	err = waitForNats(awaitCtx, esOpts.Url, esOpts.Cluster)
	if err != nil {
		logrus.Error(err)
	}
	logrus.WithField("config", esOpts).Info("NATS Server running")

	return esOpts
}

// Wait for NATS to come online, ignoring ErrNoServer as it could mean that NATS is still being setup
func waitForNats(ctx context.Context, url string, cluster string) error {
	conn, err := stan.Connect(cluster, "setupEventStore-alive-test", stan.NatsURL(url), stan.ConnectWait(time.Duration(10)*time.Second))
	if err == nats.ErrNoServers {
		logrus.WithFields(logrus.Fields{
			"cluster": cluster,
			"url":     url,
		}).Warn(err)
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
