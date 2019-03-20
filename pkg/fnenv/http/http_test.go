package http

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestRuntime_ResolveValid(t *testing.T) {
	runtime := New()
	for input, expected := range map[string]string{
		"http://foo.bar":             "http://foo.bar",
		"https://foo.bar":            "https://foo.bar",
		"https://foo.bar/":           "https://foo.bar",
		"https://foo.bar/a/b":        "https://foo.bar/a/b",
		"https://foo.bar/fn?ac=me":   "https://foo.bar/fn",
		"https://foo.bar/fn#hashval": "https://foo.bar/fn",
	} {
		t.Run(input, func(t *testing.T) {
			fnref, err := types.ParseFnRef(input)
			assert.NoError(t, err)
			fnID, err := runtime.Resolve(fnref)
			assert.NoError(t, err)
			assert.Equal(t, expected, fnID)
		})
	}
}

func TestRuntime_ResolveInvalid(t *testing.T) {
	runtime := New()
	for input, expected := range map[*types.FnRef]error{
		{Runtime: "", Namespace: "", ID: ""}:          types.ErrFnRefNoID,
		{Runtime: "http", Namespace: "", ID: ""}:      types.ErrFnRefNoID,
		{Runtime: "nohttp", Namespace: "", ID: "foo"}: ErrUnsupportedScheme,
		{Runtime: "", Namespace: "ns", ID: ""}:        types.ErrFnRefNoID,
	} {
		fnref := input.Format()
		t.Run(fnref, func(t *testing.T) {
			_, err := runtime.Resolve(*input)
			assert.EqualError(t, err, expected.Error())
		})
	}
}
