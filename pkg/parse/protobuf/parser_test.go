package protobuf

import (
	"bytes"
	"strings"
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestParseProto(t *testing.T) {
	originalWfSpec := &types.WorkflowSpec{
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
				},
			},
		},
	}
	msg, err := proto.Marshal(originalWfSpec)
	assert.NoError(t, err)
	parsedWfSpec, err := Parse(bytes.NewReader(msg))
	assert.NoError(t, err)
	util.AssertProtoEqual(t, originalWfSpec, parsedWfSpec)
}

func TestParseJson(t *testing.T) {
	originalWfSpec := &types.WorkflowSpec{
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
				},
			},
		},
	}
	msg, err := (&jsonpb.Marshaler{}).MarshalToString(originalWfSpec)
	assert.NoError(t, err)
	parsedWfSpec, err := Parse(strings.NewReader(msg))
	assert.NoError(t, err)
	if !proto.Equal(originalWfSpec, parsedWfSpec) {
		assert.Fail(t, "not equal")
	}
}
