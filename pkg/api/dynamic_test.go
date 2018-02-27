package api

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/fes/backend/mem"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/fnenv/mock"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestApi_AddDynamicTask(t *testing.T) {
	api, es := setupDynamicApi()

	wfSpec := types.NewWorkflowSpec().
		SetOutput("task1").
		AddTask("task1", types.NewTaskSpec("someFn"))
	wfSpec.ForceId = "forcedId"
	id, err := api.wfAPI.Create(wfSpec)
	assert.NoError(t, err)

	wfiSpec := types.NewWorkflowInvocationSpec(id)
	invocationId, err := api.wfiAPI.Invoke(wfiSpec)
	assert.NoError(t, err)

	taskSpec := types.NewTaskSpec("someFn")
	err = api.AddDynamicFlow(invocationId, "task-parent", *typedvalues.FlowTask(taskSpec))
	assert.NoError(t, err)

	s := es.Snapshot()
	assert.Len(t, s, 3)
}

func setupDynamicApi() (*Dynamic, *mem.Backend) {
	es := mem.NewBackend()
	resolver := fnenv.NewMetaResolver(map[string]fnenv.RuntimeResolver{
		"mock": mock.NewResolver(),
	})

	wfiApi := NewInvocationAPI(es)
	wfApi := NewWorkflowAPI(es, resolver)

	return NewDynamicAPI(wfApi, wfiApi), es
}
