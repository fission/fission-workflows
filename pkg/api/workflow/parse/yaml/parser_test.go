package yaml

import (
	"strings"
	"testing"

	"fmt"

	"github.com/magiconair/properties/assert"
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
			assert.Equal(t, wf.Tasks[id].Name, "noop")
		} else {
			assert.Equal(t, wf.Tasks[id].Name, task.Run)
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
		assert.Equal(t, len(wf.Tasks[id].Dependencies), len(task.Requires))
		assert.Equal(t, int(wf.Tasks[id].DependenciesAwait), len(task.Requires))
	}
}
