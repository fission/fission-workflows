package native

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
)

// Note: package is called 'native' because 'internal' is not an allowed package name.

// An InternalFunction is a function that will be executed in the same process as the invoker.
type InternalFunction interface {
	Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error)
}

// FunctionEnv for executing low overhead functions, such as control flow constructs, inside the workflow engine
//
// Note: This currently supports Golang only.
type FunctionEnv struct {
	fns map[string]InternalFunction // Name -> function
}

func NewFunctionEnv(fns map[string]InternalFunction) *FunctionEnv {
	env := &FunctionEnv{
		fns: fns,
	}
	log.WithField("fns", env.fns).Debugf("Internal function runtime installed.")
	return env
}

func (fe *FunctionEnv) Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocationStatus, error) {
	fnId := spec.GetType().GetResolved()
	fn, ok := fe.fns[fnId]
	if !ok {
		return nil, fmt.Errorf("could not resolve internal function '%s'", fnId)
	}

	out, err := fn.Invoke(spec)
	if err != nil {
		log.WithFields(log.Fields{
			"fnId": fnId,
			"err":  err,
		}).Error("Internal function failed.")
		return &types.TaskInvocationStatus{
			UpdatedAt: ptypes.TimestampNow(),
			Status:    types.TaskInvocationStatus_FAILED,
		}, nil
	}

	return &types.TaskInvocationStatus{
		UpdatedAt: ptypes.TimestampNow(),
		Status:    types.TaskInvocationStatus_SUCCEEDED,
		Output:    out,
	}, nil
}

func (fe *FunctionEnv) Resolve(fnName string) (string, error) {
	_, ok := fe.fns[fnName]
	if !ok {
		return "", fmt.Errorf("could not resolve internal function '%s'", fnName)
	}
	log.WithField("name", fnName).WithField("uid", fnName).Info("Resolved internal function")
	return fnName, nil
}

func (fe *FunctionEnv) RegisterFn(name string, fn InternalFunction) {
	fe.fns[name] = fn
}
