package builtin

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestFunctionHttp_Invoke(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			panic("incorrect method")
		}

		data, err := ioutil.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		if v, ok := r.Header["Foo"]; !ok || len(v) == 0 || v[0] != "Bar" {
			panic("Header 'Foo: Bar' not present")
		}

		w.Header().Set("Content-Type", httpDefaultContentType)
		fmt.Fprintln(w, string(data))
	}))
	defer ts.Close()

	fn := &FunctionHttp{}
	body := "body"
	out, err := fn.Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*types.TypedValue{
			HttpInputMethod: typedvalues.UnsafeParse(http.MethodPost),
			HttpInputUri:    typedvalues.UnsafeParse(ts.URL),
			HttpInputBody:   typedvalues.UnsafeParse(body),
			HttpInputHeaders: typedvalues.UnsafeParse(map[string]interface{}{
				"Foo": "Bar",
			}),
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, body, typedvalues.UnsafeFormat(out))
}
