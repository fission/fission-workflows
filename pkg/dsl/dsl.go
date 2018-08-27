package dsl

import (
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
)

// TODO validation (no functions as inputs)

type Inputs = map[string]interface{}

type Flow interface {
	Build() *types.WorkflowSpec
}

// Data structures
type WorkflowSpec struct {
	OutputTask string
	Tasks      map[string]*TaskSpec
}

func (wf *WorkflowSpec) Build() *types.WorkflowSpec {
	tasks := map[string]*types.TaskSpec{}
	for taskID, spec := range wf.Tasks {
		tasks[taskID] = spec.BuildTask()
	}
	return &types.WorkflowSpec{
		ApiVersion: types.WorkflowAPIVersion,
		Tasks:      tasks,
	}
}

type TaskSpec struct {
	ID       string // TODO generate if not set
	Run      string
	Inputs   map[string]interface{} // TODO to value
	Requires []*TaskSpec
}

func (ts *TaskSpec) OnComplete(tasks ...*TaskSpec) *TaskSpec {
	for _, task := range tasks {
		task.Require(ts)
	}
	return ts
}

func (ts *TaskSpec) Require(tasks ...*TaskSpec) *TaskSpec {
	ts.Requires = append(ts.Requires, tasks...)
	return ts
}

func (ts *TaskSpec) Input(key string, value interface{}) *TaskSpec {
	ts.Inputs[key] = value
	return ts
}

// TODO gather dependencies
func (ts *TaskSpec) Build() *types.WorkflowSpec {
	return (&WorkflowSpec{
		OutputTask: "main",
		Tasks: map[string]*TaskSpec{
			"main": ts,
		},
	}).Build()
}

func (ts *TaskSpec) BuildTask() *types.TaskSpec {
	// TODO support TaskSpec in inputs
	inputs, err := typedvalues.ParseToTypedValueMap(typedvalues.DefaultParserFormatter, ts.Inputs)
	if err != nil {
		panic(err)
	}
	deps := map[string]*types.TaskDependencyParameters{}
	for _, dep := range ts.Requires {
		deps[dep.ID] = nil
	}
	return &types.TaskSpec{
		FunctionRef: ts.Run,
		Inputs:      inputs,
		Requires:    deps,
	}
}

// Standard Library. Should be moved to a separate util package
// TODO validate / parse expression
func If(expr interface{}, then Flow, orElse Flow) *TaskSpec {
	return &TaskSpec{
		Run: "native://if",
		Inputs: map[string]interface{}{
			"if":   expr,
			"then": then,
			"else": orElse,
		},
	}
}

func Fail(reason string) *TaskSpec {
	return &TaskSpec{
		Run: "native://fail",
		Inputs: map[string]interface{}{
			types.InputMain: reason,
		},
	}
}

func Noop() *TaskSpec {
	return &TaskSpec{
		Run: "native://noop",
	}
}

func Sleep(d time.Duration) *TaskSpec {
	return &TaskSpec{
		Run: "native://sleep",
		Inputs: map[string]interface{}{
			types.InputMain: d.String(),
		},
	}
}

// Experimental
func Expression(expr string) string {
	return "{" + expr + "}"
}

// Ideas:
// - Runtime() / Promise() function on tasks for referencing in and outputs

// Idea for input values as opposed to untyped interface{}. Implement basic types (number, string, bytes...)
type Value interface {
	Type() string
	Value() interface{}
	Raw() []byte
}
