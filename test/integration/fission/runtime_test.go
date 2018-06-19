package fission

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/fission/fission-workflows/pkg/fnenv/fission"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	controllerclient "github.com/fission/fission/controller/client"
	executorclient "github.com/fission/fission/executor/client"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

// Future: discover executor of fission deployment
const (
	executorLocalPort   = 9031
	routerLocalPort     = 9032
	controllerLocalPort = 9033
)

var executor = executorclient.MakeClient(localhost(executorLocalPort))
var controller = controllerclient.MakeClient(localhost(controllerLocalPort))
var testFnName = "fission-runtime-test"

// Currently we assume that fission is present (along with the CLI) and kubectl.
func TestMain(m *testing.M) {
	var status int
	if testing.Short() {
		log.Info("Short test; skipping Fission integration tests")
		return
	}

	ctx, cancelFn := context.WithTimeout(context.Background(), 10*time.Second)
	defer func() {
		log.Info("Awaiting completion of goroutines...")
		cancelFn()
		time.Sleep(5 * time.Second)
		os.Exit(status)
	}()

	// Test if Fission is present
	if err := exec.CommandContext(ctx, "fission", "fn", "list").Run(); err != nil {
		log.Panicf("Fission is not present: %v", err)
	}

	err := portForwardComponents(ctx, "fission")
	if err != nil {
		log.Panicf("Failed to forward components: %v", err)
	}
	time.Sleep(time.Minute)

	// Test if cluster is accessible
	if err := exec.CommandContext(ctx, "curl", localhost(routerLocalPort)).Run(); err != nil {
		log.Panicf("Fission cluster (%v) is not accessible: %v", localhost(controllerLocalPort), err)
	}

	// Setup test environment and function (use fission CLI because client is not exposed correctly)
	if err := exec.Command("fission", "spec", "apply").Run(); err != nil {
		log.Panicf("Failed to create test resources in fission: %v", err)
	}
	log.Info("Fission integration test resources setup.")

	status = m.Run()

	// Clean up test function and env
	cancelFn()
	if err := exec.Command("fission", "spec", "destroy").Run(); err != nil {
		log.Fatalf("Failed to destroy test resources in fission: %v", err)
	}
	log.Info("Cleaned up fission test resources.")
}

func TestFnenvResolve(t *testing.T) {
	resolver := fission.NewResolver(controller)
	resolved, err := resolver.Resolve(testFnName)
	assert.NoError(t, err)
	assert.Equal(t, testFnName, resolved)
}

func TestFnenvNotify(t *testing.T) {
	fnref := types.NewFnRef(fission.Name, testFnName)
	fnenv := fission.NewFunctionEnv(executor, localhost(routerLocalPort))
	err := fnenv.Notify(fnref, time.Now().Add(100*time.Millisecond))
	assert.NoError(t, err)
}

func TestFnenvInvoke(t *testing.T) {
	fnref := types.NewFnRef(fission.Name, testFnName)
	fnenv := fission.NewFunctionEnv(executor, localhost(routerLocalPort))
	body := "stubBodyVal"
	headerVal := "stub-header-val"
	headerKey := "stub-header-key"

	result, err := fnenv.Invoke(&types.TaskInvocationSpec{
		TaskId:       "fooTask",
		InvocationId: "fooInvocation",
		Inputs: types.Inputs{
			"default": typedvalues.MustParse(body),
			"headers": typedvalues.MustParse(map[string]interface{}{
				headerKey: headerVal,
			}),
		},
		FnRef: &fnref,
	})
	output := typedvalues.MustFormat(result.Output)
	assert.NoError(t, err)
	assert.True(t, result.Finished())
	assert.NotEmpty(t, output)
	assert.Contains(t, output, body)
	assert.Contains(t, output, headerVal)
	assert.Contains(t, output, headerKey)
}

func portForwardComponents(ctx context.Context, ns string) error {
	buf := bytes.NewBuffer(nil)
	cmd := exec.CommandContext(ctx, "kubectl", "-n", ns, "get", "pods", "-o", "name")
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	pods := strings.Split(buf.String(), "\n")
	for _, pod := range pods {
		if kindSplit := strings.Index(pod, "/"); kindSplit >= 0 {
			pod = pod[kindSplit+1:]
		}

		var portMap string
		var err error
		switch {
		case strings.HasPrefix(pod, "controller"):
			portMap = fmt.Sprintf("%d:80", controllerLocalPort)
		case strings.HasPrefix(pod, "executor"):
			portMap = fmt.Sprintf("%d:80", executorLocalPort)
		case strings.HasPrefix(pod, "router"):
			portMap = fmt.Sprintf("%d:80", routerLocalPort)
		}
		if len(portMap) != 0 {
			err = portForward(ctx, ns, pod, portMap)
			fmt.Printf("Setup proxy from %v to %v\n", portMap, pod)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func portForward(ctx context.Context, ns string, pod string, portMap string) error {
	cmd := exec.CommandContext(ctx, "kubectl", "-n", ns, "port-forward", pod, portMap)
	log.Info("exec: ", strings.Join(cmd.Args, " "))
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	err := cmd.Start()
	if err != nil {
		return err
	}
	go func() {
		state, err := cmd.Process.Wait()
		if err != nil {
			log.Error(err)
		}
		log.Infof("portForward: %v.%v - %v", pod, ns, state)

	}()
	return nil
}

func localhost(port int) string {
	return fmt.Sprintf("http://localhost:%d", port)
}
