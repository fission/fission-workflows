package fission

import (
	"net/http"

	"context"

	"io/ioutil"

	"os"

	"bytes"

	"encoding/json"

	"fmt"

	"github.com/fission/fission"
	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission/router"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/sirupsen/logrus"
)

// Proxy between Fission and Workflow to ensure that workflowInvocations comply with Fission function interface
type Proxy struct {
	invocationServer apiserver.WorkflowInvocationAPIServer
	workflowServer   apiserver.WorkflowAPIServer
	fissionIds       map[string]bool
}

func NewFissionProxyServer(wfiSrv apiserver.WorkflowInvocationAPIServer, wfSrv apiserver.WorkflowAPIServer) *Proxy {
	return &Proxy{
		invocationServer: wfiSrv,
		workflowServer:   wfSrv,
		fissionIds:       map[string]bool{},
	}
}

func (fp *Proxy) RegisterServer(mux *http.ServeMux) {
	mux.HandleFunc("/", fp.handleRequest)
	mux.HandleFunc("/v2/specialize", fp.handleSpecialize)
}

func (fp *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logrus.Infof("Handling incoming request '%s'... ", r.URL)

	meta := router.HeadersToMetadata(router.HEADERS_FISSION_FUNCTION_PREFIX, r.Header)

	fnId := string(meta.UID)
	if len(meta.UID) == 0 {
		logrus.WithField("meta", meta).Error("Fission function name is missing")
		http.Error(w, "Fission function name is missing", 400)
		return
	}

	// Map fission function name to workflow id
	_, ok := fp.fissionIds[fnId]

	if !ok {
		// Fallback 1 : check if it is in the event store somewhere
		wf, err := fp.workflowServer.Get(ctx, &apiserver.WorkflowIdentifier{Id: fnId})
		if err != nil || wf == nil {
			logrus.WithField("fnId", fnId).
				WithField("err", err).
				WithField("map", fp.fissionIds).
				Error("Unknown fission function name")

			logrus.Warn(fp.fissionIds)
			http.Error(w, "Unknown fission function name", 400)
			return
		}
		fp.fissionIds[fnId] = true
	}

	// Map Inputs to function parameters
	inputs := map[string]*types.TypedValue{}
	err := ParseRequest(r, inputs)
	if err != nil {
		logrus.Errorf("Failed to parse inputs: %v", err)
		http.Error(w, "Failed to parse inputs", 400)
		return
	}

	// Temporary: in case of query header 'X-Async' being present, make request async
	if len(r.Header.Get("X-Async")) > 0 {
		invocatinId, err := fp.invocationServer.Invoke(ctx, &types.WorkflowInvocationSpec{
			WorkflowId: fnId,
			Inputs:     inputs,
		})
		if err != nil {
			logrus.Errorf("Failed to invoke: %v", err)
			http.Error(w, err.Error(), 500)
			return
		}
		w.WriteHeader(200)
		w.Write([]byte(invocatinId.Id))
		return
	}

	// Otherwise, the request synchronous like other Fission functions
	invocation, err := fp.invocationServer.InvokeSync(ctx, &types.WorkflowInvocationSpec{
		WorkflowId: fnId,
		Inputs:     inputs,
	})
	if err != nil {
		logrus.Errorf("Failed to invoke: %v", err)
		http.Error(w, err.Error(), 500)
		return
	}

	if !invocation.Status.Status.Successful() {
		logrus.Errorf("Invocation not successful, was '%v'", invocation.Status.Status.String())
		http.Error(w, invocation.Status.Status.String(), 500)
		return
	}

	// TODO determine header based on the output value
	var resp []byte
	if invocation.Status.Output != nil {
		resp = invocation.Status.Output.Value
		w.Header().Add("Content-Type", ToContentType(invocation.Status.Output))
	} else {
		logrus.Infof("Invocation '%v' has no output.", fnId)
	}
	w.WriteHeader(200)
	w.Write(resp)
}

func (fp *Proxy) handleSpecialize(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logrus.Info("Specializing...")
	body := r.Body
	bs, err := ioutil.ReadAll(body)
	if err != nil {
		logrus.Errorf("Failed to read body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to read body: " + err.Error()))
		return
	}

	logrus.Info("Received specialization body:", string(bs))

	flr := &fission.FunctionLoadRequest{}
	err = json.Unmarshal(bs, flr)
	if err != nil {
		logrus.Errorf("Failed to parse body: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to parse body: " + err.Error()))
		return
	}

	metadata := flr.FunctionMetadata
	if metadata == nil {
		logrus.Errorf("No workflow metadata provided: %v", metadata)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("No workflow metadata provided."))
		return
	}

	wfId, err := fp.specialize(ctx, flr)
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
	logrus.Infof("Specialized successfully wf '%s'", wfId)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(wfId))
}

func (fp *Proxy) specialize(ctx context.Context, flr *fission.FunctionLoadRequest) (string, error) {
	_, err := os.Stat(flr.FilePath)
	if err != nil {
		return "", fmt.Errorf("no file present at '%v", flr.FilePath)
	}

	rdr, err := os.Open(flr.FilePath)
	if err != nil {
		return "", fmt.Errorf("failed to open file '%v': %v", flr.FilePath, err)
	}
	raw, err := ioutil.ReadAll(rdr)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%v': %v", flr.FilePath, err)
	}
	logrus.Infof("received definition: %s", string(raw))

	wfSpec := &types.WorkflowSpec{}
	err = jsonpb.Unmarshal(bytes.NewReader(raw), wfSpec)
	if err != nil {
		return "", fmt.Errorf("failed to parse bytes into WorkflowSpec: %v", err)
	}
	logrus.WithField("wfSpec", wfSpec).Info("Received valid WorkflowSpec from fetcher.")

	// Synchronize the workflow ID with the fission ID
	wfSpec.Id = string(flr.FunctionMetadata.GetUID())
	wfSpec.Name = flr.FunctionMetadata.Name

	resp, err := fp.workflowServer.Create(ctx, wfSpec)
	if err != nil {
		return "", fmt.Errorf("failed to store workflow internally: %v", err)
	}
	wfId := resp.Id

	// Cache the ID so we don't have to check whether the workflow engine already has it.
	fp.fissionIds[wfSpec.Id] = true

	// Cleanup
	err = os.Remove(flr.FilePath)
	if err != nil {
		logrus.Warnf("Failed to remove userFunc from fs: %v", err)
	}

	return wfId, nil
}
