package mediatype

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	input  string
	valid  bool
	parsed *MediaType
}

var testCases = []testCase{
	{ // simple
		input: "application/json",
		valid: true,
		parsed: &MediaType{
			Type:       "application",
			Subtype:    "json",
			Parameters: map[string]string{},
		},
	},
	{ // single tag
		input: "text/plain; proto=org.some.Message",
		valid: true,
		parsed: &MediaType{
			Type:    "text",
			Subtype: "plain",
			Parameters: map[string]string{
				"proto": "org.some.Message",
			},
		},
	},
	{ // media type with suffix
		input: "application/protobuf+json; charset=utf-8",
		valid: true,
		parsed: &MediaType{
			Type:    "application",
			Subtype: "protobuf",
			Suffix:  "json",
			Parameters: map[string]string{
				"charset": "utf-8",
			},
		},
	},
	{ // multiple tags
		input: "application/vnd.google.protobuf; proto=com.example.SomeMessage; zoo=bar",
		valid: true,
		parsed: &MediaType{
			Type:    "application",
			Subtype: "vnd.google.protobuf",
			Parameters: map[string]string{
				"proto": "com.example.SomeMessage",
				"zoo":   "bar",
			},
		},
	},
}

func TestParse(t *testing.T) {
	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("test_%d", i), func(t *testing.T) {
			// Test parsing
			mt, err := Parse(testCase.input)
			if !testCase.valid {
				assert.Error(t, err)
				assert.Nil(t, mt)
				return
			}
			assert.NoError(t, err)
			assert.EqualValues(t, testCase.parsed, mt)

			// Test formatting (assuming that the input is correctly formatted)
			assert.Equal(t, testCase.input, testCase.parsed.String())
			assert.Equal(t, testCase.input, mt.String())
		})
	}
}
