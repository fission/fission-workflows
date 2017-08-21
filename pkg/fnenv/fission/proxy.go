package fission

import (
	"net/http"

	"context"

	"io/ioutil"

	"github.com/fission/fission-workflow/pkg/apiserver"
	"github.com/fission/fission-workflow/pkg/types"
)

// Proxy between Fission and Workflow to ensure that workflowInvocations comply with Fission function interface
type Proxy struct {
	invocationServer apiserver.WorkflowInvocationAPIServer
}

func NewFissionProxyServer(srv apiserver.WorkflowInvocationAPIServer) *Proxy {
	return &Proxy{srv}
}

func (fp *Proxy) RegisterServer(mux *http.ServeMux) {
	mux.HandleFunc("/", fp.handleRequest)
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
	input := map[string]*types.TypedValue{
		types.INPUT_MAIN: {
			// TODO infer type from headers
			Value: body,
		},
	}

	ctx := context.Background()
	invocation, err := fp.invocationServer.InvokeSync(ctx, &types.WorkflowInvocationSpec{
		WorkflowId: id,
		Input:      input,
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
