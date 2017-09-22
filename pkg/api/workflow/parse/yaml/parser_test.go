package yaml

import (
	"strings"
	"testing"

	"fmt"

	"github.com/fission/fission-workflow/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {

	data := `
apiversion: 123
output: $.tasks.foo.output
tasks:
  foo:
    run: someSh     # No run specified defaults to 'noop'
    requires:
    - bar
  bar:
    run: bla
    inputs: $.invocation.inputs 		# Inputs takes a map or literal, when inputs are referenced it defaults to the 'default' key
`
	wf, err := Parse(strings.NewReader(strings.TrimSpace(data)))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, wf.Output, "$.tasks.foo.output")
	assert.Equal(t, len(wf.Tasks), 2)
}

func TestTransform(t *testing.T) {

	data := `
apiversion: 123
output: $.tasks.foo.output
tasks:
  nothing: ~
  foo:
    run: someSh
  bar:
    run: bla
    inputs: $.invocation.inputs 		# Inputs takes a map or literal, when inputs are referenced it defaults to the 'default' key
  acme:
    run: abc
    inputs:
      default:
        a: b
`
	wfd, err := Parse(strings.NewReader(strings.TrimSpace(data)))
	if err != nil {
		t.Fatal(err)
	}

	wf, err := Transform(wfd)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, wf.OutputTask, wfd.Output)
	assert.Equal(t, wf.ApiVersion, wfd.ApiVersion)
	for id, task := range wfd.Tasks {
		if len(task.Run) == 0 {
			assert.Equal(t, wf.Tasks[id].FunctionRef, "noop")
		} else {
			assert.Equal(t, wf.Tasks[id].FunctionRef, task.Run)
		}

		expectedInputLength := 1
		switch ty := task.Inputs.(type) {
		case nil:
			expectedInputLength = 0
		case map[string]interface{}:
			expectedInputLength = len(ty)
		}
		assert.Equal(t, len(wf.Tasks[id].Inputs), expectedInputLength, fmt.Sprintf("was: %v", task.Inputs))

		assert.Equal(t, wf.Tasks[id].Id, id)
		assert.Equal(t, len(wf.Tasks[id].Requires), len(task.Requires))
		assert.Equal(t, int(wf.Tasks[id].Await), len(task.Requires))

		//assert.IsType(t, wf.Tasks["acme"].Inputs, map[string]interface{}{})
		//acmeInputs := wf.Tasks["acme"].Inputs.(map[string]interface{})
		acmeDefaultInput := wf.Tasks["acme"].Inputs["default"]
		assert.Equal(t, typedvalues.FormatType(typedvalues.FORMAT_JSON, typedvalues.TYPE_OBJECT), acmeDefaultInput.Type)
		i, err := typedvalues.Format(acmeDefaultInput)
		assert.NoError(t, err)
		assert.Equal(t, i, map[string]interface{}{
			"a": "b",
		})
	}
}
