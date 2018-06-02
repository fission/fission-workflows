package yaml

import (
	"strings"
	"testing"

	"fmt"

	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestRead(t *testing.T) {

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
	wf, err := read(strings.NewReader(strings.TrimSpace(data)))
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, wf.Output, "$.tasks.foo.output")
	assert.Equal(t, len(wf.Tasks), 2)
}

func TestParseSimple(t *testing.T) {
	data := `
apiversion: 123
output: $.tasks.foo.output
tasks:
  foo:
    run: someSh
  bar:
    run: bla
    inputs: $.invocation.inputs
  acme:
    run: abc
    inputs:
      default:
        a: b
`
	wfd, err := read(strings.NewReader(strings.TrimSpace(data)))
	if err != nil {
		t.Fatal(err)
	}

	wf, err := parseWorkflow(wfd)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, wf.OutputTask, wfd.Output)
	assert.Equal(t, wf.ApiVersion, wfd.APIVersion)
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

		assert.Equal(t, len(wf.Tasks[id].Requires), len(task.Requires))
		assert.Equal(t, int(wf.Tasks[id].Await), len(task.Requires))

		acmeDefaultInput := wf.Tasks["acme"].Inputs["default"]
		assert.Equal(t, typedvalues.TypeMap, acmeDefaultInput.Type)
		i, err := typedvalues.Format(acmeDefaultInput)
		assert.NoError(t, err)
		assert.Equal(t, i, map[string]interface{}{
			"a": "b",
		})
	}
}

func TestParseDynamicWorkflow(t *testing.T) {

	data := `
apiversion: 123
output: $.tasks.foo.output
tasks:
  nothing: ~
  foo:
    run: someSh
  bar:
    run: bla
    inputs: 
      default:
        apiversion: v42
        output: $.tasks.inner.dynamic
        tasks:
          inner:
            run: dynamic
            inputs: foobar
`

	wf, err := Parse(strings.NewReader(data))
	assert.NoError(t, err)
	barTask, ok := wf.Tasks["bar"]
	assert.True(t, ok)
	wfInput, ok := barTask.Inputs["default"]
	assert.True(t, ok)
	innerWf, err := typedvalues.FormatWorkflow(wfInput)
	assert.NoError(t, err)
	assert.Equal(t, "$.tasks.inner.dynamic", innerWf.OutputTask)
	assert.Equal(t, "v42", innerWf.ApiVersion)
	assert.Equal(t, "foobar", typedvalues.MustFormat(innerWf.Tasks["inner"].Inputs["default"]))
	assert.Equal(t, "dynamic", innerWf.Tasks["inner"].FunctionRef)
}

func TestParseWorkflowWithArray(t *testing.T) {

	data := `
tasks:
  taskWithArray:
    run: bla
    inputs:
      cases:
      - case: a
        value: b
      - case: b
        value: c
`

	wf, err := Parse(strings.NewReader(data))
	assert.NoError(t, err)
	assert.NotNil(t, wf)
}

func TestParseWorkflowWithMap(t *testing.T) {

	data := `
tasks:
  taskWithArray:
    run: bla
    inputs:
      cases:
        a: b
        c: d
`

	wf, err := Parse(strings.NewReader(data))
	assert.NoError(t, err)
	assert.NotNil(t, wf)
}
