package function

import (
	"github.com/fission/fission-workflow/pkg/api"
	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/eventstore/eventids"
	"github.com/fission/fission-workflow/pkg/eventstore/events"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/util"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

// Responsible for executing functions
type Api struct {
	runtime  api.FunctionRuntimeEnv
	esClient eventstore.Client
}

func NewFissionFunctionApi(runtime api.FunctionRuntimeEnv, esClient eventstore.Client) *Api {
	return &Api{
		runtime:  runtime,
		esClient: esClient,
	}
}

func (ap *Api) InvokeSync(invocationId string, fnSpec *types.FunctionInvocationSpec) (*types.FunctionInvocation, error) {
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

	fnResult, err := ap.runtime.InvokeSync(fnSpec) // TODO spec or container?
	if err != nil {
		failedEvent := events.New(eventid, types.InvocationEvent_TASK_FAILED.String(), fnAny) // TODO record error message
		err = ap.esClient.Append(failedEvent)
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

func (ap *Api) Invoke(invocationId string, spec *types.FunctionInvocationSpec) (string, error) {
	panic("implement me")
}

func (ap *Api) Cancel(fnInvocationId string) error {
	panic("implement me")
}

func (ap *Api) Status(fnInvocationId string) (*types.FunctionInvocationStatus, error) {
	panic("implement me")
}
