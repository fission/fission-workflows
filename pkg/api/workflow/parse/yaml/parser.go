package yaml

import (
	"io"
	"io/ioutil"

	"encoding/json"
	"fmt"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
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
	case map[interface{}]interface{}:
		//case map[string]interface{}:
		result := map[string]*types.TypedValue{}
		for inputKey, inputVal := range v {
			k := fmt.Sprintf("%v", inputKey)
			typedVal, err := parseSingleInput(k, inputVal)
			if err != nil {
				return nil, err
			}
			result[k] = typedVal
		}
		return result, nil
	}
	p, err := parseSingleInput("default", i)
	if err != nil {
		return nil, err
	}
	return map[string]*types.TypedValue{
		types.INPUT_MAIN: p,
	}, nil
}

func parseSingleInput(id string, i interface{}) (*types.TypedValue, error) {
	// Handle special cases
	switch t := i.(type) {
	case map[interface{}]interface{}:
		res := map[string]interface{}{}
		for k, v := range t {
			res[fmt.Sprintf("%v", k)] = v
		}
		bs, err := json.Marshal(res)
		if err != nil {
			panic(err)
		}
		td := &TaskDef{}
		err = json.Unmarshal(bs, td)
		if err != nil {
			panic(err)
		}
		p, err := parseTask(id, td)
		if err != nil {
			i = res
		} else {
			i = p
		}
	case *TaskDef: // Handle TaskDef because it cannot be parsed by standard parser
		p, err := parseTask(id, t)
		if err != nil {
			return nil, err
		}
		i = p
	}

	p, err := typedvalues.Parse(i)
	if err != nil {
		return nil, err
	}

	logrus.WithField("in", i).WithField("out", p).Infof("parsed input")
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

	result := &types.Task{
		Id:          id,
		FunctionRef: fn,
		Requires:    deps,
		Await:       int32(len(deps)),
		Inputs:      inputs,
	}

	logrus.WithField("id", id).WithField("in", t).WithField("out", result).Infof("parsed task")
	return result, nil
}

//
// YAML data structures
//

type WorkflowDef struct {
	ApiVersion  string
	Description string
	Output      string
	Tasks       map[string]*TaskDef
}

type TaskDef struct {
	Id       string
	Run      string
	Inputs   interface{}
	Requires []string
}
