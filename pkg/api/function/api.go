package function

import (
	"github.com/fission/fission-workflow/pkg/fes"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/aggregates"
	"github.com/fission/fission-workflow/pkg/types/events"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

// Api that servers mainly as a function.Runtime wrapper that deals with the higher-level logic workflow-related logic.
type Api struct {
	runtime map[string]Runtime // TODO support AsyncRuntime
	es      fes.EventStore
}

func NewFissionFunctionApi(runtime map[string]Runtime, esClient fes.EventStore) *Api {
	return &Api{
		runtime: runtime,
		es:      esClient,
	}
}

func (ap *Api) Invoke(invocationId string, spec *types.TaskInvocationSpec) (*types.TaskInvocation, error) {
	id := spec.TaskId // assumption: 1 task == 1 TaskInvocation (How to deal with retries? Same invocation?)
	fn := &types.TaskInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        id,
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec: spec,
	}

	fnAny, err := proto.Marshal(fn)
	if err != nil {
		return nil, err
	}

	err = ap.es.HandleEvent(&fes.Event{
		Type:      events.Function_TASK_STARTED.String(),
		Parent:    aggregates.NewWorkflowInvocationAggregate(invocationId),
		Aggregate: aggregates.NewTaskInvocationAggregate(id),
		Timestamp: ptypes.TimestampNow(),
		Data:      fnAny,
	})
	if err != nil {
		return nil, err
	}

	fnResult, err := ap.runtime[spec.Type.Runtime].Invoke(spec)
	if err != nil {
		// TODO record error message
		esErr := ap.es.HandleEvent(&fes.Event{
			Type:      events.Function_TASK_FAILED.String(),
			Parent:    aggregates.NewWorkflowInvocationAggregate(invocationId),
			Aggregate: aggregates.NewTaskInvocationAggregate(id),
			Timestamp: ptypes.TimestampNow(),
			Data:      fnAny,
		})
		if esErr != nil {
			return nil, esErr
		}
		return nil, err
	}
	fn.Status = fnResult
	fnStatusAny, err := proto.Marshal(fn)
	if err != nil {
		return nil, err
	}

	err = ap.es.HandleEvent(&fes.Event{
		Type:      events.Function_TASK_SUCCEEDED.String(),
		Parent:    aggregates.NewWorkflowInvocationAggregate(invocationId),
		Aggregate: aggregates.NewTaskInvocationAggregate(id),
		Timestamp: ptypes.TimestampNow(),
		Data:      fnStatusAny,
	})
	if err != nil {
		return nil, err
	}

	fn.Status = fnResult
	return fn, nil
}
