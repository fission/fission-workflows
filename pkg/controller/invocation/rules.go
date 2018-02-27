package invocation

import (
	"errors"

	"github.com/fission/fission-workflows/pkg/api/function"
	"github.com/fission/fission-workflows/pkg/api/invocation"
	"github.com/fission/fission-workflows/pkg/controller"
	"github.com/fission/fission-workflows/pkg/scheduler"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
)

//
// Invocation-specific rules
//

type EvalContext interface {
	controller.EvalContext
	Workflow() *types.Workflow
	Invocation() *types.WorkflowInvocation
}

type evalContext struct {
	controller.EvalContext
	wf  *types.Workflow
	wfi *types.WorkflowInvocation
}

func NewEvalContext(state *controller.EvalState, wf *types.Workflow, wfi *types.WorkflowInvocation) evalContext {
	return evalContext{
		EvalContext: controller.NewEvalContext(state),
		wf:          wf,
		wfi:         wfi,
	}
}

func (ec evalContext) Workflow() *types.Workflow {
	return ec.wf
}

func (ec evalContext) Invocation() *types.WorkflowInvocation {
	return ec.wfi
}

type RuleHasCompleted struct{}

func (cf *RuleHasCompleted) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureInvocationContext(cec)
	wfi := ec.Invocation()
	if wfi.Status.Finished() {
		log.Infof("No need to evaluate finished invocation %v", wfi.Metadata.Id)
	}
	return nil
}

type RuleWorkflowIsReady struct {
}

func (wr *RuleWorkflowIsReady) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureInvocationContext(cec)
	wf := ec.Workflow()
	// Check if workflow is in the right state to use.
	if !wf.Status.Ready() {
		log.WithField("wf.status", wf.Status.Status).Error("Workflow is not ready yet.")
		return &controller.ActionSkip{} // TODO backoff action
	}
	return nil
}

type RuleSchedule struct {
	Scheduler     *scheduler.WorkflowScheduler
	InvocationApi *invocation.Api
	FunctionApi   *function.Api
}

func (sf *RuleSchedule) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureInvocationContext(cec)
	wf := ec.Workflow()
	wfi := ec.Invocation()
	// Request a execution plan from the Scheduler
	schedule, err := sf.Scheduler.Evaluate(&scheduler.ScheduleRequest{
		Invocation: wfi,
		Workflow:   wf,
	})
	if err != nil {
		return nil
	}

	// Execute the actions as specified in the execution plan
	for _, a := range schedule.Actions {
		switch a.Type {
		case scheduler.ActionType_ABORT:
			return &ActionFail{
				Api:          sf.InvocationApi,
				InvocationId: wfi.Id(),
			}
		case scheduler.ActionType_INVOKE_TASK:
			invokeAction := &scheduler.InvokeTaskAction{}
			err := ptypes.UnmarshalAny(a.Payload, invokeAction)
			if err != nil {
				log.Errorf("Failed to unpack Scheduler action: %v", err)
			}
			return &ActionInvokeTask{
				Wf:   wf,
				Wfi:  wfi,
				Api:  sf.FunctionApi,
				Task: invokeAction,
			}
		default:
			log.Warnf("Unknown Scheduler action: '%v'", a)
		}
	}
	return nil
}

type RuleCheckIfCompleted struct {
	InvocationApi *invocation.Api
}

func (cc *RuleCheckIfCompleted) Eval(cec controller.EvalContext) controller.Action {
	ec := EnsureInvocationContext(cec)
	wf := ec.Workflow()
	wfi := ec.Invocation()
	// Check if the workflow invocation is complete
	tasks := types.GetTasks(wf, wfi)
	var err error
	finished := true
	for id := range tasks {
		t, ok := wfi.Status.Tasks[id]
		if !ok || !t.Status.Finished() {
			finished = false
			break
		}
	}
	if finished {
		var finalOutput *types.TypedValue
		if len(wf.Spec.OutputTask) != 0 {
			t, ok := wfi.Status.Tasks[wf.Spec.OutputTask]
			if !ok {
				return &controller.ActionError{
					Err: errors.New("could not find output task status in completed invocation"),
				}
			}
			finalOutput = t.Status.Output
		}

		// TODO extract to action
		err = cc.InvocationApi.MarkCompleted(wfi.Id(), finalOutput)
		return &controller.ActionError{
			Err: err,
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
