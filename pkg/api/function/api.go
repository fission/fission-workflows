package function

import (
	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/eventstore/eventids"
	"github.com/fission/fission-workflow/pkg/eventstore/events"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/util"
	"github.com/golang/protobuf/ptypes"
)

// A function.Runtime wrapper that deals with the higher-level logic workflow-related logic
type Api struct {
	runtime  Runtime // TODO support async
	esClient eventstore.Client
}

func NewFissionFunctionApi(runtime Runtime, esClient eventstore.Client) *Api {
	return &Api{
		runtime:  runtime,
		esClient: esClient,
	}
}

func (ap *Api) Invoke(invocationId string, fnSpec *types.FunctionInvocationSpec) (*types.FunctionInvocation, error) {
	eventid := eventids.NewSubject(types.SUBJECT_INVOCATION, invocationId)

	fn := &types.FunctionInvocation{
		Metadata: &types.ObjectMetadata{
			Id:        util.Uid(),
			CreatedAt: ptypes.TimestampNow(),
		},
		Spec: fnSpec,
	}

	fnAny, err := ptypes.MarshalAny(fn)
	if err != nil {
		panic(err)
	}

	startEvent := events.New(eventid, types.InvocationEvent_TASK_STARTED.String(), fnAny)

	err = ap.esClient.Append(startEvent)
	if err != nil {
		return nil, err
	}

	fnResult, err := ap.runtime.Invoke(fnSpec) // TODO spec or container?
	if err != nil {
		failedEvent := events.New(eventid, types.InvocationEvent_TASK_FAILED.String(), fnAny) // TODO record error message
		esErr := ap.esClient.Append(failedEvent)
		if esErr != nil {
			return nil, esErr
		}
		return nil, err
	}
	fn.Status = fnResult
	fnStatusAny, err := ptypes.MarshalAny(fn)
	if err != nil {
		return nil, err
	}

	succeededEvent := events.New(eventid, types.InvocationEvent_TASK_SUCCEEDED.String(), fnStatusAny)
	err = ap.esClient.Append(succeededEvent)
	if err != nil {
		return nil, err
	}

	fn.Status = fnResult
	return fn, nil
}
