package fission

import (
	"net/http"

	"context"

	"io/ioutil"

	"os"

	"bytes"

	"encoding/json"

	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/fnenv/fission/router"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/sirupsen/logrus"
	"k8s.io/client-go/1.5/pkg/api"
)

const userFunc = "/userfunc/user"

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
	mux.HandleFunc("/specialize", fp.handleSpecialize)
}

func (fp *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logrus.Infof("Handling incoming request '%s'... ", r.URL)

	meta, _ := router.HeadersToMetadata(router.HEADERS_FISSION_FUNCTION_PREFIX, r.Header)

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
		wf, err := fp.workflowServer.Get(ctx, &apiserver.WorkflowIdentifier{fnId})
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
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to read body: " + err.Error()))
		return
	}

	metadata := &api.ObjectMeta{}
	err = json.Unmarshal(bs, metadata)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to parse body: " + err.Error()))
		return
	}

	wfId, err := fp.specialize(ctx, metadata, userFunc)
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

func (fp *Proxy) specialize(ctx context.Context, fissionMetadata *api.ObjectMeta, fnPath string) (string, error) {
	_, err := os.Stat(fnPath)
	if err != nil {
		return "", err
	}

	rdr, err := os.Open(userFunc)
	if err != nil {
		return "", err
	}
	raw, err := ioutil.ReadAll(rdr)
	if err != nil {
		panic(err)
	}
	logrus.Infof("received definition: %s", string(raw))

	wfSpec := &types.WorkflowSpec{}
	err = jsonpb.Unmarshal(bytes.NewReader(raw), wfSpec)
	if err != nil {
		return "", err
	}
	logrus.WithField("wfSpec", wfSpec).Info("Received valid WorkflowSpec from fetcher.")

	// Synchronize the workflow ID with the fission ID
	wfSpec.Id = string(fissionMetadata.GetUID())
	wfSpec.Name = fissionMetadata.Name

	resp, err := fp.workflowServer.Create(ctx, wfSpec)
	if err != nil {
		return "", err
	}
	wfId := resp.Id

	// Cache the ID so we don't have to check whether the workflow engine already has it.
	fp.fissionIds[wfSpec.Id] = true

	// Cleanup
	err = os.Remove(userFunc)
	if err != nil {
		logrus.Warnf("Failed to remove userFunc from fs: %v", err)
	}

	return wfId, nil
}
