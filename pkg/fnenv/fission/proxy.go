package fission

import (
	"net/http"

	"context"

	"io/ioutil"

	"os"

	"github.com/fission/fission-workflow/pkg/apiserver"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/sirupsen/logrus"
)

const userFunc = "/userfunc/user"

// Proxy between Fission and Workflow to ensure that workflowInvocations comply with Fission function interface
type Proxy struct {
	invocationServer apiserver.WorkflowInvocationAPIServer
	workflowServer   apiserver.WorkflowAPIServer
}

func NewFissionProxyServer(wfiSrv apiserver.WorkflowInvocationAPIServer, wfSrv apiserver.WorkflowAPIServer) *Proxy {
	return &Proxy{wfiSrv, wfSrv}
}

func (fp *Proxy) RegisterServer(mux *http.ServeMux) {
	mux.HandleFunc("/", fp.handleRequest)
	mux.HandleFunc("/specialize", fp.handleSpecialize)
}

func (fp *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("workflowId")
	if len(id) == 0 {
		http.Error(w, "WorkflowId is missing", 400)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		panic(err)
	}

	// Map Inputs to function parameters
	inputs := map[string]*types.TypedValue{
		types.INPUT_MAIN: {
			// TODO infer type from headers
			Value: body,
		},
	}

	ctx := context.Background()
	invocation, err := fp.invocationServer.InvokeSync(ctx, &types.WorkflowInvocationSpec{
		WorkflowId: id,
		Inputs:     inputs,
	})
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if !invocation.Status.Status.Successful() {
		http.Error(w, invocation.Status.Status.String(), 500)
		return
	}

	// TODO determine header based on the output value
	resp := invocation.Status.Output.Value
	w.WriteHeader(200)
	w.Write(resp)
}

func (fp *Proxy) handleSpecialize(w http.ResponseWriter, r *http.Request) {
	_, err := os.Stat(userFunc)
	if err != nil {
		if os.IsNotExist(err) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(userFunc + ": not found"))
			return
		} else {
			panic(err)
		}
	}

	rdr, err := os.Open(userFunc)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to read executable."))
		return
	}

	wfSpec := &types.WorkflowSpec{}
	err = jsonpb.Unmarshal(rdr, wfSpec)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to parse workflow."))
		return
	}

	ctx := context.Background()
	wfId, err := fp.workflowServer.Create(ctx, wfSpec)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Failed to store workflow."))
		return
	}

	err = os.Remove(userFunc)
	if err != nil {
		logrus.Warnf("Failed to remove userFunc: %v", err)
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(wfId.Id))
}
