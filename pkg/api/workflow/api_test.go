package workflow

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/fes/backend/mem"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/fnenv/mock"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/stretchr/testify/assert"
)

func TestApi_Create(t *testing.T) {
	api, es := setupApi()

	spec := types.NewWorkflowSpec().
		SetOutput("task1").
		AddTask("task1", types.NewTaskSpec("someFn"))
	id1, err := api.Create(spec)
	assert.NoError(t, err)
	id2, err := api.Create(spec)
	assert.NoError(t, err)
	assert.NotEqual(t, id1, id2)

	s := es.Snapshot()
	assert.Len(t, s, 2)
}

func TestApi_CreateInvalid(t *testing.T) {
	api, es := setupApi()

	spec := types.NewWorkflowSpec()
	id, err := api.Create(spec)
	assert.IsType(t, validate.Error{}, err)
	assert.Empty(t, id)

	s := es.Snapshot()
	assert.Len(t, s, 0)
}

func TestApi_CreateForceId(t *testing.T) {
	api, es := setupApi()

	spec := types.NewWorkflowSpec().
		SetOutput("task1").
		AddTask("task1", types.NewTaskSpec("someFn"))

	spec.ForceId = "forcedId"

	id, err := api.Create(spec)
	assert.NoError(t, err)
	assert.Equal(t, spec.ForceId, id)

	s := es.Snapshot()
	assert.Len(t, s, 1)
}

func TestApi_CreateDuplicate(t *testing.T) {
	api, es := setupApi()

	spec := types.NewWorkflowSpec().
		SetOutput("task1").
		AddTask("task1", types.NewTaskSpec("someFn"))
	spec.ForceId = "forcedId"

	id, err := api.Create(spec)
	assert.NoError(t, err)
	assert.Equal(t, spec.ForceId, id)

	id2, err := api.Create(spec)
	assert.Error(t, ErrWorkflowAlreadyExists, err)
	assert.Equal(t, id, id2)

	s := es.Snapshot()
	assert.Len(t, s, 1)
}

func TestApi_Delete(t *testing.T) {
	api, es := setupApi()

	spec := types.NewWorkflowSpec().
		SetOutput("task1").
		AddTask("task1", types.NewTaskSpec("someFn"))
	spec.ForceId = "forcedId"

	id, err := api.Create(spec)
	assert.NoError(t, err)
	assert.Equal(t, spec.ForceId, id)

	err = api.Delete(id)
	assert.NoError(t, err)

	s := es.Snapshot()
	assert.Len(t, s, 1)
	for k, v := range s {
		assert.Equal(t, id, k.Id)
		assert.Len(t, v, 2)
		break
	}
}

func TestApi_DeleteNonExistent(t *testing.T) {
	api, es := setupApi()
	id := "non-existent"
	err := api.Delete(id)
	assert.NoError(t, err)

	s := es.Snapshot()
	assert.Len(t, s, 1)
}

func TestApi_Parse(t *testing.T) {
	api, _ := setupApi()

	spec := types.NewWorkflowSpec().
		SetOutput("task1").
		AddTask("task1", types.NewTaskSpec("someFn"))
	spec.ForceId = "forcedId"

	id, err := api.Create(spec)
	assert.NoError(t, err)
	assert.Equal(t, spec.ForceId, id)

}

func setupApi() (*Api, *mem.Backend) {
	es := mem.NewBackend()
	resolver := fnenv.NewMetaResolver(map[string]fnenv.RuntimeResolver{
		"mock": mock.NewResolver(),
	})
	return NewApi(es, resolver), es
}
