package builtin

import (
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

const (
	Sleep             = "sleep"
	SleepInput        = types.InputMain
	SleepInputDefault = time.Duration(1) * time.Second
)

/*
FunctionSleep is similarly to `noop` a "no operation" function.
However, the `sleep` function will wait for a specific amount of time before "completing".
This can be useful to mock or stub out functions during development, while simulating the realistic execution time.

**Specification**

**input**       | required | types             | description
----------------|----------|-------------------|--------------------------------------------------------
default         | no       | string            | A string-based representation of the duration of the sleep. (default: 1 second)

Note: the sleep input is parsed based on the [Golang Duration string notation](https://golang.org/pkg/time/#ParseDuration).
Examples: 1 hour and 10 minutes: `1h10m`, 2 minutes and 300 milliseconds: `2m300ms`.

**output** None

**Example**

```yaml
# ...
NoopExample:
  run: sleep
  inputs: 1h
# ...
```

A complete example of this function can be found in the [sleepalot](../examples/misc/sleepalot.wf.yaml) example.
*/
type FunctionSleep struct{}

func (f *FunctionSleep) Invoke(spec *types.TaskInvocationSpec) (*typedvalues.TypedValue, error) {
	duration := SleepInputDefault
	input, ok := spec.Inputs[SleepInput]
	if ok {
		i, err := typedvalues.Format(input)
		if err != nil {
			return nil, err
		}

		switch t := i.(type) {
		case string:
			d, err := time.ParseDuration(t)
			if err != nil {
				return nil, err
			}
			duration = d
		case float64:
			duration = time.Duration(t) * time.Millisecond
		default:
			return nil, fmt.Errorf("invalid input '%v'", input.Type)
		}
	}

	time.Sleep(duration)

	return nil, nil
}
