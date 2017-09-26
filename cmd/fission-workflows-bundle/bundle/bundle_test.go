package bundle

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

	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestMain(m *testing.M) {

	ctx, cancelFn := context.WithCancel(context.Background())
	setup(ctx)

	time.Sleep(time.Duration(4) * time.Second)

	exitCode := m.Run()
	defer os.Exit(exitCode)
	// Teardown
	cancelFn()
	<-time.After(time.Duration(2) * time.Second) // Needed in order to let context cancel propagate
}

// Tests the submission of a workflow
func TestWorkflowCreate(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.Dial(GRPC_ADDRESS, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	cl := apiserver.NewWorkflowAPIClient(conn)

	// Test workflow creation
	spec := &types.WorkflowSpec{
		ApiVersion: "v1",
		OutputTask: "fakeFinalTask",
		Tasks: map[string]*types.Task{
			"fakeFinalTask": {
				FunctionRef: "noop",
			},
		},
	}
	wfId, err := cl.Create(ctx, spec)
	assert.NoError(t, err)
	assert.NotNil(t, wfId)
	assert.NotEmpty(t, wfId.GetId())

	time.Sleep(time.Duration(3) * time.Second)
	// Test workflow list
	l, err := cl.List(ctx, &empty.Empty{})
	assert.NoError(t, err)
	if len(l.Workflows) != 1 || l.Workflows[0] != wfId.Id {
		t.Errorf("Listed workflows '%v' did not match expected workflow '%s'", l.Workflows, wfId.Id)
	}

	// Test workflow get
	wf, err := cl.Get(ctx, wfId)
	assert.NoError(t, err)
	assert.Equal(t, wf.Spec, spec)
	assert.Equal(t, wf.Status.Status, types.WorkflowStatus_READY)
}

func TestWorkflowInvocation(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.Dial(GRPC_ADDRESS, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	cl := apiserver.NewWorkflowAPIClient(conn)
	wi := apiserver.NewWorkflowInvocationAPIClient(conn)

	// Test workflow creation
	wfSpec := &types.WorkflowSpec{
		ApiVersion: "v1",
		OutputTask: "fakeFinalTask",
		Tasks: map[string]*types.Task{
			"fakeFinalTask": {
				FunctionRef: "noop",
				Inputs: map[string]*types.TypedValue{
					types.INPUT_MAIN: typedvalues.Expr("{$.Tasks.FirstTask.Output}"),
				},
				Requires: map[string]*types.TaskDependencyParameters{
					"FirstTask": {},
				},
			},
			"FirstTask": {
				FunctionRef: "noop",
				Inputs: map[string]*types.TypedValue{
					types.INPUT_MAIN: typedvalues.Expr("{$.Invocation.Inputs.default.toUpperCase()}"),
				},
			},
		},
	}
	wfResp, err := cl.Create(ctx, wfSpec)
	assert.NoError(t, err)
	if wfResp == nil || len(wfResp.GetId()) == 0 {
		t.Errorf("Invalid ID returned '%v'", wfResp)
	}

	// Create invocation
	expectedOutput := "Hello world!"
	tv, err := typedvalues.Parse(expectedOutput)
	etv, err := typedvalues.Parse(strings.ToUpper(expectedOutput))
	assert.NoError(t, err)

	wiSpec := &types.WorkflowInvocationSpec{
		WorkflowId: wfResp.Id,
		Inputs: map[string]*types.TypedValue{
			types.INPUT_MAIN: tv,
		},
	}
	result, err := wi.InvokeSync(ctx, wiSpec)
	assert.NoError(t, err)
	wiId := result.Metadata.Id

	// Test invocation list
	l, err := wi.List(ctx, &empty.Empty{})
	assert.NoError(t, err)
	if len(l.Invocations) != 1 || l.Invocations[0] != wiId {
		t.Errorf("Listed invocations '%v' did not match expected invocation '%s'", l.Invocations, wiId)
	}

	// Test invocation get, give some slack to actually invoke it
	var invocation *types.WorkflowInvocation
	deadline := time.Now().Add(time.Duration(1) * time.Second)
	tick := time.NewTicker(time.Duration(100) * time.Millisecond)
	for ti := range tick.C {
		invoc, err := wi.Get(ctx, &apiserver.WorkflowInvocationIdentifier{Id: wiId})
		assert.NoError(t, err)
		if invoc.Status.Status.Finished() || ti.After(deadline) {
			invocation = invoc
			tick.Stop()
			break
		}
	}

	assert.Equal(t, wiSpec, invocation.Spec)
	assert.Equal(t, etv, invocation.Status.Output)
	assert.True(t, invocation.Status.Status.Successful())
}

func TestDynamicWorkflowInvocation(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.Dial(GRPC_ADDRESS, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	cl := apiserver.NewWorkflowAPIClient(conn)
	wi := apiserver.NewWorkflowInvocationAPIClient(conn)

	// Test workflow creation
	wfSpec := &types.WorkflowSpec{
		ApiVersion: "v1",
		OutputTask: "fakeFinalTask",
		Tasks: map[string]*types.Task{
			"fakeFinalTask": {
				FunctionRef: "noop",
				Inputs: map[string]*types.TypedValue{
					types.INPUT_MAIN: typedvalues.Expr("{$.Tasks.someConditionalTask.Output}"),
				},
				Requires: map[string]*types.TaskDependencyParameters{
					"FirstTask":           {},
					"someConditionalTask": {},
				},
			},
			"FirstTask": {
				FunctionRef: "noop",
				Inputs: map[string]*types.TypedValue{
					types.INPUT_MAIN: typedvalues.Expr("{$.Invocation.Inputs.default.toUpperCase()}"),
				},
			},
			"someConditionalTask": {
				FunctionRef: "if",
				Inputs: map[string]*types.TypedValue{
					"condition": typedvalues.Expr("{$.Invocation.Inputs.default == 'FOO'}"),
					"consequent": typedvalues.Flow(&types.Task{
						FunctionRef: "noop",
						Inputs: map[string]*types.TypedValue{
							types.INPUT_MAIN: typedvalues.Expr("'consequent'"),
						},
					}),
					"alternative": typedvalues.Flow(&types.Task{
						FunctionRef: "noop",
						Inputs: map[string]*types.TypedValue{
							types.INPUT_MAIN: typedvalues.Expr("'alternative'"),
						},
					}),
				},
				Requires: map[string]*types.TaskDependencyParameters{
					"FirstTask": {},
				},
			},
		},
	}
	wfResp, err := cl.Create(ctx, wfSpec)
	assert.NoError(t, err)
	assert.NotNil(t, wfResp)
	assert.NotEmpty(t, wfResp.Id)

	wiSpec := &types.WorkflowInvocationSpec{
		WorkflowId: wfResp.Id,
		Inputs: map[string]*types.TypedValue{
			types.INPUT_MAIN: typedvalues.Expr("'foo'"),
		},
	}
	result, err := wi.InvokeSync(ctx, wiSpec)
	assert.NoError(t, err)

	typedvalues.Format(result.Status.Output)
}

func setup(ctx context.Context) {

	natsOptions := setupEventStore(ctx)

	go Run(ctx, &Options{
		// No fission for now
		InternalRuntime:       true,
		Controller:            true,
		ApiHttp:               true,
		ApiWorkflowInvocation: true,
		ApiWorkflow:           true,
		ApiAdmin:              true,
		Nats:                  natsOptions,
	})
}

func setupEventStore(ctx context.Context) *NatsOptions {
	clusterId := fmt.Sprintf("fission-workflows-tests-%d", time.Now().UnixNano())
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
	esOpts := &NatsOptions{
		Cluster: clusterId,
		Client:  "someClient",
		Url:     fmt.Sprintf("nats://%s:%d", address, port),
	}

	logrus.WithField("config", esOpts).Info("Setting up NATS server")

	// wait for a bit to set it up
	awaitCtx, cancel := context.WithTimeout(ctx, time.Duration(10)*time.Second)
	defer cancel()
	err = waitForNats(awaitCtx, esOpts.Url, esOpts.Cluster)
	if err != nil {
		logrus.Error(err)
	}
	logrus.WithField("config", esOpts).Info("NATS Server running")

	return esOpts
}

// Wait for NATS to come online, ignoring ErrNoServer as it could mean that NATS is still being setup
func waitForNats(ctx context.Context, url string, cluster string) error {
	conn, err := stan.Connect(cluster, "setupEventStore-alive-test", stan.NatsURL(url),
		stan.ConnectWait(time.Duration(10)*time.Second))
	if err == nats.ErrNoServers {
		logrus.WithFields(logrus.Fields{
			"cluster": cluster,
			"url":     url,
		}).Warnf("retrying due to err: %v", err)
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
