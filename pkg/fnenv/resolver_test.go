package fnenv

import (
	"errors"
	"strings"
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/stretchr/testify/assert"
)

func TestResolve(t *testing.T) {

	fooClient := "foo"

	clients := map[string]RuntimeResolver{
		fooClient: uppercaseResolver,
		"failing": failingResolver,
	}

	resolver := NewMetaResolver(clients)

	task1 := "task1"
	task1Name := "lowercase"
	tasks := map[string]*types.TaskSpec{
		task1: {
			FunctionRef: task1Name,
		},
	}
	wf, err := ResolveTasks(resolver, tasks)
	assert.NoError(t, err)
	assert.Equal(t, wf[task1Name].ID, strings.ToUpper(task1Name))
	assert.Equal(t, wf[task1Name].Runtime, fooClient)
}

func TestResolveForced(t *testing.T) {

	fooClient := "foo"

	clients := map[string]RuntimeResolver{
		"bar":     uppercaseResolver,
		fooClient: uppercaseResolver,
		"failing": failingResolver,
	}

	resolver := NewMetaResolver(clients)

	task1 := "task1"
	task1Ref := types.NewFnRef(fooClient, "", "lowercase")
	task1Name := task1Ref.Format()
	tasks := map[string]*types.TaskSpec{
		task1: {
			FunctionRef: task1Name,
		},
	}
	wf, err := ResolveTasks(resolver, tasks)
	assert.NoError(t, err)
	assert.Equal(t, strings.ToUpper(task1Ref.ID), wf[task1Name].ID)
	assert.Equal(t, wf[task1Name].Runtime, fooClient)
}

func TestResolveInputs(t *testing.T) {

	fooClient := "foo"

	clients := map[string]RuntimeResolver{
		fooClient: uppercaseResolver,
		fooClient: uppercaseResolver,
	}

	resolver := NewMetaResolver(clients)

	task1 := "task1"
	task1Name := "lowercase"
	nestedTaskName := "nestedLowercase"
	nestedNestedTaskName := "nestedLowercase"
	tasks := map[string]*types.TaskSpec{
		task1: {
			FunctionRef: task1Name,
			Inputs: map[string]*types.TypedValue{
				"nested": typedvalues.ParseTask(&types.TaskSpec{
					FunctionRef: nestedTaskName,
					Inputs: map[string]*types.TypedValue{
						"nested2": typedvalues.ParseTask(&types.TaskSpec{
							FunctionRef: nestedNestedTaskName,
						}),
					},
				}),
			},
		},
	}

	wf, err := ResolveTasks(resolver, tasks)
	assert.NoError(t, err)
	assert.Equal(t, strings.ToUpper(task1Name), wf[task1Name].ID)
	//assert.Equal(t, strings.ToUpper(nestedTaskName), wf[nestedTaskName].RuntimeId)
	//assert.Equal(t, strings.ToUpper(nestedNestedTaskName), wf[nestedNestedTaskName].RuntimeId)
	//assert.Equal(t, fooClient, wf[nestedNestedTaskName].Runtime)
	assert.Nil(t, wf[nestedTaskName])
}

func TestResolveNotFound(t *testing.T) {

	clients := map[string]RuntimeResolver{
		"bar":     uppercaseResolver,
		"failing": failingResolver,
	}

	resolver := NewMetaResolver(clients)

	task1 := "task1"
	task1Name := "foo:lowercase"
	tasks := map[string]*types.TaskSpec{
		task1: {
			FunctionRef: task1Name,
		},
	}
	_, err := ResolveTasks(resolver, tasks)
	assert.Error(t, err)
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

func (mk *MockedFunctionResolver) Resolve(ref types.FnRef) (string, error) {
	return mk.Fn(ref.ID)
}
