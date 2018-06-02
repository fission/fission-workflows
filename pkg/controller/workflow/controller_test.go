package workflow

import (
	"context"
	"testing"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fes/backend/mem"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/fnenv/mock"
	"github.com/stretchr/testify/assert"
)

func TestController_Lifecycle(t *testing.T) {
	cache := fes.NewMapCache()
	es := mem.NewBackend()
	mockResolver := fnenv.NewMetaResolver(map[string]fnenv.RuntimeResolver{
		"mock": mock.NewResolver(),
	})
	wfAPI := api.NewWorkflowAPI(es, mockResolver)

	ctr := NewController(cache, wfAPI)

	err := ctr.Init(context.TODO())
	assert.NoError(t, err)

	err = ctr.Close()
	assert.NoError(t, err)
}

/*
TODO test informer
TODO test ticks
TODO test concurrency
TODO test evaluation
- completed
- scheduler
- errored
TODO test individual rules
*/
