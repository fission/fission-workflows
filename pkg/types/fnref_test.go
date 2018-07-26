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
	"noRuntime":                         {NewFnRef("", "default", "noRuntime"), nil, "default/noRuntime"},
	"a://b":                             {NewFnRef("a", "default", "b"), nil, "a://default/b"},
	"http://foobar":                     {NewFnRef("http", "default", "foobar"), nil, "http://default/foobar"},
	"fission://fission-function/foobar": {NewFnRef("fission", "fission-function", "foobar"), nil, "fission://fission-function/foobar"},

	"":             {FnRef{}, ErrInvalidFnRef, ""},
	"://":          {FnRef{}, ErrInvalidFnRef, ""},
	"://runtimeId": {FnRef{}, ErrInvalidFnRef, ""},
	"runtime://":   {FnRef{}, ErrInvalidFnRef, ""},
}

func TestParse(t *testing.T) {
	for input, expected := range parseCases {
		t.Run(input, func(t *testing.T) {
			valid := IsFnRef(input)
			ref, err := ParseFnRef(input)
			assert.Equal(t, expected.err == nil, valid)
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
