package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var parseCases = map[string]struct {
	FnRef
	err  error
	full string
}{
	"noRuntime":                         {NewFnRef("", "", "noRuntime"), nil, "noRuntime"},
	"a://b":                             {NewFnRef("a", "", "b"), nil, "a://b"},
	"http://foobar":                     {NewFnRef("http", "", "foobar"), nil, "http://foobar"},
	"fission://fission-function/foobar": {NewFnRef("fission", "fission-function", "foobar"), nil, "fission://fission-function/foobar"},

	"":             {FnRef{}, ErrInvalidFnRef, ""},
	"://":          {FnRef{}, ErrInvalidFnRef, ""},
	"://runtimeId": {FnRef{}, ErrInvalidFnRef, ""},
	"runtime://":   {FnRef{}, ErrInvalidFnRef, ""},
}

func TestParse(t *testing.T) {
	for input, expected := range parseCases {
		t.Run(input, func(t *testing.T) {
			ref, err := ParseFnRef(input)
			assert.Equal(t, expected.err, err)
			assert.EqualValues(t, expected.FnRef, ref)
		})
	}
}

func TestFnRef_Format(t *testing.T) {
	for expected, input := range parseCases {
		t.Run(expected, func(t *testing.T) {
			fnRef := input.FnRef
			if !fnRef.IsEmpty() {
				assert.Equal(t, input.full, fnRef.Format())
			}
		})
	}
}
