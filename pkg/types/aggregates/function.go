package aggregates

import (
	"fmt"

	"github.com/fission/fission-workflow/pkg/fes"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/events"
	"github.com/gogo/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

const (
	TYPE_FUNCTION_INVOCATION = "function"
)

type FunctionInvocation struct {
	*fes.AggregatorMixin
	*types.FunctionInvocation
}

func NewFunctionInvocation(id string, fi *types.FunctionInvocation) *FunctionInvocation {
	fia := &FunctionInvocation{
		FunctionInvocation: fi,
	}

	fia.AggregatorMixin = fes.NewAggregatorMixin(fia, *NewFunctionInvocationAggregate(id))

	return fia
}

func NewFunctionInvocationAggregate(id string) *fes.Aggregate {
	return &fes.Aggregate{
		Id:   id,
		Type: TYPE_FUNCTION_INVOCATION,
	}
}

func (fi *FunctionInvocation) ApplyEvent(event *fes.Event) error {

	eventType, err := events.ParseFunction(event.Type)
	if err != nil {
		return err
	}

	switch eventType {
	case events.Function_TASK_STARTED:
		fn := &types.FunctionInvocation{}
		err = proto.Unmarshal(event.Data, fn)
		if err != nil {
			return fmt.Errorf("failed to unmarshal event: '%v' (%v)", event, err)
		}

		fi.FunctionInvocation = &types.FunctionInvocation{
			Metadata: fn.Metadata,
			Spec:     fn.Spec,
			Status: &types.FunctionInvocationStatus{
				Status:    types.FunctionInvocationStatus_SCHEDULED,
				UpdatedAt: event.Timestamp,
			},
		}
	case events.Function_TASK_SUCCEEDED:
		fn := &types.FunctionInvocation{}
		err = proto.Unmarshal(event.Data, fn)
		if err != nil {
			return fmt.Errorf("failed to unmarshal event: '%v' (%v)", event, err)
		}

		fi.Status.Output = fn.Status.Output
		fi.Status.Status = types.FunctionInvocationStatus_SUCCEEDED
		fi.Status.UpdatedAt = event.Timestamp
	case events.Function_TASK_ABORTED:
		fi.Status.Status = types.FunctionInvocationStatus_ABORTED
		fi.Status.UpdatedAt = event.Timestamp
	case events.Function_TASK_FAILED:
		fi.Status.Status = types.FunctionInvocationStatus_FAILED
		fi.Status.UpdatedAt = event.Timestamp
	case events.Function_TASK_SKIPPED:
		fi.Status.Status = types.FunctionInvocationStatus_SKIPPED
		fi.Status.UpdatedAt = event.Timestamp
	case events.Function_TASK_HEARTBEAT_REQUEST:
		panic("NA")
	case events.Function_TASK_HEARTBEAT_RESPONSE:
		panic("NA")
	default:
		log.WithFields(log.Fields{
			"event": event,
		}).Warn("Skipping unimplemented event.")
	}
	return nil
}
