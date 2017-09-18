package fission

import (
	"net/http"

	"context"

	"io/ioutil"

	"os"

	"bytes"

	"github.com/fission/fission-workflow/pkg/apiserver"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/sirupsen/logrus"
)

const userFunc = "/userfunc/user"

// Proxy between Fission and Workflow to ensure that workflowInvocations comply with Fission function interface
type Proxy struct {
	invocationServer          apiserver.WorkflowInvocationAPIServer
	workflowServer            apiserver.WorkflowAPIServer
	fissionNameWorkflowUidMap map[string]string // TODO replace this with better solution
}

func NewFissionProxyServer(wfiSrv apiserver.WorkflowInvocationAPIServer, wfSrv apiserver.WorkflowAPIServer) *Proxy {
	return &Proxy{wfiSrv, wfSrv, map[string]string{}}
}

func (fp *Proxy) RegisterServer(mux *http.ServeMux) {
	mux.HandleFunc("/", fp.handleRequest)
	mux.HandleFunc("/specialize", fp.handleSpecialize)
}

func (fp *Proxy) handleRequest(w http.ResponseWriter, r *http.Request) {
	logrus.Infof("Handling incoming request '%s'... ", r.URL)
	fnName := r.Header.Get("X-Fission-Function-Name")
	if len(fnName) == 0 {
		logrus.Error("Fission function name is missing")
		http.Error(w, "Fission function name is missing", 400)
		return
	}
	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		panic(err)
	}

	// Map fission function name to workflow id
	id, ok := fp.fissionNameWorkflowUidMap[fnName]
	if !ok {
		logrus.WithField("fnName", fnName).WithField("map", fp.fissionNameWorkflowUidMap).
			Error("Unknown fission function name")
		http.Error(w, "Unknown fission function name", 400)
		return
	}

	// Map Inputs to function parameters
	inputs := map[string]*types.TypedValue{
		types.INPUT_MAIN: {
			Type:  typedvalues.FormatType(typedvalues.TYPE_RAW),
			Value: body,
		},
	}

	ctx := context.Background()
	invocation, err := fp.invocationServer.InvokeSync(ctx, &types.WorkflowInvocationSpec{
		WorkflowId: id,
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
	resp := invocation.Status.Output.Value
	w.WriteHeader(200)
	w.Write(resp)
}

func (fp *Proxy) handleSpecialize(w http.ResponseWriter, r *http.Request) {
	logrus.Info("Specializing...")

	body := r.Body
	bs, err := ioutil.ReadAll(body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("failed to read body: " + err.Error()))
		return
	}
	fnName := string(bs)
	logrus.Infof("Fission function name: %s", fnName)

	wfId, err := fp.specialize(fnName, userFunc)
	if err != nil {
		logrus.Errorf("failed to specialize: %v", err)
		// TODO differentiate between not found and internal error
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

func (fp *Proxy) specialize(name string, fnPath string) (string, error) {
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

	ctx := context.Background()
	resp, err := fp.workflowServer.Create(ctx, wfSpec)
	if err != nil {
		return "", err
	}
	wfId := resp.Id

	// Map fission function name to generated workflow id
	fp.fissionNameWorkflowUidMap[name] = wfId

	// Cleanup
	err = os.Remove(userFunc)
	if err != nil {
		logrus.Warnf("Failed to remove userFunc from fs: %v", err)
	}

	return wfId, nil
}
