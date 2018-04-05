package builtin

import (
	"fmt"
	"strconv"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/gogo/protobuf/proto"
)

const (
	Repeat           = "repeat"
	RepeatInputTimes = "times"
	RepeatInputDo    = "do"
	RepeatInputPrev  = "prev"
)

// TODO minor: chose between unrolled loop and dynamic loop based on number of tasks for performance
type FunctionRepeat struct{}

func (fn *FunctionRepeat) Invoke(spec *types.TaskInvocationSpec) (*types.TypedValue, error) {

	n, ok := spec.Inputs[RepeatInputTimes]
	if !ok {
		return nil, fmt.Errorf("repeat needs '%s'", RepeatInputTimes)
	}

	// Parse condition to a int
	i, err := typedvalues.Format(n)
	if err != nil {
		return nil, err
	}

	// TODO fix int typedvalue
	var times int64
	f, ok := i.(float64)
	if ok {
		times = int64(f)
	} else {
		// Fallback: attempt to convert string to int
		t, err := strconv.Atoi(fmt.Sprintf("%s", i))
		if err != nil {
			return nil, fmt.Errorf("condition '%s' needs to be a 'int64', but was '%T'", i, i)
		}
		times = int64(t)
	}

	// Parse do task
	// TODO does a workflow also work?
	doVal, ok := spec.Inputs[RepeatInputDo]
	if !ok {
		return nil, fmt.Errorf("repeat needs '%s'", RepeatInputDo)
	}
	doTask, err := typedvalues.FormatTask(doVal)
	if err != nil {
		return nil, err
	}
	doTask.Requires = map[string]*types.TaskDependencyParameters{}

	if times > 0 {
		// TODO add context
		return typedvalues.UnsafeParse(&types.WorkflowSpec{
			OutputTask: taskId(times - 1),
			Tasks:      createRepeatTasks(doTask, times),
		}), nil
	} else {
		return nil, nil
	}
}

func createRepeatTasks(task *types.TaskSpec, times int64) map[string]*types.TaskSpec {
	tasks := map[string]*types.TaskSpec{}

	for n := int64(0); n < times; n += 1 {
		id := taskId(n)
		do := proto.Clone(task).(*types.TaskSpec)
		if n > 0 {
			prev := taskId(n - 1)
			do.Require(prev)
			do.Input(RepeatInputPrev, typedvalues.UnsafeParse(fmt.Sprintf("{output(%s)}", prev)))
		}
		tasks[id] = do
	}

	return tasks
}

func taskId(n int64) string {
	return fmt.Sprintf("do_%d", n)
}
