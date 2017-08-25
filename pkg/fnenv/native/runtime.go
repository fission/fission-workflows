package native

import (
	"fmt"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
)

// An InternalFunction is a function that will be executed in the same process as the invoker.
type InternalFunction interface {
	Invoke(spec *types.FunctionInvocationSpec) (*types.TypedValue, error)
}

// Internal InternalFunction Environment for executing low overhead functions, such as control flow constructs
//
// Currently this is golang only.
type FunctionEnv struct {
	fns map[string]InternalFunction // Name -> function
}

func NewFunctionEnv() *FunctionEnv {
	env := &FunctionEnv{
		fns: map[string]InternalFunction{
			"if":   &FunctionIf{},
			"noop": &FunctionNoop{},
		},
	}
	log.WithField("fns", env.fns).Debugf("Internal function runtime installed.")
	return env
}

func (fe *FunctionEnv) Invoke(spec *types.FunctionInvocationSpec) (*types.FunctionInvocationStatus, error) {
	fnId := spec.GetType().GetResolved()
	fn, ok := fe.fns[fnId]
	if !ok {
		return nil, fmt.Errorf("Could not resolve internal function '%s'.", fnId)
	}

	out, err := fn.Invoke(spec)
	if err != nil {
		log.WithFields(log.Fields{
			"fnId": fnId,
			"err":  err,
		}).Error("Internal function failed.")
		return &types.FunctionInvocationStatus{
			UpdatedAt: ptypes.TimestampNow(),
			Status:    types.FunctionInvocationStatus_FAILED,
		}, nil
	}

	return &types.FunctionInvocationStatus{
		UpdatedAt: ptypes.TimestampNow(),
		Status:    types.FunctionInvocationStatus_SUCCEEDED,
		Output:    out,
	}, nil
}

func (fe *FunctionEnv) Resolve(fnName string) (string, error) {
	_, ok := fe.fns[fnName]
	if !ok {
		return "", fmt.Errorf("Could not resolve internal function '%s'.", fnName)
	}
	return fnName, nil
}

func (fe *FunctionEnv) RegisterFn(name string, fn InternalFunction) {
	fe.fns[name] = fn
}
