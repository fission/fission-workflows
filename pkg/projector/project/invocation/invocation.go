package invocation

import (
	"errors"

	"fmt"

	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/events"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

/*
	Invocation Projection
*/
type reduceFunc func(invocation types.WorkflowInvocation, event *eventstore.Event) (newState *types.WorkflowInvocation, err error)

var eventMapping = map[events.Invocation]reduceFunc{
	events.Invocation_INVOCATION_CREATED:   created,       // Data: WorkflowInvocation
	events.Invocation_INVOCATION_CANCELED:  canceled,      // Data: {}
	events.Invocation_INVOCATION_COMPLETED: completed,     // Data: {}
	events.Invocation_TASK_STARTED:         taskStarted,   // Data: FunctionInvocation
	events.Invocation_TASK_FAILED:          taskFailed,    // Data: FunctionInvocation
	events.Invocation_TASK_SUCCEEDED:       taskSucceeded, // Data: FunctionInvocation
	events.Invocation_TASK_ABORTED:         taskAborted,
}

func Initial() *types.WorkflowInvocation {
	return &types.WorkflowInvocation{}
}

func From(events ...*eventstore.Event) (currentState *types.WorkflowInvocation, err error) {
	return Apply(*Initial(), events...)
}

func Apply(currentState types.WorkflowInvocation, seqEvents ...*eventstore.Event) (newState *types.WorkflowInvocation, err error) {
	// Check if it is indeed next event (maybe wrap in a projectionContainer)
	newState = &currentState
	for _, event := range seqEvents {

		eventType, err := events.ParseInvocation(event.GetType())
		if err != nil {
			return nil, err
		}
		newState, err = eventMapping[eventType](currentState, event)
		if err != nil {
			return nil, err
		}
	}

	return newState, nil
}

func created(currentState types.WorkflowInvocation, event *eventstore.Event) (newState *types.WorkflowInvocation, err error) {
	// Check if state
	if currentState != (types.WorkflowInvocation{}) {
		//return nil, fmt.Errorf("invalid event '%v' for state '%v'", event, currentState) // TODO fix errors
		logrus.Warnf("invalid event '%v' for state '%v'", event, currentState)
		return &currentState, nil
	}

	spec := &types.WorkflowInvocationSpec{}
	err = ptypes.UnmarshalAny(event.Data, spec)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal event: '%v' (%v)", event, err)
	}

	return &types.WorkflowInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        event.GetEventId().GetSubjects()[1], // TODO remove this hardcoding,
			CreatedAt: event.GetTime(),
		},
		Spec: spec,
		Status: &types.WorkflowInvocationStatus{
			Status:    types.WorkflowInvocationStatus_IN_PROGRESS,
			Tasks:     map[string]*types.FunctionInvocation{},
			UpdatedAt: event.Time,
		},
	}, nil
}

func canceled(currentState types.WorkflowInvocation, event *eventstore.Event) (newState *types.WorkflowInvocation, err error) {
	// Canceling non-existent / already invocationCanceled invocation does nothing
	if currentState == (types.WorkflowInvocation{}) {
		return nil, errors.New("Unknown state") // TODO fix errors
	}
	currentState.GetStatus().Status = types.WorkflowInvocationStatus_ABORTED
	currentState.GetStatus().UpdatedAt = event.Time
	return &currentState, nil
}

func completed(currentState types.WorkflowInvocation, event *eventstore.Event) (newState *types.WorkflowInvocation, err error) {
	status := &types.WorkflowInvocationStatus{}
	err = ptypes.UnmarshalAny(event.Data, status)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal event: '%v' (%v)", event, err)
	}

	currentState.GetStatus().Status = types.WorkflowInvocationStatus_SUCCEEDED
	currentState.GetStatus().UpdatedAt = event.Time
	currentState.GetStatus().Output = status.Output
	return &currentState, nil
}

func skip(currentState types.WorkflowInvocation, event *eventstore.Event) (newState *types.WorkflowInvocation, err error) {
	logrus.WithFields(logrus.Fields{
		"currentState": currentState,
		"event":        event,
	}).Debug("Skipping unimplemented event.")
	return &currentState, nil
}

func taskStarted(currentState types.WorkflowInvocation, event *eventstore.Event) (newState *types.WorkflowInvocation, err error) {
	fn := &types.FunctionInvocation{}
	err = ptypes.UnmarshalAny(event.Data, fn)
	if err != nil {
		return nil, fmt.Errorf("Failed to unmarshal event: '%v' (%v)", event, err)
	}

	if existing, ok := currentState.GetStatus().Tasks[fn.Spec.TaskId]; ok {
		logrus.Debug("Overwriting existing function invocation '%v'", existing)
	}

	currentState.GetStatus().Tasks[fn.Spec.TaskId] = &types.FunctionInvocation{
		Metadata: fn.Metadata,
		Spec:     fn.Spec,
		Status: &types.FunctionInvocationStatus{
			Status:    types.FunctionInvocationStatus_SCHEDULED,
			UpdatedAt: event.Time,
		},
	}
	return &currentState, nil
}

func taskFailed(currentState types.WorkflowInvocation, event *eventstore.Event) (newState *types.WorkflowInvocation, err error) {
	fn := &types.FunctionInvocation{}
	err = ptypes.UnmarshalAny(event.Data, fn)

	logrus.WithFields(logrus.Fields{
		"fn":           fn,
		"currentState": currentState,
		"event":        event,
	}).Infof(">>>")

	existing, ok := currentState.GetStatus().Tasks[fn.Spec.TaskId]
	if !ok {
		logrus.Warn("Non-recorded function failed '%v'", fn)
		return &currentState, nil
	}

	existing.Status.Status = types.FunctionInvocationStatus_FAILED
	existing.Status.UpdatedAt = event.Time

	return &currentState, nil
}

func taskSucceeded(currentState types.WorkflowInvocation, event *eventstore.Event) (newState *types.WorkflowInvocation, err error) {
	fn := &types.FunctionInvocation{}
	err = ptypes.UnmarshalAny(event.Data, fn)

	fnExisting, ok := currentState.GetStatus().Tasks[fn.Spec.TaskId]
	if !ok {
		logrus.Warn("Non-recorded function succeeded '%v'", fn)
		return &currentState, nil
	}

	fnExisting.Status.Status = types.FunctionInvocationStatus_SUCCEEDED
	fnExisting.Status.UpdatedAt = event.Time
	fnExisting.Status.Output = fn.Status.Output

	// If a task returned a flow, restructure the status
	//if typedvalues.IsFormat(fn.Status.Output.Type, typedvalues.TYPE_FLOW) {
	//
	//	flow, err := typedvalues.NewDefaultParserFormatter().Format(fn.Status.Output)
	//	if err != nil {
	//		return nil, err
	//	}
	//
	//	outputTask, ok := flow.(*types.Task)
	//	if ok {
	//		return nil, err
	//	}
	//
	//	// TODO Copy src task dependencies
	//	// TODO limit scope
	//	outputTask.Dependencies = map[string]*types.TaskDependencyParameters{
	//		taskId: nil,
	//	}
	//	outputTask.DependenciesAwait = 1
	//
	//	// Generate a name
	//	outputTaskId := fmt.Sprintf("%s_generated", taskId)
	//
	//	invocation.GetStatus().Workflow.Tasks[outputTaskId] = outputTask
	//	// TODO change dependencies in other tasks to depend on the new task
	//	for _, task := range invocation.GetStatus().Workflow.Tasks {
	//		dependee := false
	//		for depKey := range task.Dependencies {
	//			if depKey == taskId {
	//				dependee = true
	//			}
	//		}
	//		if dependee {
	//			task.Dependencies[outputTaskId] = nil
	//		}
	//	}
	//}

	return &currentState, nil
}

func taskAborted(currentState types.WorkflowInvocation, event *eventstore.Event) (newState *types.WorkflowInvocation, err error) {
	fn := &types.FunctionInvocation{}
	err = ptypes.UnmarshalAny(event.Data, fn)

	existing, ok := currentState.GetStatus().Tasks[fn.Spec.TaskId]
	if !ok {
		logrus.Warn("Non-recorded function aborted '%v'", fn)
		return &currentState, nil
	}

	existing.Status.Status = types.FunctionInvocationStatus_ABORTED
	existing.Status.UpdatedAt = event.Time

	return &currentState, nil
}
