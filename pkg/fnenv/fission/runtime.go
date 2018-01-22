package fission

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/sirupsen/logrus"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"

	executor "github.com/fission/fission/executor/client"
	"github.com/fission/fission/router"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

// FunctionEnv adapts the Fission platform to the function execution runtime. This allows the workflow engine
// to invoke Fission functions.
type FunctionEnv struct {
	executor *executor.Client
}

const (
	defaultHttpMethod = http.MethodPost
	defaultProtocol   = "http"
)

func NewFunctionEnv(executor *executor.Client) *FunctionEnv {
	return &FunctionEnv{
		executor: executor,
	}
}

func (fe *FunctionEnv) Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocationStatus, error) {
	meta := &metav1.ObjectMeta{
		Name:      spec.GetType().GetSrc(),
		UID:       k8stypes.UID(spec.GetType().GetResolved()),
		Namespace: metav1.NamespaceDefault,
	}
	logrus.WithFields(logrus.Fields{
		"name": meta.Name,
		"uid":  meta.UID,
		"ns":   meta.Namespace,
	}).Info("Invoking Fission function.")

	// Get reqUrl
	// TODO use router instead once we can route to a specific function uid
	serviceUrl, err := fe.executor.GetServiceForFunction(meta)
	if err != nil {
		logrus.WithFields(logrus.Fields{
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

	// Construct request and add body
	req, err := http.NewRequest(defaultHttpMethod, reqUrl.String(), nil)
	if err != nil {
		panic(fmt.Errorf("failed to make request for '%s': %v", serviceUrl, err))
	}

	// Map task inputs to request
	formatRequest(req, spec.Inputs)

	// Add parameters normally added by Fission
	router.MetadataToHeaders(router.HEADERS_FISSION_FUNCTION_PREFIX, meta, req)

	// Perform request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error for reqUrl '%s': %v", serviceUrl, err)
	}

	// Parse output
	output, err := parseBody(resp.Body, resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse output: %v", err)
	}

	logrus.Infof("[%s][Content-Type]: %v ", meta.Name, resp.Header.Get("Content-Type"))
	logrus.Infof("[%s][output]: %v", meta.Name, output.Short())
	logrus.Infof("[%s][status]: %v", meta.Name, resp.StatusCode)

	// Determine status of the task invocation
	if resp.StatusCode >= 400 {
		msg, _ := typedvalues.Format(output)
		logrus.Warn("[%s] Failed %v: %v", resp.StatusCode, msg)
		return &types.TaskInvocationStatus{
			Status: types.TaskInvocationStatus_FAILED,
			Error: &types.Error{
				Code:    fmt.Sprintf("%v", resp.StatusCode),
				Message: fmt.Sprintf("%s", msg),
			},
		}, nil
	}

	return &types.TaskInvocationStatus{
		Status: types.TaskInvocationStatus_SUCCEEDED,
		Output: output,
	}, nil
}
