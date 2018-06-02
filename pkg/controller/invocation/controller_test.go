package invocation

import (
	"context"
	"testing"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/controller/expr"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fes/backend/mem"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/fnenv/mock"
	"github.com/fission/fission-workflows/pkg/scheduler"
	"github.com/stretchr/testify/assert"
)

func TestController_Lifecycle(t *testing.T) {
	cache := fes.NewMapCache()
	s := &scheduler.WorkflowScheduler{}
	es := mem.NewBackend()
	mockRuntime := mock.NewRuntime()
	mockResolver := fnenv.NewMetaResolver(map[string]fnenv.RuntimeResolver{
		"mock": mock.NewResolver(),
	})

	wfiApi := api.NewInvocation(es)
	wfApi := api.NewWorkflow(es, mockResolver)
	dynamicApi := api.NewDynamic(wfApi, wfiApi)
	taskApi := api.NewTaskApi(map[string]fnenv.Runtime{
		"mock": mockRuntime,
	}, es, dynamicApi)

	ctr := NewController(cache, cache, s, taskApi, wfiApi, expr.NewStore())

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
