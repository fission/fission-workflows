package fission

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/types/typedvalues/httpconv"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/opentracing/opentracing-go"
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
func (fe *FunctionEnv) Invoke(spec *types.TaskInvocationSpec, opts ...fnenv.InvokeOption) (*types.TaskInvocationStatus, error) {
	cfg := fnenv.ParseInvokeOptions(opts)
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
	}
	resp, err := http.DefaultClient.Do(req.WithContext(cfg.Ctx))
	if err != nil {
		return nil, fmt.Errorf("error for reqUrl '%v': %v", url, err)
	}

	fnenv.FnActive.WithLabelValues(Name).Dec()
	fnenv.FnActive.WithLabelValues(Name).Inc()

	ctxLog.Infof("Fission function response: %d - %s", resp.StatusCode, resp.Header.Get("Content-Type"))
	if logrus.GetLevel() == logrus.DebugLevel {
		fmt.Println("--- HTTP Response ---")
		bs, err := httputil.DumpResponse(resp, true)
		if err != nil {
			logrus.Error(err)
		}
		fmt.Println(string(bs))
		fmt.Println("--- HTTP Response end ---")
	}

	// Parse output
	output, err := httpconv.ParseResponse(resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse output: %v", err)
	}

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
		Status: types.TaskInvocationStatus_SUCCEEDED,
		Output: output,
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
	id := strings.TrimLeft(fn.ID, "/")
	baseUrl := strings.TrimRight(fe.routerURL, "/")
	var ns string
	if fn.Namespace != metav1.NamespaceDefault && len(fn.Namespace) != 0 {
		ns = strings.Trim(fn.Namespace, "/") + "/"
	}
	return fmt.Sprintf("%s/fission-function/%s%s", baseUrl, ns, id)
}
