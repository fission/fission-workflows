package typedvalues

import (
	"github.com/fission/fission-workflow/pkg/types"
)

func ResolveTaskOutput(taskId string, invoc *types.WorkflowInvocation) *types.TypedValue {
	val, ok := invoc.Status.Tasks[taskId]
	if !ok {
		return nil
	}

	output := val.Status.Output
	if output == nil {
		return nil
	}

	if output.Type == TYPE_FLOW {
		for outputTaskId, outputTask := range invoc.Status.DynamicTasks {
			if dep, ok := outputTask.Dependencies[taskId]; ok && dep.Type == types.TaskDependencyParameters_FUNKTOR_OUTPUT {
				return ResolveTaskOutput(outputTaskId, invoc)
			}
		}
		return nil
	}
	return output
}
