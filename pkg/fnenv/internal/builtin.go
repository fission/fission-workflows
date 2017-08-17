package internal

import (
	"errors"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/sirupsen/logrus"
)

// Temporary file containing built-in internal functions
//
// Should be refactored to a extensible system, using go plugins for example.

type FunctionIf struct {}

func (fn *FunctionIf) Invoke(spec *types.FunctionInvocationSpec) ([]byte, error) {
	return nil, errors.New("Not implemented FunctionIf")
}

type FunctionNoop struct {}

func (fn *FunctionNoop) Invoke(spec *types.FunctionInvocationSpec) ([]byte, error) {
	logrus.WithField("spec", spec).Debug("Noop")
	return nil, nil
}
