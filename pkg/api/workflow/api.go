package workflow

import (
	"fmt"

	"github.com/fission/fission-workflows/pkg/api/workflow/parse"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/types/aggregates"
	"github.com/fission/fission-workflows/pkg/types/events"
	"github.com/fission/fission-workflows/pkg/types/validate"
	"github.com/fission/fission-workflows/pkg/util"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

type Api struct {
	es       fes.EventStore
	Resolver *parse.Resolver
}

func NewApi(esClient fes.EventStore, parser *parse.Resolver) *Api {
	return &Api{esClient, parser}
}

// TODO check if id already exists
func (wa *Api) Create(workflow *types.WorkflowSpec) (string, error) {
	err := validate.WorkflowSpec(workflow)
	if err != nil {
		logrus.Info(validate.Format(err))
		return "", err
	}

	// If no id is provided generate an id
	id := workflow.ForceId
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
	if err := validate.WorkflowSpec(workflow.Spec); err != nil {
		return nil, err
	}

	resolvedFns, err := wa.Resolver.ResolveMap(workflow.Spec.Tasks)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow: %v", err)
	}

	wfStatus := types.NewWorkflowStatus()
	for id, t := range workflow.Spec.Tasks {
		resolved := resolvedFns[t.FunctionRef]

		wfStatus.AddTaskStatus(id, &types.TaskStatus{
			UpdatedAt: ptypes.TimestampNow(),
			Resolved:  resolved,
			Status:    types.TaskStatus_READY,
		})
	}

	parsedData, err := proto.Marshal(wfStatus)
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

	return wfStatus, nil
}
