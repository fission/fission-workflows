package yaml

import (
	"errors"

	"io"
	"io/ioutil"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
	"gopkg.in/yaml.v2"
)

const (
	DEFAULT_FUNCTIONREF = "noop"
)

func Parse(r io.Reader) (*WorkflowDef, error) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	i := &WorkflowDef{}
	err = yaml.Unmarshal(bs, i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func Transform(def *WorkflowDef) (*types.WorkflowSpec, error) {

	tasks := map[string]*types.Task{}

	for id, task := range def.Tasks {
		p, err := parseTask(id, task)
		if err != nil {
			return nil, err
		}
		tasks[id] = p
	}

	return &types.WorkflowSpec{
		ApiVersion: def.ApiVersion,
		OutputTask: def.Output,
		Tasks:      tasks,
	}, nil
}

func parseInput(i interface{}) (map[string]*types.TypedValue, error) {
	if i == nil {
		return map[string]*types.TypedValue{}, nil
	}

	switch v := i.(type) {
	case map[string]interface{}:
		result := map[string]*types.TypedValue{}
		for inputKey, inputVal := range v {

			typedVal, err := parseSingleInput(inputVal)
			if err != nil {
				return nil, err
			}
			result[inputKey] = typedVal
		}
		return result, nil
	}
	p, err := parseSingleInput(i)
	if err != nil {
		return nil, err
	}
	return map[string]*types.TypedValue{
		types.INPUT_MAIN: p,
	}, nil
}

func parseSingleInput(i interface{}) (*types.TypedValue, error) {
	switch v := i.(type) {
	case map[string]interface{}:
		return nil, errors.New("unexpected collection")
	case *TaskDef: // Handle TaskDef because it cannot be parsed by standard parser
		p, err := parseTask("", v)
		if err != nil {
			return nil, err
		}
		i = p
	}
	p, err := typedvalues.Parse(i)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func parseTask(id string, t *TaskDef) (*types.Task, error) {
	deps := map[string]*types.TaskDependencyParameters{}
	for _, dep := range t.Requires {
		deps[dep] = &types.TaskDependencyParameters{}
	}

	if len(id) == 0 {
		id = t.Id
	}

	inputs, err := parseInput(t.Inputs)
	if err != nil {
		return nil, err
	}

	fn := t.Run
	if len(fn) == 0 {
		fn = DEFAULT_FUNCTIONREF
	}

	return &types.Task{
		Id:          id,
		FunctionRef: fn,
		Requires:    deps,
		Await:       int32(len(deps)),
		Inputs:      inputs,
	}, nil
}

type WorkflowDef struct {
	ApiVersion string
	Output     string
	Tasks      map[string]*TaskDef
}

type TaskDef struct {
	Id       string
	Run      string
	Inputs   interface{}
	Requires []string
}
