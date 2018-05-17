// Bundle package contains integration tests that run using the bundle
package bundle

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/fnenv/native/builtin"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/test/integration"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

const (
	gRPCAddress = ":5555"
)

func TestMain(m *testing.M) {
	if testing.Short() {
		fmt.Println("Skipping bundle tests...")
		return
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), time.Duration(1)*time.Minute)
	integration.SetupBundle(ctx)

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
	conn, err := grpc.Dial(gRPCAddress, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	cl := apiserver.NewWorkflowAPIClient(conn)

	// Test workflow creation
	spec := &types.WorkflowSpec{
		ApiVersion: types.WorkflowApiVersion,
		OutputTask: "fakeFinalTask",
		Tasks: map[string]*types.TaskSpec{
			"fakeFinalTask": {
				FunctionRef: "noop",
			},
		},
	}
	wfId, err := cl.Create(ctx, spec)
	defer cl.Delete(ctx, wfId)
	assert.NoError(t, err)
	assert.NotNil(t, wfId)
	assert.NotEmpty(t, wfId.GetId())

	time.Sleep(time.Duration(2) * time.Second)
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
	conn, err := grpc.Dial(gRPCAddress, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	cl := apiserver.NewWorkflowAPIClient(conn)
	wi := apiserver.NewWorkflowInvocationAPIClient(conn)

	// Test workflow creation
	wfSpec := &types.WorkflowSpec{
		ApiVersion: types.WorkflowApiVersion,
		OutputTask: "fakeFinalTask",
		Tasks: map[string]*types.TaskSpec{
			"fakeFinalTask": {
				FunctionRef: "noop",
				Inputs: map[string]*types.TypedValue{
					types.INPUT_MAIN: typedvalues.MustParse("{$.Tasks.FirstTask.Output}"),
				},
				Requires: map[string]*types.TaskDependencyParameters{
					"FirstTask": {},
				},
			},
			"FirstTask": {
				FunctionRef: "noop",
				Inputs: map[string]*types.TypedValue{
					types.INPUT_MAIN: typedvalues.MustParse("{$.Invocation.Inputs.default.toUpperCase()}"),
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
	deadline := time.Now().Add(time.Duration(10) * time.Second)
	tick := time.NewTicker(time.Duration(100) * time.Millisecond)
	for ti := range tick.C {
		invoc, err := wi.Get(ctx, &apiserver.WorkflowInvocationIdentifier{Id: wiId})
		assert.NoError(t, err)
		if invoc.Status.Finished() || ti.After(deadline) {
			invocation = invoc
			tick.Stop()
			break
		}
	}
	assert.Equal(t, wiSpec, invocation.Spec)
	assert.Equal(t, etv.Value, invocation.Status.Output.Value)
	assert.True(t, invocation.Status.Successful())
}

func TestDynamicWorkflowInvocation(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.Dial(gRPCAddress, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	cl := apiserver.NewWorkflowAPIClient(conn)
	wi := apiserver.NewWorkflowInvocationAPIClient(conn)

	// Test workflow creation
	wfSpec := &types.WorkflowSpec{
		ApiVersion: types.WorkflowApiVersion,
		OutputTask: "fakeFinalTask",
		Tasks: map[string]*types.TaskSpec{
			"fakeFinalTask": {
				FunctionRef: "noop",
				Inputs: map[string]*types.TypedValue{
					types.INPUT_MAIN: typedvalues.MustParse("{$.Tasks.someConditionalTask.Output}"),
				},
				Requires: map[string]*types.TaskDependencyParameters{
					"FirstTask":           {},
					"someConditionalTask": {},
				},
			},
			"FirstTask": {
				FunctionRef: "noop",
				Inputs: map[string]*types.TypedValue{
					types.INPUT_MAIN: typedvalues.MustParse("{$.Invocation.Inputs.default.toUpperCase()}"),
				},
			},
			"someConditionalTask": {
				FunctionRef: "if",
				Inputs: map[string]*types.TypedValue{
					builtin.IfInputCondition: typedvalues.MustParse("{$.Invocation.Inputs.default == 'FOO'}"),
					builtin.IfInputThen: typedvalues.ParseTask(&types.TaskSpec{
						FunctionRef: "noop",
						Inputs: map[string]*types.TypedValue{
							types.INPUT_MAIN: typedvalues.MustParse("{'consequent: ' + $.Tasks.FirstTask.Output}"),
						},
					}),
					builtin.IfInputElse: typedvalues.ParseTask(&types.TaskSpec{
						FunctionRef: "noop",
						Inputs: map[string]*types.TypedValue{
							types.INPUT_MAIN: typedvalues.MustParse("{'alternative: ' + $.Tasks.FirstTask.Output}"),
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
	defer cl.Delete(ctx, wfResp)
	assert.NoError(t, err)
	assert.NotNil(t, wfResp)
	assert.NotEmpty(t, wfResp.Id)

	wiSpec := &types.WorkflowInvocationSpec{
		WorkflowId: wfResp.Id,
		Inputs: map[string]*types.TypedValue{
			types.INPUT_MAIN: typedvalues.MustParse("foo"),
		},
	}
	wfi, err := wi.InvokeSync(ctx, wiSpec)
	assert.NoError(t, err)
	assert.NotEmpty(t, wfi.Status.DynamicTasks)
	assert.True(t, wfi.Status.Finished())
	assert.True(t, wfi.Status.Successful())
	assert.Equal(t, 4, len(wfi.Status.Tasks))

	output := typedvalues.MustFormat(wfi.Status.Output)
	assert.Equal(t, "alternative: FOO", output)
}

func TestInlineWorkflowInvocation(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.Dial(gRPCAddress, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	cl := apiserver.NewWorkflowAPIClient(conn)
	wi := apiserver.NewWorkflowInvocationAPIClient(conn)

	// Test workflow creation
	wfSpec := &types.WorkflowSpec{
		ApiVersion: types.WorkflowApiVersion,
		OutputTask: "finalTask",
		Tasks: map[string]*types.TaskSpec{
			"nestedTask": {
				FunctionRef: "noop",
				Inputs: map[string]*types.TypedValue{
					builtin.NoopInput: typedvalues.ParseWorkflow(&types.WorkflowSpec{
						OutputTask: "b",
						Tasks: map[string]*types.TaskSpec{
							"a": {
								FunctionRef: "noop",
								Inputs: map[string]*types.TypedValue{
									types.INPUT_MAIN: typedvalues.MustParse("inner1"),
								},
							},
							"b": {
								FunctionRef: "noop",
								Inputs: map[string]*types.TypedValue{
									types.INPUT_MAIN: typedvalues.MustParse("{output('a')}"),
								},
								Requires: map[string]*types.TaskDependencyParameters{
									"a": nil,
								},
							},
						},
					}),
				},
			},
			"finalTask": {
				FunctionRef: "noop",
				Inputs: map[string]*types.TypedValue{
					types.INPUT_MAIN: typedvalues.MustParse("output('nestedTask')"),
				},
				Requires: map[string]*types.TaskDependencyParameters{
					"nestedTask": {},
				},
			},
		},
	}
	wfResp, err := cl.Create(ctx, wfSpec)
	defer cl.Delete(ctx, wfResp)
	assert.NoError(t, err)
	assert.NotNil(t, wfResp)
	assert.NotEmpty(t, wfResp.Id)

	wiSpec := &types.WorkflowInvocationSpec{
		WorkflowId: wfResp.Id,
	}
	wfi, err := wi.InvokeSync(ctx, wiSpec)
	assert.NoError(t, err)
	assert.NotEmpty(t, wfi.Status.DynamicTasks)
	assert.True(t, wfi.Status.Finished())
	assert.True(t, wfi.Status.Successful())
	assert.Equal(t, 3, len(wfi.Status.Tasks))

	_, err = typedvalues.Format(wfi.Status.Output)
	assert.NoError(t, err)
}

func TestLongRunningWorkflowInvocation(t *testing.T) {
	ctx := context.Background()
	conn, err := grpc.Dial(gRPCAddress, grpc.WithInsecure())
	if err != nil {
		panic(err)
	}
	cl := apiserver.NewWorkflowAPIClient(conn)
	wi := apiserver.NewWorkflowInvocationAPIClient(conn)

	// Test workflow creation
	wfSpec := &types.WorkflowSpec{
		ApiVersion: types.WorkflowApiVersion,
		OutputTask: "final",
		Tasks: types.Tasks{
			"longSleep": {
				FunctionRef: builtin.Sleep,
				Inputs:      typedvalues.Input("5s"),
			},
			"afterSleep": {
				FunctionRef: builtin.Noop,
				Inputs:      typedvalues.Input("{ '4' }"),
				Requires:    types.Require("longSleep"),
			},
			"parallel1": {
				FunctionRef: builtin.Noop,
				Inputs:      typedvalues.Input("{ '1' }"),
				Requires:    types.Require("longSleep"),
			},
			"parallel2": {
				FunctionRef: builtin.Noop,
				Inputs:      typedvalues.Input("{ output('parallel1') + '2' }"),
				Requires:    types.Require("parallel1"),
			},
			"parallel3": {
				FunctionRef: builtin.Noop,
				Inputs:      typedvalues.Input("{ output('parallel2') + '3' }"),
				Requires:    types.Require("parallel2"),
			},
			"merge": {
				FunctionRef: builtin.Noop,
				Inputs:      typedvalues.Input("{ output('parallel3') + output('afterSleep') }"),
				Requires:    types.Require("parallel3", "afterSleep"),
			},
			"final": {
				FunctionRef: builtin.Noop,
				Inputs:      typedvalues.Input("{ output('merge') }"),
				Requires:    types.Require("merge"),
			},
		},
	}
	wfResp, err := cl.Create(ctx, wfSpec)
	defer cl.Delete(ctx, wfResp)
	assert.NoError(t, err, err)
	assert.NotNil(t, wfResp)
	assert.NotEmpty(t, wfResp.Id)

	wiSpec := types.NewWorkflowInvocationSpec(wfResp.Id)
	wfi, err := wi.InvokeSync(ctx, wiSpec)
	assert.NoError(t, err)
	assert.Empty(t, wfi.Status.DynamicTasks)
	assert.True(t, wfi.Status.Finished())
	assert.True(t, wfi.Status.Successful())
	assert.Equal(t, len(wfSpec.Tasks), len(wfi.Status.Tasks))

	output := typedvalues.MustFormat(wfi.Status.Output)
	assert.Equal(t, "1234", output)
}
