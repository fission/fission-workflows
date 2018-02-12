package workflow

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/api/workflow/parse"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/fission/fission-workflows/pkg/util"
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
	// If no id is provided generate an id
	id := workflow.Id
	if len(id) == 0 {
		id = fmt.Sprintf("wf-%s", util.Uid())
	}

	data, err := proto.Marshal(workflow)
	if err != nil {
		return "", err
	}

	err = wa.es.Append(&fes.Event{
		Type:      events.Workflow_WORKFLOW_CREATED.String(),
		Aggregate: aggregates.NewWorkflowAggregate(id),
		Timestamp: ptypes.TimestampNow(),
		Data:      data,
	})
	if err != nil {
		return "", err
	}

	return id, nil
}

func (wa *Api) Delete(id string) error {
	return wa.es.Append(&fes.Event{
		Type:      events.Workflow_WORKFLOW_DELETED.String(),
		Aggregate: aggregates.NewWorkflowAggregate(id),
		Timestamp: ptypes.TimestampNow(),
		Hints:     &fes.EventHints{Completed: true},
	})
}

func (wa *Api) Parse(workflow *types.Workflow) (*types.WorkflowStatus, error) {
	parsed, err := wa.Resolver.Resolve(workflow.Spec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow: %v", err)
	}

	parsedData, err := proto.Marshal(parsed)
	if err != nil {
		return nil, err
	}

	err = wa.es.Append(&fes.Event{
		Type:      events.Workflow_WORKFLOW_PARSED.String(),
		Aggregate: aggregates.NewWorkflowAggregate(workflow.Metadata.Id),
		Timestamp: ptypes.TimestampNow(),
		Data:      parsedData,
	})
	if err != nil {
		return nil, err
	}

	return parsed, nil
}
