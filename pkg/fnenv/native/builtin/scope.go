package builtin

import (
	"errors"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/sirupsen/logrus"
)

const (
	SCOPE_INPUT = types.INPUT_MAIN
)

type FunctionScope struct{}

func (fn *FunctionScope) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {
	scope, ok := spec.GetInputs()[NOOP_INPUT]
	if !ok {
		return nil, errors.New("missing scope input")
	}

	logrus.WithFields(logrus.Fields{
		"spec":        spec,
		"output.type": scope.Type,
	}).Info("Internal Scope-function invoked.")

	return scope, nil
}
