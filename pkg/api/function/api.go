package function

import (
	"errors"

	"github.com/fission/fission-workflows/pkg/api/dynamic"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fnenv"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

// TODO move events here

// Api that servers mainly as a function.Runtime wrapper that deals with the higher-level logic workflow-related logic.
type Api struct {
	runtime    map[string]fnenv.Runtime // TODO support AsyncRuntime
	es         fes.Backend
	dynamicApi *dynamic.Api
}

func NewApi(runtime map[string]fnenv.Runtime, esClient fes.Backend, api *dynamic.Api) *Api {
	return &Api{
		runtime:    runtime,
		es:         esClient,
		dynamicApi: api,
	}
}

func (ap *Api) Invoke(spec *types.TaskInvocationSpec) (*types.TaskInvocation, error) {
	err := validate.TaskInvocationSpec(spec)
	if err != nil {
		return nil, err
	}

	aggregate := aggregates.NewWorkflowInvocationAggregate(spec.InvocationId)
	taskId := spec.TaskId // assumption: 1 task == 1 TaskInvocation (How to deal with retries? Same invocation?)
	fn := &types.TaskInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        taskId,
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec: spec,
	}

	fnAny, err := proto.Marshal(fn)
	if err != nil {
		return nil, err
	}

	err = ap.es.Append(&fes.Event{
		Type:      events.Task_TASK_STARTED.String(),
		Parent:    aggregate,
		Aggregate: aggregates.NewTaskInvocationAggregate(taskId),
		Timestamp: ptypes.TimestampNow(),
		Data:      fnAny,
	})
	if err != nil {
		return nil, err
	}

	fnResult, err := ap.runtime[spec.FnRef.Runtime].Invoke(spec)
	if fnResult == nil && err == nil {
		err = errors.New("function crashed")
	}
	if err != nil {
		// TODO improve error handling here (retries? internal or task related error?)
		logrus.WithField("task", spec.InvocationId).Infof("ParseTask failed: %v", err)
		esErr := ap.es.Append(&fes.Event{
			Type:      events.Task_TASK_FAILED.String(),
			Parent:    aggregate,
			Aggregate: aggregates.NewTaskInvocationAggregate(taskId),
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
			err = ap.dynamicApi.AddDynamicTask(spec.InvocationId, taskId, taskSpec)
			if err != nil {
				return nil, err
			}
		case typedvalues.TypeWorkflow:
			logrus.Info("Adding dynamic workflow")
			workflowSpec, err := typedvalues.FormatWorkflow(fnResult.Output)
			if err != nil {
				return nil, err
			}
			err = ap.dynamicApi.AddDynamicWorkflow(spec.InvocationId, taskId, workflowSpec)
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
			Type:      events.Task_TASK_SUCCEEDED.String(),
			Parent:    aggregate,
			Aggregate: aggregates.NewTaskInvocationAggregate(taskId),
			Timestamp: ptypes.TimestampNow(),
			Data:      fnStatusAny,
		})
	} else {
		err = ap.es.Append(&fes.Event{
			Type:      events.Task_TASK_FAILED.String(),
			Parent:    aggregate,
			Aggregate: aggregates.NewTaskInvocationAggregate(taskId),
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
		Type:      events.Task_TASK_FAILED.String(),
		Parent:    aggregates.NewWorkflowInvocationAggregate(invocationId),
		Aggregate: aggregates.NewTaskInvocationAggregate(taskId),
		Timestamp: ptypes.TimestampNow(),
	})
}
