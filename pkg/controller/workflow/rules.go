package workflow

import (
	"github.com/fission/fission-workflows/pkg/api/workflow"
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/types"
)

//
// Workflow-specific rules
//

type EvalContext interface {
	controller.EvalContext
	Workflow() *types.Workflow
}

type evalContext struct {
	controller.EvalContext
	wf *types.Workflow
}

func (ec evalContext) Workflow() *types.Workflow {
	return ec.wf
}

func NewEvalContext(state *controller.EvalState, wf *types.Workflow) evalContext {
	return evalContext{
		EvalContext: controller.NewEvalContext(state),
		wf:          wf,
	}
}

type RuleSkipIfReady struct{}

func (r *RuleSkipIfReady) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureWorkflowContext(cec)
	wf := ec.Workflow()
	if wf.Status != nil && wf.Status.Ready() {
		return &controller.ActionSkip{}
	}
	return nil
}

type RuleEnsureParsed struct {
	WfApi *workflow.Api
}

func (r *RuleEnsureParsed) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureWorkflowContext(cec)
	wf := ec.Workflow()
	if wf.Status == nil || !wf.Status.Ready() {
		return &ActionParseWorkflow{
			WfApi: r.WfApi,
			Wf:    wf,
		}
	}
	return nil
}

func EnsureWorkflowContext(cec controller.EvalContext) EvalContext {
	ec, ok := cec.(EvalContext)
	if !ok {
		panic("invalid evaluation context")
	}
	return ec
}

type RuleRemoveIfDeleted struct {
	evalCache *controller.EvalCache
}

func (r *RuleRemoveIfDeleted) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureWorkflowContext(cec)
	wf := ec.Workflow()
	if wf.Status.Status == types.WorkflowStatus_DELETED {
		return &controller.ActionRemoveFromEvalCache{
			EvalCache: r.evalCache,
			Id:        wf.Id(),
		}
	}
	return nil
}
