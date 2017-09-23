package parse

import (
	"testing"

	"strings"

	"errors"

	"github.com/fission/fission-workflows/pkg/api/function"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {

	fooClient := "foo"

	clients := map[string]function.Resolver{
		fooClient: uppercaseResolver,
		"failing": failingResolver,
	}

	resolver := NewResolver(clients)

	task1 := "task1"
	task1Name := "lowercase"
	tasks := map[string]*types.Task{
		task1: {
			FunctionRef: task1Name,
		},
	}
	wf, err := resolver.Resolve(&types.WorkflowSpec{Tasks: tasks})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, wf.ResolvedTasks[task1Name].Resolved, strings.ToUpper(task1Name))
	assert.Equal(t, wf.ResolvedTasks[task1Name].Runtime, fooClient)
}

func TestResolveForced(t *testing.T) {

	fooClient := "foo"

	clients := map[string]function.Resolver{
		"bar":     uppercaseResolver,
		fooClient: uppercaseResolver,
		"failing": failingResolver,
	}

	resolver := NewResolver(clients)

	task1 := "task1"
	task1Name := "foo:lowercase"
	tasks := map[string]*types.Task{
		task1: {
			FunctionRef: task1Name,
		},
	}
	wf, err := resolver.Resolve(&types.WorkflowSpec{Tasks: tasks})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, wf.ResolvedTasks[task1Name].Resolved, strings.ToUpper(task1Name))
	assert.Equal(t, wf.ResolvedTasks[task1Name].Runtime, fooClient)
}

func TestResolveInputs(t *testing.T) {

	fooClient := "foo"

	clients := map[string]function.Resolver{
		fooClient: uppercaseResolver,
		fooClient: uppercaseResolver,
	}

	resolver := NewResolver(clients)

	task1 := "task1"
	task1Name := "lowercase"
	nestedTaskName := "nestedLowercase"
	nestedNestedTaskName := "nestedLowercase"
	tasks := map[string]*types.Task{
		task1: {
			FunctionRef: task1Name,
			Inputs: map[string]*types.TypedValue{
				"nested": typedvalues.Flow(&types.Task{
					FunctionRef: nestedTaskName,
					Inputs: map[string]*types.TypedValue{
						"nested2": typedvalues.Flow(&types.Task{
							FunctionRef: nestedNestedTaskName,
						}),
					},
				}),
			},
		},
	}

	wf, err := resolver.Resolve(&types.WorkflowSpec{Tasks: tasks})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, strings.ToUpper(task1Name), wf.ResolvedTasks[task1Name].Resolved)
	assert.Equal(t, strings.ToUpper(nestedTaskName), wf.ResolvedTasks[nestedTaskName].Resolved)
	assert.Equal(t, strings.ToUpper(nestedNestedTaskName), wf.ResolvedTasks[nestedNestedTaskName].Resolved)
	assert.Equal(t, fooClient, wf.ResolvedTasks[nestedNestedTaskName].Runtime)
}

func TestResolveNotFound(t *testing.T) {

	clients := map[string]function.Resolver{
		"bar":     uppercaseResolver,
		"failing": failingResolver,
	}

	resolver := NewResolver(clients)

	task1 := "task1"
	task1Name := "foo:lowercase"
	tasks := map[string]*types.Task{
		task1: {
			FunctionRef: task1Name,
		},
	}
	wf, err := resolver.Resolve(&types.WorkflowSpec{Tasks: tasks})
	if err == nil {
		t.Fatal(wf)
	}
}

var (
	uppercaseResolver = &MockedFunctionResolver{func(name string) (string, error) {
		return strings.ToUpper(name), nil
	}}
	failingResolver = &MockedFunctionResolver{func(i string) (string, error) {
		return "", errors.New("mock error")
	}}
)

type MockedFunctionResolver struct {
	Fn func(string) (string, error)
}

func (mk *MockedFunctionResolver) Resolve(fnName string) (string, error) {
	return mk.Fn(fnName)
}
