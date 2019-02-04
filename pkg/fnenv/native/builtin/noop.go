package builtin

import (
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
)

const (
	Noop      = "noop"
	NoopInput = types.InputMain
)

/*
FunctionNoop represents a "no operation" task; it does not do anything.
The input it receives in its default key, will be outputted in the output

**Specification**

**input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
default         | no       | *                 | The input to pass to the output.

**output** (*) The output of the default input if provided.

**Example**

```yaml
# ...
NoopExample:
  run: noop
  inputs: foobar
# ...
```

A complete example of this function can be found in the [fortunewhale](../examples/whales/fortunewhale.wf.yaml) example.
*/
type FunctionNoop struct{}

func (fn *FunctionNoop) Invoke(spec *types.TaskInvocationSpec) (*typedvalues.TypedValue, error) {

	var output *typedvalues.TypedValue
	switch len(spec.GetInputs()) {
	case 0:
		output = nil
	default:
		defaultInput, ok := spec.GetInputs()[NoopInput]
		if ok {
			output = defaultInput
			break
		}
	}
	logrus.WithFields(logrus.Fields{
		"invocation": spec.InvocationId,
		"task":       spec.GetTask().ID(),
	}).Infof("[internal://%s] %v", Noop,
		typedvalues.MustUnwrap(output))
	return output, nil
}
