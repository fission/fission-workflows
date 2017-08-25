package query

import (
	"errors"
	"strings"

	"reflect"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
)

/**

Data Structure presented to the user:
---
workflow:
	ID
	name TODO not supported yet
invocation:
	ID
tasks:
	Type
	status
	startedAt
---
*/

var (
	JSONPATH_CHILD        = "."
	JSONPATH_ROOT         = "$"
	JSONPATH_CURRENT_TASK = "@"
)

var (
	ErrInvalidQuery        = errors.New("Invalid selector query")
	ErrUnsupportedDataType = errors.New("Query targets value of unsupported type")
)

/**
This selector allows the use of JSONPath (http://goessner.net/articles/JsonPath/index.html#e2) for selecting data.

Supported functionality / semantics:
- `$` = from workflow
- `@` = current task
- `.` = child element
*/
type JsonPathSelector struct {
	pf typedvalues.ParserFormatter
}

func (jp *JsonPathSelector) Select(root *types.WorkflowInvocation, query string, cwd ...string) (*types.TypedValue, error) {
	// Check preconditions
	if strings.HasPrefix(query, JSONPATH_CURRENT_TASK) && len(cwd) > 0 {
		query = cwd[0]
	}

	if !strings.HasPrefix(query, JSONPATH_ROOT) {
		return nil, ErrInvalidQuery
	}

	// Normalize
	// TODO make scase-insensitive
	query = strings.Trim(query, " ")

	// TODO fix hard-coding certain look-ups to more consist format (this is just for prototyping the view)
	switch {
	case strings.EqualFold(query, "$.workflow.id"):
		return jp.pf.Parse(root.Spec.WorkflowId)

	case strings.EqualFold(query, "$.invocation.id"):
		return jp.pf.Parse(root.Metadata.Id)

	case strings.EqualFold(query, "$.invocation.startedAt"):
		return jp.pf.Parse(root.Metadata.CreatedAt.String())
	case strings.HasPrefix(query, "$.invocation.inputs"):
		c := strings.Split(query, JSONPATH_CHILD)
		input, ok := root.Spec.Inputs[c[3]]
		if !ok {
			return nil, nil
		}
		return jp.selectJsonTypedValue(input, c[4:])

	case strings.HasPrefix(query, "$.tasks"):
		// TODO currently just use tasks in status to catch most use cases, should be refactored to also include non-started tasks.
		c := strings.Split(query, JSONPATH_CHILD)
		task, ok := root.Status.Tasks[c[2]]
		if !ok {
			return nil, nil
		}
		return jp.selectTask(task, c[3:])
	}

	// Nothing could be found
	return nil, nil
}

// taskQuery consists of the task-scoped query (e.g. input.<xyz>)
func (jp *JsonPathSelector) selectTask(task *types.FunctionInvocation, taskQuery []string) (*types.TypedValue, error) {
	switch taskQuery[0] {
	case "status":
		return jp.pf.Parse(task.Status.Status.String())
	case "startedAt":
		return jp.pf.Parse(task.Metadata.CreatedAt.String())
	case "completedAt":
		if task.Status.Status.Finished() {
			return jp.pf.Parse(task.Status.Status.String())
		}
	case "output":
		return jp.selectJsonTypedValue(task.Status.Output, taskQuery[1:])
	}
	return nil, nil
}

// TODO move this out to typedvalues to support more data formats
func (jp *JsonPathSelector) selectJsonTypedValue(root *types.TypedValue, query []string) (*types.TypedValue, error) {
	if len(query) == 0 {
		return root, nil
	}

	val, err := jp.pf.Format(root)
	if err != nil {
		return nil, err
	}

	result, err := jp.traverse(val, query)
	if err != nil {
		return nil, err
	}

	return jp.pf.Parse(result)
}

// query: foo.bar
func (jp *JsonPathSelector) traverse(root interface{}, query []string) (interface{}, error) {
	src := reflect.Indirect(reflect.ValueOf(root))

	var result interface{}
	for _, node := range query {

		field := src.FieldByName(node)
		if !field.IsValid() {
			return nil, nil
		}
		result = src.Interface()
	}

	return result, nil
}
