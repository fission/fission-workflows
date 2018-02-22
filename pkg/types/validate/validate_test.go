package validate

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/stretchr/testify/assert"
)

func validSpec() *types.WorkflowSpec {
	return &types.WorkflowSpec{
		ApiVersion: types.WorkflowApiVersion,
		OutputTask: "last",
		Tasks: map[string]*types.TaskSpec{
			"first": {
				FunctionRef: "fn",
			},
			"middle": {
				FunctionRef: "fn",
				Requires: map[string]*types.TaskDependencyParameters{
					"first": nil,
				},
			},
			"last": {
				FunctionRef: "fn",
				Requires: map[string]*types.TaskDependencyParameters{
					"middle": nil,
				},
			},
		},
	}
}

func TestWorkflowSpecValid(t *testing.T) {
	err := WorkflowSpec(validSpec())
	assert.NoError(t, err, Format(err))
}

func TestWorkflowSpecInvalidApiVersion(t *testing.T) {
	spec := validSpec()
	spec.ApiVersion = "nonExistent"
	assert.Error(t, WorkflowSpec(spec))
}

func TestWorkflowSpecInvalidOutputTask(t *testing.T) {
	spec := validSpec()
	spec.OutputTask = "nonExistent"
	assert.Error(t, WorkflowSpec(spec))
}

func TestWorkflowSpecNoTasks(t *testing.T) {
	spec := validSpec()
	spec.Tasks = map[string]*types.TaskSpec{}
	assert.Error(t, WorkflowSpec(spec))
}

func TestWorkflowSpecInvalidRequires(t *testing.T) {
	spec := validSpec()
	spec.Tasks["middle"].Require("nonExistentDep")
	assert.Error(t, WorkflowSpec(spec))
}

func TestWorkflowSpecInvalidCircularDependency(t *testing.T) {
	spec := validSpec()
	spec.Tasks["first"].Require("last")
	assert.Error(t, WorkflowSpec(spec))
}
