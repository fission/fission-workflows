package aggregates

import (
	"fmt"

	"errors"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/gogo/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

const (
	TYPE_WORKFLOW_INVOCATION = "invocation"
)

type WorkflowInvocation struct {
	*fes.AggregatorMixin
	*types.WorkflowInvocation
}

func NewWorkflowInvocation(invocationId string, wi *types.WorkflowInvocation) *WorkflowInvocation {
	wia := &WorkflowInvocation{
		WorkflowInvocation: wi,
	}

	wia.AggregatorMixin = fes.NewAggregatorMixin(wia, *NewWorkflowInvocationAggregate(invocationId))
	return wia
}

func NewWorkflowInvocationAggregate(invocationId string) *fes.Aggregate {
	return &fes.Aggregate{
		Id:   invocationId,
		Type: TYPE_WORKFLOW_INVOCATION,
	}
}

func (wi *WorkflowInvocation) ApplyEvent(event *fes.Event) error {
	// If the event is a function event, use the Function Aggregate to resolve it.
	if event.Aggregate.Type == TYPE_FUNCTION_INVOCATION {
		return wi.applyTaskEvent(event)
	}

	// Otherwise assume that this is a invocation event
	eventType, err := events.ParseInvocation(event.Type)
	if err != nil {
		return err
	}

	switch eventType {
	case events.Invocation_INVOCATION_CREATED:
		spec := &types.WorkflowInvocationSpec{}
		err := proto.Unmarshal(event.Data, spec)
		if err != nil {
			return fmt.Errorf("failed to unmarshal event: '%v' (%v)", event, err)
		}

		wi.AggregatorMixin = fes.NewAggregatorMixin(wi, *event.Aggregate)
		wi.WorkflowInvocation = &types.WorkflowInvocation{
			Metadata: &types.ObjectMetadata{
				Id:        event.Aggregate.Id,
				CreatedAt: event.Timestamp,
			},
			Spec: spec,
			Status: &types.WorkflowInvocationStatus{
				Status:       types.WorkflowInvocationStatus_IN_PROGRESS,
				Tasks:        map[string]*types.TaskInvocation{},
				UpdatedAt:    event.GetTimestamp(),
				DynamicTasks: map[string]*types.Task{},
			},
		}
	case events.Invocation_INVOCATION_CANCELED:
		ivErr := &types.Error{}
		err := proto.Unmarshal(event.Data, ivErr)
		if err != nil {
			ivErr.Code = "error"
			ivErr.Message = err.Error()
			log.Errorf("failed to unmarshal event: '%v' (%v)", event, err)
		}

		wi.Status.Status = types.WorkflowInvocationStatus_ABORTED
		wi.Status.UpdatedAt = event.GetTimestamp()
		wi.Status.Error = ivErr
	case events.Invocation_INVOCATION_COMPLETED:
		status := &types.WorkflowInvocationStatus{}
		err = proto.Unmarshal(event.Data, status)
		if err != nil {
			return fmt.Errorf("failed to unmarshal event: '%v' (%v)", event, err)
		}

		if wi.Status == nil {
			wi.Status = &types.WorkflowInvocationStatus{}
		}

		wi.Status.Status = types.WorkflowInvocationStatus_SUCCEEDED
		wi.Status.Output = status.Output
		wi.Status.UpdatedAt = event.GetTimestamp()
	default:
		log.WithFields(log.Fields{
			"event": event,
		}).Warn("Skipping unimplemented event.")
	}
	if err != nil {
		return err
	}

	return nil
}

func (wi *WorkflowInvocation) applyTaskEvent(event *fes.Event) error {
	if wi.Aggregate() != *event.Parent {
		return errors.New("function does not belong to invocation")
	}
	taskId := event.Aggregate.Id
	task, ok := wi.Status.Tasks[taskId]
	if !ok {
		task = &types.TaskInvocation{}
	}
	ti := NewTaskInvocation(taskId, task)
	err := ti.ApplyEvent(event)
	if err != nil {
		return err
	}
	wi.Status.Tasks[taskId] = ti.TaskInvocation

	// Handle dynamic tasks
	output := ti.TaskInvocation.Status.Output
	if output != nil && output.Type == typedvalues.TYPE_FLOW {
		i, _ := typedvalues.Format(output)
		dynamicTask := i.(*types.Task)
		id := util.CreateScopeId(taskId, dynamicTask.Id)
		dynamicTask.Id = id
		if dynamicTask.Requires == nil {
			dynamicTask.Requires = map[string]*types.TaskDependencyParameters{}
		}

		if dynamicTask.Inputs == nil {
			dynamicTask.Inputs = map[string]*types.TypedValue{}
		}

		// Ensure that the outputted task depends on the creating task
		dynamicTask.Requires[taskId] = &types.TaskDependencyParameters{
			Type: types.TaskDependencyParameters_DYNAMIC_OUTPUT, // Marks this task as the output of the
		}

		log.WithFields(log.Fields{
			"id":          dynamicTask.Id,
			"functionRef": dynamicTask.FunctionRef,
		}).Debug("Adding dynamic task.")

		wi.Status.DynamicTasks[id] = dynamicTask
	}

	return nil
}
