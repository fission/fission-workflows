package builtin

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
)

const (
	Fail         = "fail"
	FailInputMsg = types.InputMain
)

var defaultErrMsg = typedvalues.MustWrap("fail function triggered")

/*
FunctionFail is a function that always fails. This can be used to short-circuit workflows in
specific branches. Optionally you can provide a custom message to the failure.

**Specification**

**input**   | required | types  | description
------------|----------|--------|---------------------------------
default     | no       | string | custom message to show on error

**output** None

**Example**

```yaml
# ...
foo:
  run: fail
  inputs: "all has failed"
# ...
```

A runnable example of this function can be found in the [failwhale](../examples/whales/failwhale.wf.yaml) example.
*/
type FunctionFail struct{}

func (fn *FunctionFail) Invoke(spec *types.TaskInvocationSpec) (*typedvalues.TypedValue, error) {
	var output *typedvalues.TypedValue
	switch len(spec.GetInputs()) {
	case 0:
		output = defaultErrMsg
	default:
		defaultInput, ok := spec.GetInputs()[FailInputMsg]
		if ok {
			output = defaultInput
			break
		}
	}
	logrus.WithFields(logrus.Fields{
		"spec":   spec,
		"output": output,
	}).Info("Internal Fail-function invoked.")

	msg, err := typedvalues.Unwrap(output)
	if err != nil {
		return nil, err
	}

	return nil, fmt.Errorf("%v", msg)
}
