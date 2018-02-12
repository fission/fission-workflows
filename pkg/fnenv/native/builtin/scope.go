package builtin

import (
	"errors"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/sirupsen/logrus"
)

const (
	ScopeInput = types.INPUT_MAIN
)

type FunctionScope struct{}

func (fn *FunctionScope) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {
	scope, ok := spec.GetInputs()[NoopInput]
	if !ok {
		return nil, errors.New("missing scope input")
	}

	logrus.WithFields(logrus.Fields{
		"spec":        spec,
		"output.type": scope.Type,
	}).Info("Internal Scope-function invoked.")

	return scope, nil
}
