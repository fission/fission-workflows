package typedvalues

import (
	"fmt"
	"testing"
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestValueTester(t *testing.T) {
	var i int
	for _, testCase := range parseFormatTests {
		testName := testCase.name
		if len(testName) == 0 {
			testName = fmt.Sprintf("%d_%v", i, testCase.expectedType)
		}
		t.Run(testName, func(t *testing.T) {
			fmt.Printf("Input: %+v\n", testCase)
			tv, err := Parse(testCase.input)
			fmt.Printf("Typed value: %+v\n", tv)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedType, ValueType(tv.Type))
			i, err := Format(tv)
			assert.NoError(t, err)
			assert.Equal(t, testCase.input, i)
			fmt.Printf("Output: %+v\n", i)
		})
		i++
	}
	time.Sleep(100 * time.Millisecond)
}

type testCase struct {
	name         string
	input        interface{}
	expectedType ValueType
}

var parseFormatTests = []testCase{
	{
		input:        nil,
		expectedType: TypeNil,
	},
	{
		input:        true,
		expectedType: TypeBool,
	},
	{
		input:        false,
		expectedType: TypeBool,
	},
	{
		input:        float64(0),
		expectedType: TypeNumber,
	},
	{
		input:        float64(42),
		expectedType: TypeNumber,
	},
	{
		input:        []byte("foo bar"),
		expectedType: TypeBytes,
	},
	{
		input:        []byte{},
		expectedType: TypeBytes,
	},
	{
		input:        "",
		expectedType: TypeString,
	},
	{
		input:        "foo bar",
		expectedType: TypeString,
	},
	{
		input:        "{",
		expectedType: TypeString,
	},
	{
		input:        "{foo}",
		expectedType: TypeExpression,
	},
	{
		input:        "{}",
		expectedType: TypeExpression,
	},
	{
		input:        []interface{}{},
		expectedType: TypeList,
	},
	{
		input:        []interface{}{float64(42), "foo"},
		expectedType: TypeList,
	},
	{
		input:        map[string]interface{}{"foo": float64(42), "bar": true},
		expectedType: TypeMap,
	},
	{
		input:        map[string]interface{}{},
		expectedType: TypeMap,
	},
	{
		// Complex
		name:         "recursiveList",
		input:        []interface{}{[]interface{}{[]interface{}{"foo"}}},
		expectedType: TypeList,
	},
	{
		// Complex
		name:         "recursiveMap",
		input:        map[string]interface{}{"a": map[string]interface{}{"b": map[string]interface{}{"c": "{d}"}}},
		expectedType: TypeMap,
	},
	{
		input:        types.NewTaskSpec("fn1").Input("inputK", MustParse("e2")),
		expectedType: TypeTask,
	},
	{
		input: &types.WorkflowSpec{
			ApiVersion: "v1",
			OutputTask: "t1",
			Tasks: types.Tasks{
				"t1": types.NewTaskSpec("fn1").Input("inputK", MustParse("e2")),
			},
		},
		expectedType: TypeWorkflow,
	},
}
