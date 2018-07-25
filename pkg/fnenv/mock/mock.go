// Package mock contains a minimal, mocked implementation of a fnenv for test purposes
package mock

import (
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

// Func is the type for mocked functions used in the mock.Runtime
type Func func(spec *types.TaskInvocationSpec) (*types.TypedValue, error)

// Runtime mocks the implementation of the various runtime.
//
// Mock functions can be added to Functions, and should have the mocked function id as the key.
// For AsyncRuntime the results are stored and retrieved from the AsyncResults. The result is added
// automatically/instantly using the function, but can be avoided by enabling ManualExecution.
//
// Note it does not mock the resolver, which is mocked by the mock.Resolver
type Runtime struct {
	Functions       map[string]Func
	AsyncResults    map[string]*types.TaskInvocation
	ManualExecution bool
}

func NewRuntime() *Runtime {
	return &Runtime{
		Functions:    map[string]Func{},
		AsyncResults: map[string]*types.TaskInvocation{},
	}
}

func (mk *Runtime) InvokeAsync(spec *types.TaskInvocationSpec) (string, error) {
	fnName := spec.FnRef.ID

	if _, ok := mk.Functions[fnName]; !ok {
		return "", fmt.Errorf("could not invoke unknown function '%s'", fnName)
	}

	invocationID := util.UID()
	mk.AsyncResults[invocationID] = &types.TaskInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        invocationID,
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec: spec,
		Status: &types.TaskInvocationStatus{
			Status:    types.TaskInvocationStatus_IN_PROGRESS,
			UpdatedAt: ptypes.TimestampNow(),
		},
	}

	if !mk.ManualExecution {
		err := mk.MockComplete(invocationID)
		if err != nil {
			panic(err)
		}
	}

	return invocationID, nil
}

func (mk *Runtime) MockComplete(fnInvocationID string) error {
	invocation, ok := mk.AsyncResults[fnInvocationID]
	if !ok {
		return fmt.Errorf("could not invoke unknown invocation '%s'", fnInvocationID)
	}

	fnName := invocation.Spec.FnRef.ID
	fn, ok := mk.Functions[fnName]
	if !ok {
		return fmt.Errorf("could not invoke unknown function '%s'", fnName)
	}

	result, err := fn(invocation.Spec)
	if err != nil {
		logrus.Infof("Function '%s' invocation resulted in an error: %v", fnName, err)
		mk.AsyncResults[fnInvocationID].Status = &types.TaskInvocationStatus{
			Output:    nil,
			UpdatedAt: ptypes.TimestampNow(),
			Status:    types.TaskInvocationStatus_FAILED,
		}
	} else {
		mk.AsyncResults[fnInvocationID].Status = &types.TaskInvocationStatus{
			Output:    result,
			UpdatedAt: ptypes.TimestampNow(),
			Status:    types.TaskInvocationStatus_SUCCEEDED,
		}
	}

	return nil
}

func (mk *Runtime) Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocationStatus, error) {
	logrus.Info("Starting invocation...")
	invocationID, err := mk.InvokeAsync(spec)
	if err != nil {
		return nil, err
	}
	err = mk.MockComplete(invocationID)
	if err != nil {
		return nil, err
	}

	logrus.Infof("...completing function execution for '%v'", invocationID)
	return mk.Status(invocationID)
}

func (mk *Runtime) Cancel(fnInvocationID string) error {
	invocation, ok := mk.AsyncResults[fnInvocationID]
	if !ok {
		return fmt.Errorf("could not invoke unknown invocation '%s'", fnInvocationID)
	}

	invocation.Status = &types.TaskInvocationStatus{
		Output:    nil,
		UpdatedAt: ptypes.TimestampNow(),
		Status:    types.TaskInvocationStatus_ABORTED,
	}

	return nil
}

func (mk *Runtime) Status(fnInvocationID string) (*types.TaskInvocationStatus, error) {
	invocation, ok := mk.AsyncResults[fnInvocationID]
	if !ok {
		return nil, fmt.Errorf("could not invoke unknown invocation '%s'", fnInvocationID)
	}

	return invocation.Status, nil
}

func (mk *Runtime) Notify(taskID string, fn types.FnRef, expectedAt time.Time) error {
	return nil
}

// Resolver is a mocked implementation of a RuntimeResolver.
//
// Use FnNameIDs to setup a mapping to mock resolving function references to IDs.
type Resolver struct {
	FnNameIDs map[string]string
}

func (mf *Resolver) Resolve(ref types.FnRef) (string, error) {
	fnID, ok := mf.FnNameIDs[ref.ID]
	if !ok {
		return "", fmt.Errorf("could not resolve function '%s' using resolve-map '%v'", ref.ID, mf.FnNameIDs)
	}

	return fnID, nil
}

func NewResolver() *Resolver {
	return &Resolver{
		FnNameIDs: map[string]string{},
	}
}
