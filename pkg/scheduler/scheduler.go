package scheduler

import (
	"fmt"

	"math"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
	"github.com/fission/fission-workflow/pkg/util"
	"github.com/golang/protobuf/ptypes"
	log "github.com/sirupsen/logrus"
)

type WorkflowScheduler struct {
}

func (ws *WorkflowScheduler) Evaluate(request *ScheduleRequest) (*Schedule, error) {
	schedule := &Schedule{
		InvocationId: request.Invocation.Metadata.Id,
		CreatedAt:    ptypes.TimestampNow(),
		Actions:      []*Action{},
	}

	ctxLog := log.WithFields(log.Fields{
		"workflow": request.Workflow,
		"invoke":   request.Invocation,
	})

	ctxLog.Info("Scheduler evaluating...")

	cwf := calculateCurrentWorkflow(request)

	openTasks := map[string]*TaskStatus{} // bool = nothing
	// Fill open tasks
	for id, t := range cwf {
		invokedTask, ok := request.Invocation.Status.Tasks[id]
		if !ok {
			openTasks[id] = t
			continue
		}
		if invokedTask.Status.Status == types.FunctionInvocationStatus_FAILED {
			AbortActionAny, _ := ptypes.MarshalAny(&AbortAction{
				Reason: fmt.Sprintf("Task '%s' failed!", invokedTask),
			})

			abortAction := &Action{
				Type:    ActionType_ABORT,
				Payload: AbortActionAny,
			}
			schedule.Actions = append(schedule.Actions, abortAction)
		}
	}
	if len(schedule.Actions) > 0 {
		return schedule, nil
	}

	// Determine horizon (aka tasks that can be executed now)
	horizon := map[string]*TaskStatus{}
	for id, task := range openTasks {
		if len(task.GetDependencies()) == 0 {
			horizon[id] = task
			break
		}

		completedDeps := 0
		for depName := range task.GetDependencies() {
			if _, ok := openTasks[depName]; !ok {
				completedDeps = completedDeps + 1
				break
			}
		}
		log.WithFields(log.Fields{
			"completedDeps": completedDeps,
			"task":          id,
			"max":           int(math.Max(float64(task.DependenciesAwait), float64(len(task.GetDependencies())-1))),
		}).Infof("Checking if dependencies have been satisfied")
		if completedDeps > int(math.Max(float64(task.DependenciesAwait), float64(len(task.GetDependencies())-1))) {
			horizon[id] = task
		}
	}

	ctxLog.WithField("horizon", horizon).Debug("Determined horizon")

	// Determine schedule nodes
	for taskId, taskDef := range horizon {
		// Fetch input
		inputs := taskDef.Inputs
		invokeTaskAction, _ := ptypes.MarshalAny(&InvokeTaskAction{
			Id:     taskId,
			Inputs: inputs,
		})

		schedule.Actions = append(schedule.Actions, &Action{
			Type:    ActionType_INVOKE_TASK,
			Payload: invokeTaskAction,
		})
	}

	ctxLog.WithField("schedule", schedule).Info("Determined schedule")

	return schedule, nil
}

func calculateCurrentWorkflow(req *ScheduleRequest) map[string]*TaskStatus {
	wf := map[string]*TaskStatus{}

	// Fill with static tasks
	for id, task := range req.Workflow.Spec.Tasks {
		taskStatus := req.Invocation.Status.Tasks[id]

		wf[id] = &TaskStatus{
			Task: task,
			Run:  taskStatus,
		}
	}

	// Mix in dynamic tasks naively
	for originId, originTask := range req.Invocation.Status.Tasks {
		output := originTask.Status.Output
		v, err := typedvalues.Format(output)
		if err != nil {
			log.Warnf("Error while parsing for tasks: %v", err)
			continue
		}
		t, ok := v.(*types.Task)
		if !ok {
			continue
		}

		// Generate ID
		id := util.CreateScopeId(originId, t.Id)
		taskStatus := req.Invocation.Status.Tasks[id]

		if t.Dependencies == nil {
			t.Dependencies = map[string]*types.TaskDependencyParameters{}
		}
		t.Dependencies[originId] = &types.TaskDependencyParameters{}

		// TODO recursive support
		wf[id] = &TaskStatus{
			Task: t,
			Run:  taskStatus,
		}

		// Fix dependencies
		for dId, dTask := range wf {
			if _, ok := dTask.Task.Dependencies[originId]; ok && dId != id {
				dTask.Task.Dependencies[id] = &types.TaskDependencyParameters{}
			}
		}
	}
	return wf
}

func InjectTask(originId string, originTask *types.FunctionInvocation, status map[string]*types.FunctionInvocation,
	target map[string]*TaskStatus) map[string]*TaskStatus {
	output := originTask.Status.Output
	v, err := typedvalues.Format(output)
	if err != nil {
		log.Warnf("Error while parsing for tasks: %v", err)
		return target
	}
	t, ok := v.(*types.Task)
	if !ok {
		return target
	}

	// Generate ID
	// TODO allow some scope modifications to avoid origin_step_step_step_step_step (instead origin_step_5)
	id := util.CreateScopeId(originId, t.Id)
	taskInvocation := status[id]

	if t.Dependencies == nil {
		t.Dependencies = map[string]*types.TaskDependencyParameters{}
	}
	t.Dependencies[originId] = &types.TaskDependencyParameters{}

	target[id] = &TaskStatus{
		Task: t,
		Run:  taskInvocation,
	}

	// Fix dependencies
	for dId, dTask := range target {
		if _, ok := dTask.Task.Dependencies[originId]; ok && dId != id {
			dTask.Task.Dependencies[id] = &types.TaskDependencyParameters{}
		}
	}

	if taskInvocation != nil {
		v, err := typedvalues.Format(output)
		if err != nil {
			log.Warnf("Error while parsing for tasks: %v", err)
			return target
		}
		_, ok := v.(*types.Task)
		if !ok {
			return target
		}
		return InjectTask(id, taskInvocation, status, target)
	}
	return target

}

type TaskStatus struct {
	*types.Task
	Run *types.FunctionInvocation
}
