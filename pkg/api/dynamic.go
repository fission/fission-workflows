package api

import (
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/golang/protobuf/proto"
	"github.com/mitchellh/hashstructure"
	"github.com/sirupsen/logrus"
)

// Dynamic contains the API functionality for creating dynamic tasks and workflows.
type Dynamic struct {
	wfAPI  *Workflow
	wfiAPI *Invocation
}

// NewDynamicAPI creates the Dynamic API.
func NewDynamicAPI(wfAPI *Workflow, wfiAPI *Invocation) *Dynamic {
	return &Dynamic{
		wfAPI:  wfAPI,
		wfiAPI: wfiAPI,
	}
}

// AddDynamicFlow inserts the flow as a 'dynamic task' into the workflow invocation with id invocationID as the child
// of the parent task.
func (ap *Dynamic) AddDynamicFlow(invocationID string, parentTaskID string, flow typedvalues.Flow) error {
	if err := validate.Flow(flow); err != nil {
		return err
	}
	switch flow.Type() {
	case typedvalues.Workflow:
		return ap.addDynamicWorkflow(invocationID, parentTaskID, flow.Workflow(), &types.TaskSpec{})
	case typedvalues.Task:
		return ap.addDynamicTask(invocationID, parentTaskID, flow.Task())
	default:
		panic("validated flow was still empty")
	}
}

func (ap *Dynamic) addDynamicTask(invocationID string, parentTaskID string, taskSpec *types.TaskSpec) error {
	// Transform TaskSpec into WorkflowSpec
	wfSpec := &types.WorkflowSpec{
		OutputTask: "main",
		Tasks: map[string]*types.TaskSpec{
			"main": taskSpec,
		},
		Dynamic:    true, // FUTURE: use internal as indicator to cleanup and hide these generated workflows
		ApiVersion: types.WorkflowAPIVersion,
	}
	hash, err := hashstructure.Hash(wfSpec, nil)
	if err == nil {
		wfSpec.ForceId = string(hash)
	} else {
		logrus.Errorf("Failed to generate hash of workflow; defaulting to random id: %v", err)
	}
	return ap.addDynamicWorkflow(invocationID, parentTaskID, wfSpec, taskSpec)
}

func (ap *Dynamic) addDynamicWorkflow(invocationID string, parentTaskID string, wfSpec *types.WorkflowSpec,
	stubTask *types.TaskSpec) error {

	// Clean-up WorkflowSpec and submit
	sanitizeWorkflow(wfSpec)
	err := validate.WorkflowSpec(wfSpec)
	if err != nil {
		return err
	}
	wfID, err := ap.wfAPI.Create(wfSpec)
	if err != nil && err != ErrWorkflowAlreadyExists {
		return err
	}

	// Create function reference to workflow
	wfRef := createFnRef(wfID)
	// Generate Proxy Task
	proxyTaskSpec := proto.Clone(stubTask).(*types.TaskSpec)
	proxyTaskSpec.FunctionRef = wfRef.Format()
	proxyTaskSpec.Input(types.InputParent, typedvalues.ParseString(invocationID))
	proxyTaskID := parentTaskID + "_child"
	proxyTask := types.NewTask(proxyTaskID, proxyTaskSpec.FunctionRef)
	proxyTask.Spec = proxyTaskSpec
	// Shortcut resolving of the function reference
	proxyTask.Status.Status = types.TaskStatus_READY
	proxyTask.Status.FnRef = &wfRef

	// Ensure that the only link of the dynamic task is with its parent
	proxyTaskSpec.Requires = map[string]*types.TaskDependencyParameters{
		parentTaskID: {
			Type: types.TaskDependencyParameters_DYNAMIC_OUTPUT,
		},
	}

	err = validate.TaskSpec(proxyTaskSpec)
	if err != nil {
		return err
	}

	// Submit added task to workflow invocation
	// TODO replace Task with TaskSpec + shortcircuit resolving of function (e.g. special label on fnref)
	return ap.wfiAPI.AddTask(invocationID, proxyTask)
}

func sanitizeWorkflow(v *types.WorkflowSpec) {
	if len(v.ApiVersion) == 0 {
		v.ApiVersion = types.WorkflowAPIVersion
	}

	// ForceID is not supported for internal workflows
	v.ForceId = ""
}

func createFnRef(wfID string) types.FnRef {
	return types.FnRef{
		Runtime: "workflows",
		ID:      wfID,
	}
}
