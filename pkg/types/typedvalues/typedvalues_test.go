package typedvalues

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name         string
	input        interface{}
	expectedType string
}

// parseFormatTestCases provides a suit of test cases.
//
// It is a function instead of variable because of the package initialization sequence.
func parseFormatTestCases() []testCase {
	return []testCase{
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
			expectedType: TypeFloat64,
		},
		{
			input:        float64(42),
			expectedType: TypeFloat64,
		},
		{
			input:        []byte("foo bar"),
			expectedType: TypeBytes,
		},
		{
			input:        []byte(nil),
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
			tv, err := Wrap(testCase.input)
			fmt.Printf("Typed value: %+v\n", tv)
			assert.NoError(t, err)
			assert.Equal(t, testCase.expectedType, tv.ValueType())
			i, err := Unwrap(tv)
			assert.NoError(t, err)
			assert.Equal(t, testCase.input, i)
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
				Wrap(testCase.input)
			}
		})
	}
	for _, testCase := range parseFormatTestCases() {
		tv, _ := Wrap(testCase.input)

		b.Run(testCase.expectedType+"_format", func(b *testing.B) {
			for n := 0; n < b.N; n++ {
				Unwrap(tv)
			}
		})
	}
}
