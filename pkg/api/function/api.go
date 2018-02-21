package function

import (
	"github.com/fission/fission-workflows/pkg/api/dynamic"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

// TODO move events here

// Api that servers mainly as a function.Runtime wrapper that deals with the higher-level logic workflow-related logic.
type Api struct {
	runtime    map[string]fnenv.Runtime // TODO support AsyncRuntime
	es         fes.EventStore
	dynamicApi *dynamic.Api
}

func NewApi(runtime map[string]fnenv.Runtime, esClient fes.EventStore, api *dynamic.Api) *Api {
	return &Api{
		runtime:    runtime,
		es:         esClient,
		dynamicApi: api,
	}
}

func (ap *Api) Invoke(invocationId string, spec *types.TaskInvocationSpec) (*types.TaskInvocation, error) {
	aggregate := aggregates.NewWorkflowInvocationAggregate(invocationId)
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

	err = ap.es.Append(&fes.Event{
		Type:      events.Function_TASK_STARTED.String(),
		Parent:    aggregate,
		Aggregate: aggregates.NewTaskInvocationAggregate(id),
		Timestamp: ptypes.TimestampNow(),
		Data:      fnAny,
	})
	if err != nil {
		return nil, err
	}

	fnResult, err := ap.runtime[spec.FnRef.Runtime].Invoke(spec)
	if err != nil {
		// TODO improve error handling here (retries? internal or task related error?)
		logrus.WithField("task", invocationId).Infof("ParseTask failed: %v", err)
		esErr := ap.es.Append(&fes.Event{
			Type:      events.Function_TASK_FAILED.String(),
			Parent:    aggregate,
			Aggregate: aggregates.NewTaskInvocationAggregate(id),
			Timestamp: ptypes.TimestampNow(),
			Data:      fnAny,
		})
		if esErr != nil {
			return nil, esErr
		}
		return nil, err
	}

	// TODO to a middleware component
	if fnResult.Output != nil {
		switch fnResult.Output.Type {
		case typedvalues.TypeTask:
			logrus.Info("Adding dynamic task")
			taskSpec, err := typedvalues.FormatTask(fnResult.Output)
			if err != nil {
				return nil, err
			}

			// add task
			err = ap.dynamicApi.AddDynamicTask(invocationId, id, taskSpec)
			if err != nil {
				return nil, err
			}
		case typedvalues.TypeWorkflow:
			logrus.Info("Adding dynamic workflow")
			workflowSpec, err := typedvalues.FormatWorkflow(fnResult.Output)
			if err != nil {
				return nil, err
			}
			err = ap.dynamicApi.AddDynamicWorkflow(invocationId, id, workflowSpec)
			if err != nil {
				return nil, err
			}
		}
	}

	fn.Status = fnResult
	fnStatusAny, err := proto.Marshal(fn)
	if err != nil {
		return nil, err
	}

	if fnResult.Status == types.TaskInvocationStatus_SUCCEEDED {
		err = ap.es.Append(&fes.Event{
			Type:      events.Function_TASK_SUCCEEDED.String(),
			Parent:    aggregate,
			Aggregate: aggregates.NewTaskInvocationAggregate(id),
			Timestamp: ptypes.TimestampNow(),
			Data:      fnStatusAny,
		})
	} else {
		err = ap.es.Append(&fes.Event{
			Type:      events.Function_TASK_FAILED.String(),
			Parent:    aggregate,
			Aggregate: aggregates.NewTaskInvocationAggregate(id),
			Timestamp: ptypes.TimestampNow(),
			Data:      fnStatusAny,
		})
	}
	if err != nil {
		return nil, err
	}

	fn.Status = fnResult
	return fn, nil
}

func (ap *Api) Fail(invocationId string, taskId string) error {
	return ap.es.Append(&fes.Event{
		Type:      events.Function_TASK_FAILED.String(),
		Parent:    aggregates.NewWorkflowInvocationAggregate(invocationId),
		Aggregate: aggregates.NewTaskInvocationAggregate(taskId),
		Timestamp: ptypes.TimestampNow(),
	})
}
