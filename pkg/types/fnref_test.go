package types

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var parseCases = map[string]struct {
	FnRef
	err error
}{
	"":              {FnRef{}, ErrInvalidFnRef},
	"noRuntime":     {NewFnRef("", "noRuntime"), ErrNoRuntime},
	"://":           {FnRef{}, ErrInvalidFnRef},
	"://runtimeId":  {FnRef{}, ErrInvalidFnRef},
	"runtime://":    {FnRef{}, ErrInvalidFnRef},
	"a://b":         {NewFnRef("a", "b"), nil},
	"http://foobar": {NewFnRef("http", "foobar"), nil},
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
				assert.Equal(t, expected, fnRef.Format())
			}
		})
	}
}
