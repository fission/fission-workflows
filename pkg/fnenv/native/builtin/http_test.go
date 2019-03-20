package builtin

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/golang/protobuf/ptypes"
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

		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		fmt.Fprint(w, string(data))
	}))
	defer ts.Close()

	fn := NewFunctionHTTP()
	body := "body"
	deadline, _ := ptypes.TimestampProto(time.Now().Add(10 * time.Second))
	out, err := fn.Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*typedvalues.TypedValue{
			types.InputMethod: typedvalues.MustWrap(http.MethodPost),
			HttpInputUrl:      typedvalues.MustWrap(ts.URL),
			types.InputMain:   typedvalues.MustWrap(body),
			types.InputHeaders: typedvalues.MustWrap(map[string]interface{}{
				"Foo": "Bar",
			}),
		},
		Deadline: deadline,
	})
	assert.NoError(t, err)
	assert.Equal(t, body, typedvalues.MustUnwrap(out))
}

func TestFunctionHttp_Invoke_Invalid(t *testing.T) {

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "expected error", 500)
	}))
	defer ts.Close()

	fn := NewFunctionHTTP()
	body := "body"
	out, err := fn.Invoke(&types.TaskInvocationSpec{
		Inputs: map[string]*typedvalues.TypedValue{
			types.InputMethod: typedvalues.MustWrap(http.MethodDelete),
			HttpInputUrl:      typedvalues.MustWrap(ts.URL),
			types.InputMain:   typedvalues.MustWrap(body),
			types.InputHeaders: typedvalues.MustWrap(map[string]interface{}{
				"Foo": "Bar",
			}),
		},
	})
	assert.Nil(t, out)
	assert.Error(t, err, "expected error\n")
}
