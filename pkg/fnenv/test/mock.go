package test

import (
	"fmt"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/util"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

type MockFunc func(spec *types.FunctionInvocationSpec) ([]byte, error)

type MockRuntimeEnv struct {
	Functions       map[string]MockFunc
	Results         map[string]*types.FunctionInvocation
	ManualExecution bool
}

// Running invoke will change the state of the function invocation to IN_PROGRESS
func (mk *MockRuntimeEnv) InvokeAsync(spec *types.FunctionInvocationSpec) (string, error) {
	fnName := spec.GetType().GetResolved()

	if _, ok := mk.Functions[fnName]; !ok {
		return "", fmt.Errorf("Could not invoke unknown function '%s'", fnName)
	}

	invocationId := util.Uid()
	mk.Results[invocationId] = &types.FunctionInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        invocationId,
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec: spec,
		Status: &types.FunctionInvocationStatus{
			Status:    types.FunctionInvocationStatus_IN_PROGRESS,
			UpdatedAt: ptypes.TimestampNow(),
		},
	}

	if !mk.ManualExecution {
		err := mk.MockComplete(invocationId)
		if err != nil {
			panic(err)
		}
	}

	return invocationId, nil
}

// Manually completes an existing invocation
func (mk *MockRuntimeEnv) MockComplete(fnInvocationId string) error {
	invocation, ok := mk.Results[fnInvocationId]
	if !ok {
		return fmt.Errorf("Could not invoke unknown invocation '%s'", fnInvocationId)
	}

	fnName := invocation.Spec.GetType().GetResolved()
	fn, ok := mk.Functions[fnName]
	if !ok {
		return fmt.Errorf("Could not invoke unknown function '%s'", fnName)
	}

	result, err := fn(invocation.Spec)
	if err != nil {
		logrus.Infof("Function '%s' invocation resulted in an error: %v", fnName, err)
		mk.Results[fnInvocationId].Status = &types.FunctionInvocationStatus{
			Output:    nil,
			UpdatedAt: ptypes.TimestampNow(),
			Status:    types.FunctionInvocationStatus_FAILED,
		}
	} else {
		mk.Results[fnInvocationId].Status = &types.FunctionInvocationStatus{
			Output:    result,
			UpdatedAt: ptypes.TimestampNow(),
			Status:    types.FunctionInvocationStatus_SUCCEEDED,
		}
	}

	return nil
}

func (mk *MockRuntimeEnv) Invoke(spec *types.FunctionInvocationSpec) (*types.FunctionInvocationStatus, error) {
	logrus.Info("Starting invocation...")
	invocationId, err := mk.InvokeAsync(spec)
	if err != nil {
		return nil, err
	}
	err = mk.MockComplete(invocationId)
	if err != nil {
		return nil, err
	}

	logrus.Infof("...completing function execution '%v'", mk.Results)
	return mk.Status(invocationId)
}

func (mk *MockRuntimeEnv) Cancel(fnInvocationId string) error {
	invocation, ok := mk.Results[fnInvocationId]
	if !ok {
		return fmt.Errorf("Could not invoke unknown invocation '%s'", fnInvocationId)
	}

	invocation.Status = &types.FunctionInvocationStatus{
		Output:    nil,
		UpdatedAt: ptypes.TimestampNow(),
		Status:    types.FunctionInvocationStatus_ABORTED,
	}

	return nil
}

func (mk *MockRuntimeEnv) Status(fnInvocationId string) (*types.FunctionInvocationStatus, error) {
	invocation, ok := mk.Results[fnInvocationId]
	if !ok {
		return nil, fmt.Errorf("Could not invoke unknown invocation '%s'", fnInvocationId)
	}

	return invocation.Status, nil
}

type MockFunctionResolver struct {
	FnNameIds map[string]string
}

func (mf *MockFunctionResolver) Resolve(fnName string) (string, error) {
	fnId, ok := mf.FnNameIds[fnName]
	if !ok {
		return "", fmt.Errorf("Could not resolve function '%s' using resolve-map '%v'", fnName, mf.FnNameIds)
	}

	return fnId, nil
}
