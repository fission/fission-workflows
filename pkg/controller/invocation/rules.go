package invocation

import (
	"errors"

	"github.com/fission/fission-workflows/pkg/api"
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/controller/expr"
	"github.com/fission/fission-workflows/pkg/scheduler"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/typedvalues/controlflow"
	"github.com/golang/protobuf/ptypes"
)

//
// Invocation-specific rules
//

type EvalContext interface {
	controller.EvalContext
	Invocation() *types.WorkflowInvocation
}

type WfiEvalContext struct {
	controller.EvalContext
	wfi *types.WorkflowInvocation
}

func NewEvalContext(state *controller.EvalState, wfi *types.WorkflowInvocation) WfiEvalContext {
	return WfiEvalContext{
		EvalContext: controller.NewEvalContext(state),
		wfi:         wfi,
	}
}

func (ec WfiEvalContext) Invocation() *types.WorkflowInvocation {
	return ec.wfi
}

type RuleWorkflowIsReady struct {
	InvocationAPI *api.Invocation
}

func (wr *RuleWorkflowIsReady) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureInvocationContext(cec)
	wf := ec.Invocation().Workflow()
	if wf.GetStatus().GetStatus() == types.WorkflowStatus_DELETED {
		return &ActionFail{
			API:          wr.InvocationAPI,
			InvocationID: ec.Invocation().ID(),
			Err:          errors.New("workflow is deleted"),
		}
	}

	// Check if workflow is still in progress.
	if !wf.GetStatus().Ready() {
		log.WithField("wf.status", wf.Status.Status).Error("Workflow is not ready yet.")
		return &controller.ActionSkip{} // TODO backoff action
	}
	return nil
}

type RuleSchedule struct {
	Scheduler     *scheduler.InvocationScheduler
	InvocationAPI *api.Invocation
	FunctionAPI   *api.Task
	StateStore    *expr.Store
}

func (sf *RuleSchedule) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureInvocationContext(cec)
	wf := ec.Invocation().Workflow()
	wfi := ec.Invocation()
	// Request a execution plan from the Scheduler
	// TODO remove once invocation is guaranteed to have workflow
	if wfi.Spec.Workflow == nil {
		wfi.Spec.Workflow = wf
	}
	schedule, err := sf.Scheduler.Evaluate(wfi)
	if err != nil {
		return nil
	}

	// Execute the actions as specified in the execution plan
	var actions []controller.Action
	for _, a := range schedule.Actions {
		switch a.Type {
		case scheduler.ActionType_ABORT:
			invokeAction := &scheduler.AbortAction{}
			err := ptypes.UnmarshalAny(a.Payload, invokeAction)
			if err != nil {
				log.Errorf("Failed to unpack Scheduler action: %v", err)
			}
			return &ActionFail{
				API:          sf.InvocationAPI,
				InvocationID: wfi.ID(),
				Err:          errors.New(invokeAction.Reason),
			}
		case scheduler.ActionType_INVOKE_TASK:
			invokeAction := &scheduler.InvokeTaskAction{}
			err := ptypes.UnmarshalAny(a.Payload, invokeAction)
			if err != nil {
				log.Errorf("Failed to unpack Scheduler action: %v", err)
			}
			actions = append(actions, &ActionInvokeTask{
				ec:         ec.EvalState(),
				Wfi:        wfi,
				API:        sf.FunctionAPI,
				Task:       invokeAction,
				StateStore: sf.StateStore,
			})
		default:
			log.Warnf("Unknown Scheduler action: '%v'", a)
		}
	}
	return &controller.MultiAction{Actions: actions}
}

type RuleCheckIfCompleted struct {
	InvocationAPI *api.Invocation
}

func (cc *RuleCheckIfCompleted) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureInvocationContext(cec)
	wf := ec.Invocation().Workflow()
	wfi := ec.Invocation()
	// Check if the workflow invocation is complete
	var err error
	finished := true
	success := true
	for id := range wfi.Tasks() {
		t, ok := wfi.Status.Tasks[id]
		if !ok || !t.Status.Finished() {
			finished = false
			break
		} else {
			success = success && t.Status.Status == types.TaskInvocationStatus_SUCCEEDED
		}
	}
	if finished {
		var finalOutput *typedvalues.TypedValue
		var finalOutputHeaders *typedvalues.TypedValue
		if len(wf.Spec.OutputTask) != 0 {
			finalOutput = controlflow.ResolveTaskOutput(wf.Spec.OutputTask, wfi)
			finalOutputHeaders = controlflow.ResolveTaskOutputHeaders(wf.Spec.OutputTask, wfi)
		}

		// TODO extract to action
		if success {
			err = cc.InvocationAPI.Complete(wfi.ID(), finalOutput, finalOutputHeaders)
		} else {
			err = cc.InvocationAPI.Fail(wfi.ID(), errors.New("not all tasks succeeded"))
		}
		if err != nil {
			return &controller.ActionError{
				Err: err,
			}
		}
	}
	return nil
}

func EnsureInvocationContext(cec controller.EvalContext) EvalContext {
	ec, ok := cec.(EvalContext)
	if !ok {
		panic("invalid evaluation context")
	}
	return ec
}
