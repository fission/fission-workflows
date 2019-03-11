package scheduler

import (
	"fmt"
	"time"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/graph"
	"github.com/golang/protobuf/ptypes"
)

var DefaultPolicy = NewHorizonPolicy()

// HorizonPolicy is the default policy of the workflow engine. It solely schedules tasks that are on the scheduling horizon.
//
// The scheduling horizon is the set of tasks that only depend on tasks that have already completed.
// If a task has failed this policy simply fails the workflow
type HorizonPolicy struct {
}

func NewHorizonPolicy() *HorizonPolicy {
	return &HorizonPolicy{}
}

func (p *HorizonPolicy) Evaluate(invocation *types.WorkflowInvocation) (*Schedule, error) {
	schedule := &Schedule{InvocationId: invocation.ID(), CreatedAt: ptypes.TimestampNow()}

	// If there are failed tasks halt the workflow
	if failedTasks := getFailedTasks(invocation); len(failedTasks) > 0 {
		for _, failedTask := range failedTasks {
			msg := fmt.Sprintf("Task '%v' failed", failedTask.ID())
			if err := failedTask.GetStatus().GetError(); err != nil {
				msg = err.Message
			}
			schedule.Abort = newAbortAction(msg)
		}
		return schedule, nil
	}

	// Find and schedule all tasks on the scheduling horizon
	openTasks := getOpenTasks(invocation)
	depGraph := graph.Parse(graph.NewTaskInstanceIterator(openTasks))
	horizon := graph.Roots(depGraph)
	for _, node := range horizon {
		schedule.AddRunTask(newRunTaskAction(node.(*graph.TaskInvocationNode).Task().ID()))
	}
	return schedule, nil
}

// PrewarmAllPolicy is the policy with the most aggressive form of prewarming.
//
// The policy, like the HorizonPolicy, schedules all tasks on the scheduling horizon optimistically.
// Similarly, it also fails workflow invocations immediately if a task has failed
//
// However, on top of the HorizonPolicy, this policy prewarms tasks aggressively. Any unstarted task not on the
// scheduling horizon will be prewarmed.
//
// This policy does not try to infer runtimes or cold starts; instead, it prewarms with a static duration.
type PrewarmAllPolicy struct {
	coldStartDuration time.Duration
}

func NewPrewarmAllPolicy(coldstartDuration time.Duration) *PrewarmAllPolicy {
	return &PrewarmAllPolicy{coldStartDuration: coldstartDuration}
}

func (p *PrewarmAllPolicy) Evaluate(invocation *types.WorkflowInvocation) (*Schedule, error) {
	schedule := &Schedule{InvocationId: invocation.ID(), CreatedAt: ptypes.TimestampNow()}

	// If there are failed tasks halt the workflow
	if failedTasks := getFailedTasks(invocation); len(failedTasks) > 0 {
		for _, failedTask := range failedTasks {
			msg := fmt.Sprintf("Task '%v' failed", failedTask.ID())
			if err := failedTask.GetStatus().GetError(); err != nil {
				msg = err.Message
			}
			schedule.Abort = newAbortAction(msg)
		}
		return schedule, nil
	}

	// Find and schedule all tasks on the scheduling horizon
	openTasks := getOpenTasks(invocation)
	depGraph := graph.Parse(graph.NewTaskInstanceIterator(openTasks))
	horizon := graph.Roots(depGraph)
	for _, node := range horizon {
		taskRun := node.(*graph.TaskInvocationNode)
		schedule.AddRunTask(newRunTaskAction(taskRun.TaskInvocation.ID()))
		delete(openTasks, taskRun.GetMetadata().GetId())
	}

	// Prewarm all other tasks
	expectedAt := time.Now().Add(p.coldStartDuration)
	for _, task := range openTasks {
		schedule.AddPrepareTask(newPrepareTaskAction(task.ID(), expectedAt))
	}
	return schedule, nil
}

// PrewarmHorizonPolicy is the policy with the most aggressive form of prewarming.
//
// The policy, like the HorizonPolicy, schedules all tasks on the scheduling horizon optimistically.
// Similarly, it also fails workflow invocations immediately if a task has failed
//
// However, on top of the HorizonPolicy, tries to policy prewarms tasks aggressively. Any unstarted task on the
// prewarm horizon will be prewarmed.
//
// This policy does not try to infer runtimes or cold starts; instead, it prewarms with a static duration.
type PrewarmHorizonPolicy struct {
	coldStartDuration time.Duration
}

func NewPrewarmHorizonPolicy(coldstartDuration time.Duration) *PrewarmHorizonPolicy {
	return &PrewarmHorizonPolicy{coldStartDuration: coldstartDuration}
}

func (p *PrewarmHorizonPolicy) Evaluate(invocation *types.WorkflowInvocation) (*Schedule, error) {
	schedule := &Schedule{InvocationId: invocation.ID(), CreatedAt: ptypes.TimestampNow()}

	// If there are failed tasks halt the workflow
	if failedTasks := getFailedTasks(invocation); len(failedTasks) > 0 {
		for _, failedTask := range failedTasks {
			msg := fmt.Sprintf("Task '%v' failed", failedTask.ID())
			if err := failedTask.GetStatus().GetError(); err != nil {
				msg = err.Message
			}
			schedule.Abort = newAbortAction(msg)
		}
		return schedule, nil
	}

	// Find and schedule all tasks on the scheduling horizon
	openTasks := getOpenTasks(invocation)
	depGraph := graph.Parse(graph.NewTaskInstanceIterator(openTasks))
	horizon := graph.Roots(depGraph)
	for _, node := range horizon {
		taskRun := node.(*graph.TaskInvocationNode)
		schedule.AddRunTask(newRunTaskAction(taskRun.TaskInvocation.ID()))
		delete(openTasks, taskRun.GetMetadata().GetId())
	}

	// Prewarm all tasks on the prewarm horizon
	// Note: we are mutating openTasks!
	expectedAt := time.Now().Add(p.coldStartDuration)
	prewarmDepGraph := graph.Parse(graph.NewTaskInstanceIterator(openTasks))
	prewarmHorizon := graph.Roots(prewarmDepGraph)
	for _, node := range prewarmHorizon {
		taskRun := node.(*graph.TaskInvocationNode)
		schedule.AddPrepareTask(newPrepareTaskAction(taskRun.Task().ID(), expectedAt))
	}

	return schedule, nil
}

func getFailedTasks(invocation *types.WorkflowInvocation) []*types.TaskInvocation {
	var failedTasks []*types.TaskInvocation
	for _, task := range invocation.TaskInvocations() {
		if task.GetStatus().GetStatus() == types.TaskInvocationStatus_FAILED {
			failedTasks = append(failedTasks, task)
		}
	}
	return failedTasks
}

func getOpenTasks(invocation *types.WorkflowInvocation) map[string]*types.TaskInvocation {
	openTasks := map[string]*types.TaskInvocation{}
	for id, task := range invocation.Tasks() {
		taskRun, ok := invocation.TaskInvocation(id)
		if !ok {
			// TODO extract this to types package
			taskRun = &types.TaskInvocation{
				Metadata: types.NewObjectMetadata(id),
				Spec: &types.TaskInvocationSpec{
					Task:         task,
					FnRef:        task.GetStatus().GetFnRef(),
					TaskId:       task.ID(),
					InvocationId: invocation.ID(),
					Inputs:       task.GetSpec().GetInputs(),
				},
				Status: &types.TaskInvocationStatus{
					Status: types.TaskInvocationStatus_UNKNOWN,
				},
			}
		}
		if taskRun.GetStatus().GetStatus() == types.TaskInvocationStatus_UNKNOWN {
			openTasks[id] = taskRun
		}
	}
	return openTasks
}
