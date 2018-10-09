package fission

import (
	"context"
	"io/ioutil"
	"testing"

	"github.com/fission/fission"
	"github.com/fission/fission-workflows/pkg/apiserver"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/ptypes/empty"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	k8stypes "k8s.io/apimachinery/pkg/types"
)

type mockWorkflowClient struct {
	mock.Mock
}

func (m *mockWorkflowClient) Create(ctx context.Context, in *types.WorkflowSpec, opts ...grpc.CallOption) (*types.ObjectMetadata, error) {
	args := m.Called(in)
	return &types.ObjectMetadata{Id: args.String(0)}, args.Error(1)
}

func (m *mockWorkflowClient) List(ctx context.Context, _ *empty.Empty, opts ...grpc.CallOption) (*apiserver.WorkflowList, error) {
	args := m.Called()
	return args.Get(0).(*apiserver.WorkflowList), args.Error(1)
}

func (m *mockWorkflowClient) Get(ctx context.Context, in *types.ObjectMetadata, opts ...grpc.CallOption) (*types.Workflow, error) {
	args := m.Called(in)
	return args.Get(0).(*types.Workflow), args.Error(1)
}

func (m *mockWorkflowClient) Delete(ctx context.Context, in *types.ObjectMetadata, opts ...grpc.CallOption) (*empty.Empty, error) {
	args := m.Called(in)
	return &empty.Empty{}, args.Error(1)
}

func (m *mockWorkflowClient) Validate(ctx context.Context, in *types.WorkflowSpec, opts ...grpc.CallOption) (*empty.Empty, error) {
	args := m.Called(in)
	return &empty.Empty{}, args.Error(1)
}

func TestProxy_Specialize(t *testing.T) {
	workflowServer := &mockWorkflowClient{}
	workflowServer.On("Create", mock.Anything).Return("mockID", nil)
	env := NewEnvironmentProxyServer(nil, workflowServer)
	wf := &types.WorkflowSpec{
		ApiVersion: types.WorkflowAPIVersion,
		OutputTask: "fakeFinalTask",
		Tasks: map[string]*types.TaskSpec{
			"fakeFinalTask": {
				FunctionRef: "noop",
				Inputs: map[string]*typedvalues.TypedValue{
					types.InputMain: typedvalues.MustWrap("{$.Tasks.FirstTask.Output}"),
				},
				Requires: map[string]*types.TaskDependencyParameters{
					"FirstTask": {},
				},
			},
			"FirstTask": {
				FunctionRef: "noop",
				Inputs: map[string]*typedvalues.TypedValue{
					types.InputMain: typedvalues.MustWrap("{$.Invocation.Inputs.default.toUpperCase()}"),
					"complex": typedvalues.MustWrap(map[string]interface{}{
						"nested": map[string]interface{}{
							"object": 42,
						},
					}),
				},
			},
		},
	}

	// Store workflow in a temporary file (akin to fetcher request)
	fd, err := ioutil.TempFile("", "test-fission-workflows-envproxy")
	assert.NoError(t, err)
	err = (&jsonpb.Marshaler{}).Marshal(fd, wf)
	assert.NoError(t, err)
	fd.Close()

	wfIds, err := env.specialize(context.Background(), &fission.FunctionLoadRequest{
		FilePath: fd.Name(),
		FunctionMetadata: &v1.ObjectMeta{
			UID:  k8stypes.UID("1"),
			Name: "testFn",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, 1, len(wfIds))
	mock.AssertExpectationsForObjects(t, workflowServer)
}
