package types

// CalculateTaskDependencyGraph combines the static workflow definition with the dynamic invocation to determine
// the current state of the workflow.
func CalculateTaskDependencyGraph(wf *Workflow, invoc *WorkflowInvocation) map[string]*TaskStatus {
	target := map[string]*TaskStatus{}

	addStaticTasks(wf, invoc, target)

	// Add dynamic tasks overriding similarly named static tasks in case of conflicts
	addDynamicTasks(invoc, target)

	return target
}

func addStaticTasks(wf *Workflow, invoc *WorkflowInvocation, target map[string]*TaskStatus) {
	for id, task := range wf.Spec.Tasks {
		taskStatus := invoc.Status.Tasks[id]
		target[id] = &TaskStatus{
			Task:       task,
			Invocation: taskStatus,
		}
	}
}

func addDynamicTasks(invoc *WorkflowInvocation, target map[string]*TaskStatus) {
	funktorMapping := map[string]string{}
	for id, task := range invoc.Status.DynamicTasks {
		taskStatus := invoc.Status.Tasks[id]
		target[id] = &TaskStatus{
			Task:       task,
			Invocation: taskStatus,
		}

		for depId, depParams := range task.Dependencies {
			if depParams.Type == TaskDependencyParameters_FUNKTOR_OUTPUT {
				funktorMapping[depId] = id
			}
		}
		// TODO Amend static task if it exists instead of overriding it completely.
	}

	// Reroute dependencies to also depend on the outputted task of funktors.
	for _, task := range target {
		for depId, depParams := range task.Dependencies {
			if outputTask, ok := funktorMapping[depId]; ok && depParams.Type != TaskDependencyParameters_FUNKTOR_OUTPUT {
				task.Dependencies[outputTask] = &TaskDependencyParameters{}
			}
		}
	}
}

type TaskStatus struct {
	*Task
	// Invocation is nil if not yet invoked
	Invocation *FunctionInvocation
}
