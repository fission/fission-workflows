package workflow

import (
	"fmt"

	"github.com/fission/fission-workflow/pkg/api/workflow/parse"
	"github.com/fission/fission-workflow/pkg/fes"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/aggregates"
	"github.com/fission/fission-workflow/pkg/types/events"
	"github.com/fission/fission-workflow/pkg/util"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
)

type Api struct {
	es       fes.EventStore
	Resolver *parse.Resolver
}

func NewApi(esClient fes.EventStore, parser *parse.Resolver) *Api {
	return &Api{esClient, parser}
}

func (wa *Api) Create(workflow *types.WorkflowSpec) (string, error) {
	id := util.Uid()

	data, err := proto.Marshal(workflow)
	if err != nil {
		return "", err
	}

	event := &fes.Event{
		Type:      events.Workflow_WORKFLOW_CREATED.String(),
		Aggregate: aggregates.NewWorkflowAggregate(id),
		Timestamp: ptypes.TimestampNow(),
		Data:      data,
	}

	err = wa.es.HandleEvent(event)
	if err != nil {
		return "", err
	}

	// TODO move this to controller or separate service or fission function, in order to make it more reliable
	// TODO more FT
	parsed, err := wa.Resolver.Resolve(workflow)
	if err != nil {
		return "", fmt.Errorf("failed to parse workflow: %v", err)
	}

	parsedData, err := proto.Marshal(parsed)
	if err != nil {
		return "", err
	}

	err = wa.es.HandleEvent(&fes.Event{
		Type:      events.Workflow_WORKFLOW_PARSED.String(),
		Aggregate: aggregates.NewWorkflowAggregate(id),
		Timestamp: ptypes.TimestampNow(),
		Data:      parsedData,
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

func (wa *Api) Delete(id string) error {
	return wa.es.HandleEvent(&fes.Event{
		Type:      events.Workflow_WORKFLOW_DELETED.String(),
		Aggregate: aggregates.NewWorkflowAggregate(id),
		Timestamp: ptypes.TimestampNow(),
	})
}
