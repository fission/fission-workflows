package parse

import (
	"testing"

	"strings"

	"errors"

	"fmt"

	"github.com/fission/fission-workflow/pkg/api/function"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/magiconair/properties/assert"
)

func TestResolve(t *testing.T) {

	fooClient := "foo"

	clients := map[string]function.Resolver{
		fooClient: uppercaseResolver,
		"failing": failingResolver,
	}

	fmt.Printf("-- %v\n", clients)
	resolver := NewResolver(clients)

	task1 := "task1"
	task1Name := "lowercase"
	tasks := map[string]*types.Task{
		task1: {
			Name: task1Name,
		},
	}
	wf, err := resolver.Resolve(&types.WorkflowSpec{Tasks: tasks})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, wf.ResolvedTasks[task1].Resolved, strings.ToUpper(task1Name))
	assert.Equal(t, wf.ResolvedTasks[task1].Runtime, fooClient)
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
			Name: task1Name,
		},
	}
	wf, err := resolver.Resolve(&types.WorkflowSpec{Tasks: tasks})
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, wf.ResolvedTasks[task1].Resolved, strings.ToUpper(task1Name))
	assert.Equal(t, wf.ResolvedTasks[task1].Runtime, fooClient)
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
			Name: task1Name,
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
		return "", errors.New("Mock error")
	}}
)

type MockedFunctionResolver struct {
	Fn func(string) (string, error)
}

func (mk *MockedFunctionResolver) Resolve(fnName string) (string, error) {
	return mk.Fn(fnName)
}
