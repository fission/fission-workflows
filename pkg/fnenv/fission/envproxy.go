package fission

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/fission/fission"
	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues/httpconv"
	"github.com/fission/fission/router"
	"github.com/golang/protobuf/jsonpb"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/syncmap"
)

// Proxy between Fission and ParseWorkflow to ensure that workflowInvocations comply with Fission function interface. This
// ensures that workflows can be executed exactly like Fission functions are executed.
type Proxy struct {
	// TODO change server to client
	invocationServer apiserver.WorkflowInvocationAPIServer
	workflowServer   apiserver.WorkflowAPIServer
	fissionIds       syncmap.Map // map[string]bool
}

// NewFissionProxyServer creates a proxy server to adheres to the Fission Environment specification.
func NewFissionProxyServer(wfiSrv apiserver.WorkflowInvocationAPIServer, wfSrv apiserver.WorkflowAPIServer) *Proxy {
	return &Proxy{
		invocationServer: wfiSrv,
		workflowServer:   wfSrv,
		fissionIds:       syncmap.Map{},
	}
}

// RegisterServer adds the endpoints needed for the Fission Environment interface to the mux server.
func (fp *Proxy) RegisterServer(mux *http.ServeMux) {
	mux.HandleFunc("/", fp.handleRequest)
	mux.HandleFunc("/v2/specialize", fp.handleSpecialize)
	mux.HandleFunc("/healthz", fp.handleHealthCheck)
}

func (fp *Proxy) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (fp *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	// Optional: Parse opentracing
	spanCtx, err := opentracing.GlobalTracer().Extract(opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(r.Header))
	if err != nil && err != opentracing.ErrSpanContextNotFound {
		logrus.Warnf("Failed to extract tracer from proxy requrest: %v", err)
	}
	var opts []opentracing.StartSpanOption
	if spanCtx != nil {
		opts = append(opts, opentracing.ChildOf(spanCtx))
	}
	span := opentracing.StartSpan("fnenv/fission/envproxy.handleRequest", opts...)
	defer span.Finish()

	// Fetch the workflow based on the received Fission function metadata
	meta := router.HeadersToMetadata(router.HEADERS_FISSION_FUNCTION_PREFIX, r.Header)
	fnID := string(meta.UID)
	if len(meta.UID) == 0 {
		logrus.WithField("meta", meta).Error("Fission function name is missing")
		http.Error(w, "Fission function name is missing", 400)
		return
	}

	// Check if the workflow engine contains the workflow associated with the Fission function UID
	_, ok := fp.fissionIds.Load(fnID)
	if !ok {
		// Fallback 1 : check if it is in the event store somewhere
		if fp.hasWorkflow(ctx, fnID) {
			logrus.WithField("fnID", fnID).
				Error("Unknown fission function name")
			http.Error(w, "Unknown fission function name; not specialized", 400)
			return
		}
		fp.fissionIds.Store(fnID, true)
	}

	// Map request to workflow inputs
	inputs, err := httpconv.ParseRequest(r)
	if err != nil {
		logrus.Errorf("Failed to parse inputs: %v", err)
		http.Error(w, "Failed to parse inputs", 400)
		return
	}
	wfSpec := &types.WorkflowInvocationSpec{
		WorkflowId: fnID,
		Inputs:     inputs,
	}

	// Temporary: in case of query header 'X-Async' being present, make request async
	if len(r.Header.Get("X-Async")) > 0 {
		invocationID, invokeErr := fp.invocationServer.Invoke(ctx, wfSpec)
		if invokeErr != nil {
			logrus.Errorf("Failed to invoke: %v", invokeErr)
			http.Error(w, invokeErr.Error(), 500)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(invocationID.Id))
		return
	}

	// Otherwise, the request synchronous like other Fission functions
	wi, err := fp.invocationServer.InvokeSync(ctx, wfSpec)
	if err != nil {
		logrus.Errorf("Failed to invoke: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Format output or error to the response
	if !wi.Status.Successful() && wi.Status.Error == nil {
		logrus.Warn("Failed invocation does not contain error")
		wi.Status.Error = &types.Error{
			Message: "Unknown error",
		}
	}
	// Logging
	if !wi.Status.Successful() {
		logrus.Errorf("Invocation not successful, was '%v': %v", wi.Status.Status.String(), wi.Status.Error.Error())
	} else if wi.Status.Output == nil {
		logrus.Infof("Invocation '%v' has no output.", fnID)
	} else {
		logrus.Infof("Response Content-Type: %v", httpconv.DetermineContentType(wi.Status.Output))
	}

	// Get output
	httpconv.FormatResponse(w, wi.Status.Output, wi.Status.Error)
}

func (fp *Proxy) handleSpecialize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logrus.Info("Specializing...")
	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		logrus.Errorf("Failed to read body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to read body: " + err.Error()))
		return
	}

	logrus.Info("Received specialization body:", string(bs))

	// Parse the function load request
	flr := &fission.FunctionLoadRequest{}
	err = json.Unmarshal(bs, flr)
	if err != nil {
		logrus.Errorf("Failed to parse body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to parse body: " + err.Error()))
		return
	}

	// Attempt to specialize with the provided function load request
	wfIDs, err := fp.Specialize(ctx, flr)
	if err != nil {
		logrus.Errorf("failed to specialize: %v", err)
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Write([]byte(err.Error()))
		return
	}
	logrus.Infof("Specialized successfully wf '%s'", wfIDs)
	w.WriteHeader(http.StatusOK)
	// TODO change this to a more structured output once specialization protocol has been formalized
	w.Write([]byte(strings.Join(wfIDs, ";")))
}

// Specialize creates workflows provided by a Fission Load Request.
//
// The Fission package can either exist out of a single workflow file, or out of a directory filled with
// solely workflow definitions.
func (fp *Proxy) Specialize(ctx context.Context, flr *fission.FunctionLoadRequest) ([]string, error) {
	if flr == nil {
		return nil, errors.New("no function load request provided")
	}

	// Check if request contains the expected metadata
	metadata := flr.FunctionMetadata
	if metadata == nil {
		logrus.Errorf("No workflow metadata provided: %v", metadata)
		return nil, errors.New("no metadata provided in function load request")
	}

	// Search provided package for workflow definitions
	wfPaths, err := fp.findWorkflowFiles(ctx, flr)
	if err != nil {
		return nil, err
	}
	var wfIDs []string

	// Create workflows for each detected workflow definition
	logrus.Infof("Files: %v", wfPaths)
	for _, wfPath := range wfPaths {
		wfID, err := fp.createWorkflowFromFile(ctx, flr, wfPath)
		if err != nil {
			return nil, fmt.Errorf("failed to specialize package: %v", err)
		}
		wfIDs = append(wfIDs, wfID)
	}

	return wfIDs, nil
}

// findWorkflowFiles searches for all of the workflow definitions in a Fission package
func (fp *Proxy) findWorkflowFiles(ctx context.Context, flr *fission.FunctionLoadRequest) ([]string, error) {
	fi, err := os.Stat(flr.FilePath)
	if err != nil {
		return nil, fmt.Errorf("no file present at '%v", flr.FilePath)
	}

	if fi.IsDir() {
		// User provided a package
		contents, err := ioutil.ReadDir(flr.FilePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read package: '%v", flr.FilePath)
		}

		if len(contents) == 0 {
			return nil, fmt.Errorf("package is empty")
		}

		var paths []string
		for _, n := range contents {
			// Note: this assumes that all files in package are workflow definitions!
			file := path.Join(flr.FilePath, n.Name())
			paths = append(paths, file)
		}

		// TODO maybe change to []string to provide all generated ids
		return paths, nil
	}
	// Provided workflow is a file
	return []string{flr.FilePath}, nil
}

// createWorkflowFromFile creates a workflow given a path to the file containing the workflow definition
func (fp *Proxy) createWorkflowFromFile(ctx context.Context, flr *fission.FunctionLoadRequest, path string) (string, error) {

	rdr, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("failed to open file '%v': %v", path, err)
	}
	raw, err := ioutil.ReadAll(rdr)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%v': %v", path, err)
	}
	logrus.Infof("received definition: '%s'", string(raw))

	wfSpec := &types.WorkflowSpec{}
	err = jsonpb.Unmarshal(bytes.NewReader(raw), wfSpec)
	if err != nil {
		return "", fmt.Errorf("failed to parse bytes into WorkflowSpec: %v", err)
	}
	logrus.WithField("wfSpec", wfSpec).Info("Received valid WorkflowSpec from fetcher.")

	// Synchronize the workflow id with the fission id
	fissionID := string(flr.FunctionMetadata.GetUID())
	wfSpec.ForceId = fissionID
	wfSpec.Name = flr.FunctionMetadata.Name

	resp, err := fp.workflowServer.Create(ctx, wfSpec)
	if err != nil {
		return "", fmt.Errorf("failed to store workflow internally: %v", err)
	}
	wfID := resp.Id

	// EvalCache the id so we don't have to check whether the workflow engine already has it.
	fp.fissionIds.Store(fissionID, true)

	return wfID, nil
}

func (fp *Proxy) hasWorkflow(ctx context.Context, fnID string) bool {
	wf, err := fp.workflowServer.Get(ctx, &apiserver.WorkflowIdentifier{Id: fnID})
	if err != nil {
		logrus.Errorf("Failed to get workflow: %v; assuming it is non-existent", err)
	}
	return wf != nil
}
