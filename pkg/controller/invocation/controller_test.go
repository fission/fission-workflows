package invocation

import (
	"context"
	"testing"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/api/store"
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/controller/expr"
	"github.com/fission/fission-workflows/pkg/fes/backend/mem"
	"github.com/fission/fission-workflows/pkg/fes/testutil"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/fnenv/mock"
	"github.com/fission/fission-workflows/pkg/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestController_Lifecycle(t *testing.T) {
	cache := testutil.NewCache()
	s := scheduler.NewInvocationScheduler(scheduler.DefaultPolicy)
	es := mem.NewBackend()
	mockRuntime := mock.NewRuntime()
	mockResolver := fnenv.NewMetaResolver(map[string]fnenv.RuntimeResolver{
		"mock": mock.NewResolver(),
	})

	wfiAPI := api.NewInvocationAPI(es)
	wfAPI := api.NewWorkflowAPI(es, mockResolver)
	dynamicAPI := api.NewDynamicApi(wfAPI, wfiAPI)
	taskAPI := api.NewTaskAPI(map[string]fnenv.Runtime{
		"mock": mockRuntime,
	}, es, dynamicAPI)

	ctr := NewController(store.NewInvocationStore(cache), store.NewWorkflowsStore(cache), s, taskAPI, wfiAPI,
		expr.NewStore(), controller.NewLocalExecutor(1, 100))

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
