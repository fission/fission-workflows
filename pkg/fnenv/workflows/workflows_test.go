package workflows

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/api/aggregates"
	"github.com/fission/fission-workflows/pkg/api/store"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fes/backend/mem"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/stretchr/testify/assert"
)

func TestRuntime_InvokeWorkflow_SubTimeout(t *testing.T) {
	runtime, _, _, _ := setup()
	runtime.timeout = 10 * time.Millisecond

	_, err := runtime.InvokeWorkflow(types.NewWorkflowInvocationSpec("123"))
	assert.Equal(t, api.ErrInvocationCanceled, err.Error())
}

func TestRuntime_InvokeWorkflow_PollTimeout(t *testing.T) {
	runtime, _, _, _ := setup()
	runtime.store = store.NewInvocationStore(fes.NewMapCache()) // ensure that cache does not support pubsub
	runtime.timeout = 10 * time.Millisecond
	runtime.pollInterval = 10 * time.Millisecond

	_, err := runtime.InvokeWorkflow(types.NewWorkflowInvocationSpec("123"))
	assert.EqualError(t, err, context.DeadlineExceeded.Error())
}

func TestRuntime_InvokeWorkflow_InvalidSpec(t *testing.T) {
	runtime, _, _, _ := setup()
	_, err := runtime.InvokeWorkflow(types.NewWorkflowInvocationSpec(""))
	assert.IsType(t, validate.Error{}, err)
}

func TestRuntime_InvokeWorkflow_SubSuccess(t *testing.T) {
	runtime, invocationAPI, _, cache := setup()
	output := typedvalues.MustParse("foo")
	go func() {
		// Simulate workflow invocation
		time.Sleep(50 * time.Millisecond)
		entities := cache.List()
		wfiID := entities[0].Id
		err := invocationAPI.Complete(wfiID, output)
		if err != nil {
			panic(err)
		}
	}()
	wfi, err := runtime.InvokeWorkflow(types.NewWorkflowInvocationSpec("123"))
	assert.NoError(t, err)
	assert.Equal(t, output, wfi.GetStatus().GetOutput())
	assert.True(t, wfi.GetStatus().Finished())
	assert.True(t, wfi.GetStatus().Successful())
}

func TestRuntime_InvokeWorkflow_PollSuccess(t *testing.T) {
	runtime, invocationAPI, _, cache := setup()
	pollCache := store.NewInvocationStore(fes.NewMapCache()) // ensure that cache does not support pubsub
	runtime.store = pollCache

	output := typedvalues.MustParse("foo")
	go func() {
		// Simulate workflow invocation
		time.Sleep(50 * time.Millisecond)
		entities := cache.List()
		wfiID := entities[0].Id
		err := invocationAPI.Complete(wfiID, output)
		if err != nil {
			panic(err)
		}
		time.Sleep(50 * time.Millisecond)
		entity, err := cache.GetAggregate(fes.NewAggregate(aggregates.TypeWorkflowInvocation, wfiID))
		assert.NoError(t, err)
		pollCache.CacheReader.(fes.CacheReaderWriter).Put(entity) // A bit hacky..
	}()
	wfi, err := runtime.InvokeWorkflow(types.NewWorkflowInvocationSpec("123"))
	assert.NoError(t, err)
	assert.Equal(t, output, wfi.GetStatus().GetOutput())
	assert.True(t, wfi.GetStatus().Finished())
	assert.True(t, wfi.GetStatus().Successful())
}

func TestRuntime_InvokeWorkflow_Fail(t *testing.T) {
	runtime, invocationAPI, _, cache := setup()
	wfiErr := errors.New("stub err")
	go func() {
		// Simulate workflow invocation
		time.Sleep(50 * time.Millisecond)
		entities := cache.List()
		wfiID := entities[0].Id
		err := invocationAPI.Fail(wfiID, wfiErr)
		if err != nil {
			panic(err)
		}
	}()
	wfi, err := runtime.InvokeWorkflow(types.NewWorkflowInvocationSpec("123"))
	assert.NoError(t, err)
	assert.Equal(t, wfiErr.Error(), wfi.GetStatus().GetError().Error())
	assert.True(t, wfi.GetStatus().Finished())
	assert.False(t, wfi.GetStatus().Successful())
}

func TestRuntime_InvokeWorkflow_Cancel(t *testing.T) {
	runtime, invocationAPI, _, cache := setup()
	go func() {
		// Simulate workflow invocation
		time.Sleep(50 * time.Millisecond)
		entities := cache.List()
		wfiID := entities[0].Id
		err := invocationAPI.Cancel(wfiID)
		if err != nil {
			panic(err)
		}
	}()
	wfi, err := runtime.InvokeWorkflow(types.NewWorkflowInvocationSpec("123"))
	assert.NoError(t, err)
	assert.Equal(t, api.ErrInvocationCanceled, wfi.GetStatus().GetError().Error())
	assert.True(t, wfi.GetStatus().Finished())
	assert.False(t, wfi.GetStatus().Successful())
}

func TestRuntime_Invoke(t *testing.T) {
	runtime, invocationAPI, _, cache := setup()

	spec := types.NewTaskInvocationSpec("wi-123", "ti-123", types.NewFnRef("internal", "", "fooFn"))
	spec.Inputs = types.Inputs{}
	spec.Inputs[types.InputParent] = typedvalues.MustParse("parentID")
	output := typedvalues.MustParse("foo")
	go func() {
		// Simulate workflow invocation
		time.Sleep(50 * time.Millisecond)
		entities := cache.List()
		wfiID := entities[0].Id
		err := invocationAPI.Complete(wfiID, output)
		if err != nil {
			panic(err)
		}
	}()

	task, err := runtime.Invoke(spec)
	assert.NoError(t, err)
	assert.Equal(t, output, task.GetOutput())
}

func setup() (*Runtime, *api.Invocation, *mem.Backend, fes.CacheReaderWriter) {
	backend := mem.NewBackend()
	invocationAPI := api.NewInvocationAPI(backend)
	cache := fes.NewSubscribedCache(context.Background(), fes.NewMapCache(), func() fes.Entity {
		return aggregates.NewWorkflowInvocation("")
	}, backend.Subscribe())
	runtime := NewRuntime(invocationAPI, store.NewInvocationStore(cache))
	runtime.timeout = 5 * time.Second
	return runtime, invocationAPI, backend, cache
}
