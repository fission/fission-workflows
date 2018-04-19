package fission

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/fission/fission-workflows/pkg/fnenv/common/httpconv"
	"github.com/sirupsen/logrus"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"

	executor "github.com/fission/fission/executor/client"
	"github.com/fission/fission/router"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

var log = logrus.WithField("component", "fnenv.fission")

// FunctionEnv adapts the Fission platform to the function execution runtime. This allows the workflow engine
// to invoke Fission functions.
type FunctionEnv struct {
	executor         *executor.Client
	timedExecService *TimedExecPool
}

const (
	defaultHttpMethod = http.MethodPost
	defaultProtocol   = "http"
	provisionDuration = time.Duration(500) - time.Millisecond
)

func NewFunctionEnv(executor *executor.Client) *FunctionEnv {
	return &FunctionEnv{
		executor: executor,
	}
}

// Invoke executes the task in a blocking way.
//
// spec contains the complete configuration needed for the execution.
// It returns the TaskInvocationStatus with a completed (FINISHED, FAILED, ABORTED) status.
// An error is returned only when error occurs outside of the runtime's control.
func (fe *FunctionEnv) Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocationStatus, error) {
	// Get fission function url
	// TODO use router instead once we can route to a specific function uid
	// TODO extract validation to validation package
	ctxLog := log.WithField("fn", spec.FnRef)
	if spec.FnRef == nil {
		return nil, errors.New("invocation does not contain FnRef")
	}
	fnRef := *spec.FnRef

	reqUrl, err := fe.getFnUrl(fnRef)

	// Construct request and add body
	ctxLog.Infof("Invoking Fission function: '%v'.", reqUrl)
	req, err := http.NewRequest(defaultHttpMethod, reqUrl.String(), nil)
	if err != nil {
		panic(fmt.Errorf("failed to make request for '%v': %v", reqUrl, err))
	}

	// Map task inputs to request
	formatRequest(req, spec.Inputs)

	// Add parameters normally added by Fission
	meta := createFunctionMeta(fnRef)
	router.MetadataToHeaders(router.HEADERS_FISSION_FUNCTION_PREFIX, meta, req)

	// Perform request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error for reqUrl '%v': %v", reqUrl, err)
	}

	// Parse output
	output, err := httpconv.ParseBody(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse output: %v", err)
	}

	ctxLog.Infof("[response][status]: %v", meta.Name, resp.StatusCode)
	ctxLog.Infof("[response][Content-Type]: %v ", meta.Name, resp.Header.Get("Content-Type"))
	ctxLog.Debugf("[%s][output]: %v", meta.Name, output)

	// Determine status of the task invocation
	if resp.StatusCode >= 400 {
		msg, _ := typedvalues.Format(&output)
		ctxLog.Warn("[%s] Failed %v: %v", resp.StatusCode, msg)
		return &types.TaskInvocationStatus{
			Status: types.TaskInvocationStatus_FAILED,
			Error: &types.Error{
				Code:    fmt.Sprintf("%v", resp.StatusCode),
				Message: fmt.Sprintf("%v", msg),
			},
		}, nil
	}

	return &types.TaskInvocationStatus{
		Status: types.TaskInvocationStatus_SUCCEEDED,
		Output: &output,
	}, nil
}

// Notify signals the Fission runtime that a function request is expected at a specific time.
func (fe *FunctionEnv) Notify(taskID string, fn types.FnRef, expectedAt time.Time) error {
	reqUrl, err := fe.getFnUrl(fn)
	if err != nil {
		return err
	}

	// For this now assume a standard cold start delay; use profiling to provide a better estimate.
	execAt := expectedAt.Add(-provisionDuration)

	// Tap the Fission function at the right time
	fe.timedExecService.Submit(func() {
		log.WithField("fn", fn).Infof("Tapping Fission function: %v", reqUrl)
		fe.executor.TapService(reqUrl)
	}, execAt)
	return nil
}

func (fe *FunctionEnv) getFnUrl(fn types.FnRef) (*url.URL, error) {
	meta := createFunctionMeta(fn)
	serviceUrl, err := fe.executor.GetServiceForFunction(meta)
	if err != nil {
		log.WithFields(logrus.Fields{
			"err":  err,
			"meta": meta,
		}).Error("Fission function could not be found!")
		return nil, err
	}
	rawUrl := fmt.Sprintf("%s://%s", defaultProtocol, serviceUrl)
	reqUrl, err := url.Parse(rawUrl)
	if err != nil {
		logrus.Errorf("Failed to parse url: '%v'", rawUrl)
		panic(err)
	}
	return reqUrl, nil
}

func createFunctionMeta(fn types.FnRef) *metav1.ObjectMeta {
	parts := strings.SplitN(fn.ID, "-", 2)
	if len(parts) < 2 {
		panic(fmt.Sprintf("invalid function ref for kubernetes object: '%v'", fn))
	}

	return &metav1.ObjectMeta{
		Name:      parts[0],
		UID:       k8stypes.UID(parts[1]),
		Namespace: metav1.NamespaceDefault,
	}
}
