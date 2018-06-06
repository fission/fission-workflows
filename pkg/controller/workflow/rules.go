package workflow

import (
	"github.com/fission/fission-workflows/pkg/api"
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

type WfEvalContext struct {
	controller.EvalContext
	wf *types.Workflow
}

func (ec WfEvalContext) Workflow() *types.Workflow {
	return ec.wf
}

func NewEvalContext(state *controller.EvalState, wf *types.Workflow) WfEvalContext {
	return WfEvalContext{
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
	WfAPI *api.Workflow
}

func (r *RuleEnsureParsed) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureWorkflowContext(cec)
	wf := ec.Workflow()
	if wf.Status == nil || !wf.Status.Ready() {
		return &ActionParseWorkflow{
			WfAPI: r.WfAPI,
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
	evalCache controller.EvalStore
}

func (r *RuleRemoveIfDeleted) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureWorkflowContext(cec)
	wf := ec.Workflow()
	if wf.Status.Status == types.WorkflowStatus_DELETED {
		return &controller.ActionRemoveFromEvalCache{
			EvalCache: r.evalCache,
			ID:        wf.ID(),
		}
	}
	return nil
}
