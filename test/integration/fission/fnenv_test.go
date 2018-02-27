package fission

import (
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/fnenv/fission"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	defaultFissionControllerAddr = "http://controller.fission"
	defaultFissionExecutorAddr   = "http://executor.fission"
)

var fissionRuntime *fission.FunctionEnv
var fissionResolver *fission.Resolver
var resolver fnenv.Resolver

func TestMain(m *testing.M) {
	fissionControllerAddr := os.Getenv("FNENV_FISSION_CONTROLLER")
	if len(fissionControllerAddr) == 0 {
		fissionControllerAddr = defaultFissionControllerAddr
	}

	fissionExecutorAddr := os.Getenv("FNENV_FISSION_Executor")
	if len(fissionExecutorAddr) == 0 {
		fissionExecutorAddr = defaultFissionExecutorAddr
	}

	var fissionMissing bool
	_, err := http.Get(fissionControllerAddr)
	if err != nil {
		logrus.Warnf("Fission Controller not available: %v", err)
		fissionMissing = true
	}
	_, err = http.Get(fissionExecutorAddr)
	if err != nil {
		logrus.Warnf("Fission Executor not available: %v", err)
		fissionMissing = true
	}
	if fissionMissing {
		fmt.Println("Fission not available; skipping Fission integration test")
		return
	}

	// Setup Fission connection
	fissionRuntime = fission.SetupRuntime(fissionExecutorAddr)
	fissionResolver = fission.SetupResolver(fissionControllerAddr)
	resolver = fnenv.NewMetaResolver(map[string]fnenv.RuntimeResolver{
		"fission": fissionResolver,
	})

	m.Run()
}

func TestFissionResolveFunction(t *testing.T) {
	// TODO add actual function
	fnId, err := fissionResolver.Resolve("hello")
	assert.NoError(t, err)
	assert.NotEmpty(t, fnId)
}

func TestFissionExecuteFunction(t *testing.T) {
	// TODO add actual function
	fnRef, err := resolver.Resolve("hello")
	assert.NoError(t, err)
	assert.NotEmpty(t, fnRef)

	taskSpec := types.NewTaskInvocationSpec("wfi-1", "sometask", fnRef)
	status, err := fissionRuntime.Invoke(taskSpec)
	assert.NoError(t, err)
	assert.True(t, status.Finished())
	assert.Nil(t, status.Error)
	assert.NotNil(t, status.Output)
}

func TestFissionNotify(t *testing.T) {
	// TODO add actual function
	fnRef, err := resolver.Resolve("hello")
	assert.NoError(t, err)
	assert.NotEmpty(t, fnRef)

	err = fissionRuntime.Notify("someTask", fnRef, time.Now())
	assert.NoError(t, err)
}
