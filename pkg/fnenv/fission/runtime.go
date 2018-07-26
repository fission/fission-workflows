package fission

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/types/typedvalues/httpconv"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/sirupsen/logrus"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"

	executor "github.com/fission/fission/executor/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	Name = "fission"
)

var log = logrus.WithField("component", "fnenv.fission")

// FunctionEnv adapts the Fission platform to the function execution runtime. This allows the workflow engine
// to invoke Fission functions.
type FunctionEnv struct {
	executor         *executor.Client
	routerURL        string
	timedExecService *timedExecPool
}

const (
	defaultHTTPMethod = http.MethodPost
	defaultProtocol   = "http"
	provisionDuration = time.Duration(500) - time.Millisecond
)

func NewFunctionEnv(executor *executor.Client, routerURL string) *FunctionEnv {
	return &FunctionEnv{
		executor:         executor,
		routerURL:        routerURL,
		timedExecService: newTimedExecPool(),
	}
}

// Invoke executes the task in a blocking way.
//
// spec contains the complete configuration needed for the execution.
// It returns the TaskInvocationStatus with a completed (FINISHED, FAILED, ABORTED) status.
// An error is returned only when error occurs outside of the runtime's control.
func (fe *FunctionEnv) Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocationStatus, error) {
	ctxLog := log.WithField("fn", spec.FnRef)
	if err := validate.TaskInvocationSpec(spec); err != nil {
		return nil, err
	}
	fnRef := *spec.FnRef

	// Construct request and add body
	url := fe.createRouterURL(fnRef)
	req, err := http.NewRequest(defaultHTTPMethod, url, nil)
	if err != nil {
		panic(fmt.Errorf("failed to create request for '%v': %v", url, err))
	}
	// Map task inputs to request
	err = httpconv.FormatRequest(spec.Inputs, req)
	if err != nil {
		return nil, err
	}

	// Perform request
	timeStart := time.Now()
	fnenv.FnActive.WithLabelValues(Name).Inc()
	defer fnenv.FnExecTime.WithLabelValues(Name).Observe(float64(time.Since(timeStart)))
	ctxLog.Infof("Invoking Fission function: '%v'.", req.URL)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error for reqUrl '%v': %v", url, err)
	}
	fnenv.FnActive.WithLabelValues(Name).Dec()
	fnenv.FnActive.WithLabelValues(Name).Inc()

	// Parse output
	output, err := httpconv.ParseBody(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse output: %v", err)
	}

	ctxLog.Infof("[%s][status]: %v", fnRef.ID, resp.StatusCode)
	ctxLog.Infof("[%s][Content-Type]: %v ", fnRef.ID, resp.Header.Get("Content-Type"))
	ctxLog.Infof("[%s][output]: '%s'", fnRef.ID, typedvalues.MustFormat(&output))

	// Determine status of the task invocation
	if resp.StatusCode >= 400 {
		msg, _ := typedvalues.Format(&output)
		ctxLog.Warnf("[%s] Failed %v: %v", resp.StatusCode, msg)
		return &types.TaskInvocationStatus{
			Status: types.TaskInvocationStatus_FAILED,
			Error: &types.Error{
				Message: fmt.Sprintf("fission function error: %v", msg),
			},
		}, nil
	}

	return &types.TaskInvocationStatus{
		Status: types.TaskInvocationStatus_SUCCEEDED,
		Output: &output,
	}, nil
}

// Notify signals the Fission runtime that a function request is expected at a specific time.
func (fe *FunctionEnv) Notify(fn types.FnRef, expectedAt time.Time) error {
	reqURL, err := fe.getFnURL(fn)
	if err != nil {
		return err
	}

	// For this now assume a standard cold start delay; use profiling to provide a better estimate.
	execAt := expectedAt.Add(-provisionDuration)

	// Tap the Fission function at the right time
	fe.timedExecService.Submit(func() {
		log.WithField("fn", fn).Infof("Tapping Fission function: %v", reqURL)
		fe.executor.TapService(reqURL)
	}, execAt)
	return nil
}

func (fe *FunctionEnv) getFnURL(fn types.FnRef) (*url.URL, error) {
	meta := createFunctionMeta(fn)
	serviceURL, err := fe.executor.GetServiceForFunction(meta)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err":  err,
			"meta": meta,
		}).Error("Fission function could not be found!")
		return nil, err
	}
	rawURL := fmt.Sprintf("%s://%s", defaultProtocol, serviceURL)
	reqURL, err := url.Parse(rawURL)
	if err != nil {
		logrus.Errorf("Failed to parse url: '%v'", rawURL)
		panic(err)
	}
	return reqURL, nil
}

func createFunctionMeta(fn types.FnRef) *metav1.ObjectMeta {

	return &metav1.ObjectMeta{
		Name:      fn.ID,
		Namespace: metav1.NamespaceDefault,
	}
}

func (fe *FunctionEnv) createRouterURL(fn types.FnRef) string {
	return fe.routerURL + "/fission-function/" + fn.ID
}
