package fission

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/fission/fission"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/types/typedvalues/httpconv"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/fission/fission-workflows/pkg/util/backoff"
	controller "github.com/fission/fission/controller/client"
	"github.com/golang/protobuf/ptypes"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"

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
	executor    *executor.Client
	executorURL string
	controller  *controller.Client
	routerURL   string
	client      *http.Client
}

const (
	defaultHTTPMethod = http.MethodPost
	defaultProtocol   = "http"
)

func New(executorURL, serverURL, routerURL string) *FunctionEnv {
	logger, _ := zap.NewProduction()
	return &FunctionEnv{
		executor:    executor.MakeClient(logger, executorURL),
		controller:  controller.MakeClient(serverURL),
		routerURL:   routerURL,
		executorURL: executorURL,
		client:      &http.Client{},
	}
}

// Invoke executes the task in a blocking way.
//
// spec contains the complete configuration needed for the execution.
// It returns the TaskInvocationStatus with a completed (FINISHED, FAILED, ABORTED) status.
// An error is returned only when error occurs outside of the runtime's control.
func (fe *FunctionEnv) Invoke(spec *types.TaskInvocationSpec, opts ...fnenv.InvokeOption) (*types.TaskInvocationStatus, error) {
	cfg := fnenv.ParseInvokeOptions(opts)
	ctxLog := log.WithField("fn", spec.FnRef)
	if err := validate.TaskInvocationSpec(spec); err != nil {
		return nil, err
	}
	span, _ := opentracing.StartSpanFromContext(cfg.Ctx, "/fnenv/fission")
	defer span.Finish()
	fnRef := *spec.FnRef
	span.SetTag("fnref", fnRef.Format())

	// Construct request and add body
	fnUrl := fe.createRouterURL(fnRef)
	span.SetTag("fnUrl", fnUrl)
	req, err := http.NewRequest(defaultHTTPMethod, fnUrl, nil)
	if err != nil {
		panic(fmt.Errorf("failed to create request for '%v': %v", fnUrl, err))
	}
	// Map task inputs to request
	err = httpconv.FormatRequest(spec.Inputs, req)
	if err != nil {
		return nil, err
	}

	// Add tracing
	if span := opentracing.SpanFromContext(cfg.Ctx); span != nil {
		err := opentracing.GlobalTracer().Inject(span.Context(), opentracing.HTTPHeaders,
			opentracing.HTTPHeadersCarrier(req.Header))
		if err != nil {
			ctxLog.Warnf("Failed to inject opentracing tracer context: %v", err)
		}
	}

	// Perform request
	timeStart := time.Now()
	fnenv.FnActive.WithLabelValues(Name).Inc()
	defer fnenv.FnExecTime.WithLabelValues(Name).Observe(float64(time.Since(timeStart)))
	ctxLog.Infof("Invoking Fission function: '%v'.", req.URL)
	if logrus.GetLevel() == logrus.DebugLevel {
		fmt.Println("--- HTTP Request ---")
		bs, err := httputil.DumpRequest(req, true)
		if err != nil {
			logrus.Error(err)
		}
		fmt.Println(string(bs))
		fmt.Println("--- HTTP Request end ---")
		span.LogKV("HTTP request", string(bs))
	}
	span.LogKV("http", fmt.Sprintf("%s %v", req.Method, req.URL))
	var resp *http.Response

	// Setup  context
	deadline, err := ptypes.Timestamp(spec.Deadline)
	if err != nil {
		return nil, err
	}
	maxAttempts := 12 // About 6 min
	ctx, cancel := context.WithDeadline(cfg.Ctx, deadline)
	for attempt := range (&backoff.Instance{
		MaxRetries:         maxAttempts,
		BaseRetryDuration:  100 * time.Millisecond,
		BackoffPolicy:      backoff.ExponentialBackoff,
		MaxBackoffDuration: 10 * time.Second,
	}).C(ctx) {
		resp, err = fe.client.Do(req.WithContext(cfg.Ctx))
		if err == nil {
			break
		}
		log.Debugf("Failed to execute Fission function at %s (%d/%d): %v", fnUrl, err, attempt, maxAttempts)
	}
	cancel()

	// Check if max try attempts was exceeded
	if resp == nil {
		return nil, fmt.Errorf("error executing fission function at %s after %d attempts: %v", fnUrl, maxAttempts, err)
	}
	span.LogKV("status code", resp.Status)

	fnenv.FnActive.WithLabelValues(Name).Dec()

	ctxLog.Infof("Fission function response: %d - %s", resp.StatusCode, resp.Header.Get("Content-Type"))
	if logrus.GetLevel() == logrus.DebugLevel {
		fmt.Println("--- HTTP Response ---")
		bs, err := httputil.DumpResponse(resp, true)
		if err != nil {
			logrus.Error(err)
		}
		fmt.Println(string(bs))
		fmt.Println("--- HTTP Response end ---")
		span.LogKV("HTTP response", string(bs))
	}

	// Parse output
	output, err := httpconv.ParseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output: %v", err)
	}

	// Parse response headers
	outHeaders := httpconv.ParseResponseHeaders(resp)

	// Determine status of the task invocation
	if resp.StatusCode >= 400 {
		msg, _ := typedvalues.Unwrap(output)
		ctxLog.Warnf("[%s] Failed %v: %v", fnRef.ID, resp.StatusCode, msg)
		return &types.TaskInvocationStatus{
			Status: types.TaskInvocationStatus_FAILED,
			Error: &types.Error{
				Message: fmt.Sprintf("fission function error: %v", msg),
			},
		}, nil
	}

	return &types.TaskInvocationStatus{
		Status:        types.TaskInvocationStatus_SUCCEEDED,
		Output:        output,
		OutputHeaders: outHeaders,
	}, nil
}

// Prepare signals the Fission runtime that a function request is expected at a specific time.
// For now this function will tap immediately regardless of the expected execution time.
func (fe *FunctionEnv) Prepare(fn types.FnRef, expectedAt time.Time) error {
	reqURL, err := fe.getFnURL(fn)
	if err != nil {
		return err
	}

	// Tap the Fission function at the right time
	log.WithField("fn", fn).Infof("Prewarming Fission function: %v", reqURL)
	return fe.tapService(reqURL.String())
}

func (fe *FunctionEnv) Resolve(ref types.FnRef) (string, error) {
	// Currently we just use the controller API to check if the function exists.
	log.Infof("Resolving function: %s", ref.ID)
	ns := ref.Namespace
	if len(ns) == 0 {
		ns = metav1.NamespaceDefault
	}
	_, err := fe.controller.FunctionGet(&metav1.ObjectMeta{
		Name:      ref.ID,
		Namespace: ns,
	})
	if err != nil {
		return "", err
	}
	id := ref.ID

	log.Infof("Resolved fission function %s to %s", ref.ID, id)
	return id, nil
}

func (fe *FunctionEnv) getFnURL(fn types.FnRef) (*url.URL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	meta := createFunctionMeta(fn)
	serviceURL, err := fe.executor.GetServiceForFunction(ctx, meta)
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
	id := strings.TrimLeft(fn.ID, "/")
	baseUrl := strings.TrimRight(fe.routerURL, "/")
	var ns string
	if fn.Namespace != metav1.NamespaceDefault && len(fn.Namespace) != 0 {
		ns = strings.Trim(fn.Namespace, "/") + "/"
	}
	return fmt.Sprintf("%s/fission-function/%s%s", baseUrl, ns, id)
}

func (fe *FunctionEnv) tapService(serviceUrlStr string) error {
	executorUrl := fe.executorURL + "/v2/tapService"

	resp, err := http.Post(executorUrl, "application/octet-stream", bytes.NewReader([]byte(serviceUrlStr)))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fission.MakeErrorFromHTTP(resp)
	}
	return nil
}
