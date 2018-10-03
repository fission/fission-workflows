package yaml

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/fission/fission-workflows/pkg/fnenv/native/builtin"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

const defaultFunctionRef = builtin.Noop

var DefaultParser = &Parser{}

func Parse(r io.Reader) (*types.WorkflowSpec, error) {
	return DefaultParser.Parse(r)
}

type Parser struct{}

func (p *Parser) Parse(r io.Reader) (*types.WorkflowSpec, error) {
	b, err := read(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow definition: %v", err)
	}

	spec, err := parseWorkflow(b)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow definition: %v", err)
	}

	return spec, nil
}

func read(r io.Reader) (*workflowSpec, error) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	i := &workflowSpec{}
	err = yaml.Unmarshal(bs, i)
	if err != nil {
		return nil, err
	}
	return i, nil
}

func parseWorkflow(def *workflowSpec) (*types.WorkflowSpec, error) {

	tasks := map[string]*types.TaskSpec{}

	for id, task := range def.Tasks {
		if task == nil {
			continue
		}

		p, err := parseTask(task)
		if err != nil {
			return nil, err
		}
		tasks[id] = p
	}

	return &types.WorkflowSpec{
		ApiVersion: def.APIVersion,
		OutputTask: def.Output,
		Tasks:      tasks,
	}, nil
}

func parseTask(t *taskSpec) (*types.TaskSpec, error) {
	deps := map[string]*types.TaskDependencyParameters{}
	for _, dep := range t.Requires {
		deps[dep] = &types.TaskDependencyParameters{}
	}

	inputs, err := parseInputs(t.Inputs)
	if err != nil {
		return nil, err
	}

	fn := t.Run
	if len(fn) == 0 {
		fn = defaultFunctionRef
	}

	result := &types.TaskSpec{
		FunctionRef: fn,
		Requires:    deps,
		Await:       int32(len(deps)),
		Inputs:      inputs,
	}

	return result, nil
}

// parseInputs parses the inputs of a task. This is typically a map[interface{}]interface{}.
func parseInputs(i interface{}) (map[string]*typedvalues.TypedValue, error) {
	if i == nil {
		return map[string]*typedvalues.TypedValue{}, nil
	}

	switch v := i.(type) {
	case map[string]interface{}:
		result := map[string]*typedvalues.TypedValue{}
		for inputKey, inputVal := range v {
			typedVal, err := parseInput(inputVal)
			if err != nil {
				return nil, err
			}
			result[inputKey] = typedVal
		}
		return result, nil
	case map[interface{}]interface{}:
		result := map[string]*typedvalues.TypedValue{}
		for inputKey, inputVal := range v {
			k := fmt.Sprintf("%v", inputKey)
			typedVal, err := parseInput(inputVal)
			if err != nil {
				return nil, err
			}
			result[k] = typedVal
		}
		return result, nil
	}
	p, err := parseInput(i)
	if err != nil {
		return nil, err
	}
	return map[string]*typedvalues.TypedValue{
		types.InputMain: p,
	}, nil
}

func parseInput(i interface{}) (*typedvalues.TypedValue, error) {
	// Handle special cases
	switch t := i.(type) {
	case []interface{}:
		// TODO shortcut - future: fix parsing of inputs to be recursive
		for k, v := range t {
			mp, ok := v.(map[interface{}]interface{})
			if ok {
				t[k] = convertInterfaceMaps(mp)
			}
		}
	case map[interface{}]interface{}:
		res := convertInterfaceMaps(t)
		if _, ok := res["run"]; ok {
			// The input might be a task
			td := &taskSpec{}
			bs, err := json.Marshal(res)
			err = json.Unmarshal(bs, td)
			if err != nil {
				panic(err)
			}

			p, err := parseTask(td)
			if err == nil {
				i = p
			} else {
				// Not a task
				i = res
			}
		} else if _, ok := res["tasks"]; ok {
			// The input might be a workflow
			td := &workflowSpec{}
			bs, err := json.Marshal(res)
			err = json.Unmarshal(bs, td)
			if err != nil {
				panic(err)
			}

			p, err := parseWorkflow(td)
			if err == nil {
				i = p
			} else {
				// Not a workflow
				i = res
			}
		} else {
			p, err := typedvalues.Wrap(res)
			if err != nil {
				return nil, err
			}
			i = p
		}
	case *taskSpec: // Handle taskSpec because it cannot be parsed by standard parser
		p, err := parseTask(t)
		if err != nil {
			return nil, err
		}
		i = p
	case *workflowSpec:
		w, err := parseWorkflow(t)
		if err != nil {
			return nil, err
		}
		i = w
	}

	p, err := typedvalues.Wrap(i)
	if err != nil {
		return nil, err
	}

	logrus.WithField("in", i).WithField("out", p).Debugf("parsed input")
	return p, nil
}

func convertInterfaceMaps(src map[interface{}]interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	for k, v := range src {
		if ii, ok := v.(map[interface{}]interface{}); ok {
			v = convertInterfaceMaps(ii)
		}
		res[fmt.Sprintf("%v", k)] = v
	}
	return res
}

//
// YAML data structures
//

type workflowSpec struct {
	APIVersion  string
	Description string
	Output      string
	Tasks       map[string]*taskSpec
}

type taskSpec struct {
	ID       string
	Run      string
	Inputs   interface{}
	Requires []string
}
