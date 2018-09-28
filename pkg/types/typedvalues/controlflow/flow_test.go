package controlflow

import (
	"fmt"
	"testing"
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestIsControlFlow(t *testing.T) {
	assert.True(t, IsControlFlow(typedvalues.MustWrap(&types.WorkflowSpec{})))
	assert.True(t, IsControlFlow(typedvalues.MustWrap(&types.TaskSpec{})))
	assert.False(t, IsControlFlow(typedvalues.MustWrap(nil)))
}

type testCase struct {
	name         string
	input        proto.Message
	expectedType string
}

// parseFormatTestCases provides a suit of test cases.
//
// It is a function instead of variable because of the package initialization sequence.
func parseFormatTestCases() []testCase {
	return []testCase{
		{
			input: &types.TaskSpec{
				FunctionRef: "someFn",
				Inputs: map[string]*typedvalues.TypedValue{
					"foo": typedvalues.MustWrap("bar"),
				},
				Requires: map[string]*types.TaskDependencyParameters{
					"prev": nil,
				},
			},
			expectedType: TypeTask,
		},
		{
			input: &types.WorkflowSpec{
				ApiVersion: types.WorkflowAPIVersion,
				OutputTask: "fakeFinalTask",
				Tasks: map[string]*types.TaskSpec{
					"fakeFinalTask": {
						FunctionRef: "noop",
					},
				},
			},
			expectedType: TypeWorkflow,
		},
		{
			input: FlowTask(&types.TaskSpec{
				FunctionRef: "someFn",
				Inputs: map[string]*typedvalues.TypedValue{
					"foo": typedvalues.MustWrap("bar"),
				},
				Requires: map[string]*types.TaskDependencyParameters{
					"prev": nil,
				},
			}),
			expectedType: TypeFlow,
		},
		{
			input: FlowWorkflow(&types.WorkflowSpec{
				ApiVersion: types.WorkflowAPIVersion,
				OutputTask: "fakeFinalTask",
				Tasks: map[string]*types.TaskSpec{
					"fakeFinalTask": {
						FunctionRef: "noop",
					},
				},
			}),
			expectedType: TypeFlow,
		},
	}
}

func TestValueTester(t *testing.T) {
	var i int
	for _, testCase := range parseFormatTestCases() {
		testName := testCase.name
		if len(testName) == 0 {
			testName = fmt.Sprintf("%d_%v", i, testCase.expectedType)
		}
		t.Run(testName, func(t *testing.T) {
			fmt.Printf("Input: %+v\n", testCase)
			tv, err := typedvalues.Wrap(testCase.input)
			fmt.Printf("Typed value: %+v\n", tv)
			assert.NoError(t, err)
			assert.True(t, IsControlFlow(tv))
			assert.Equal(t, testCase.expectedType, tv.ValueType())
			i, err := typedvalues.Unwrap(tv)
			assert.NoError(t, err)
			util.AssertProtoEqual(t, testCase.input, i.(proto.Message))
			fmt.Printf("Output: %+v\n", i)
		})
		i++
	}
	time.Sleep(100 * time.Millisecond)
}

func BenchmarkParse(b *testing.B) {
	for _, testCase := range parseFormatTestCases() {
		b.Run(testCase.expectedType+"_parse", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				typedvalues.Wrap(testCase.input)
			}
		})
	}
	for _, testCase := range parseFormatTestCases() {
		tv, _ := typedvalues.Wrap(testCase.input)

		b.Run(testCase.expectedType+"_format", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				typedvalues.Unwrap(tv)
			}
		})
	}
}
