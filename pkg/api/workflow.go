package api

import (
	"github.com/fission/fission-workflow/pkg/cache"
	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/eventstore/eventids"
	"github.com/fission/fission-workflow/pkg/eventstore/events"
	"github.com/fission/fission-workflow/pkg/projector/project"
	"github.com/fission/fission-workflow/pkg/projector/project/workflow"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/golang/protobuf/ptypes"
	"github.com/satori/go.uuid"
)

const (
	WORKFLOW_SUBJECT = "workflows"
)

type WorkflowApi struct {
	esClient  eventstore.Client
	Projector project.WorkflowProjector
}

func NewWorkflowApi(esClient eventstore.Client) *WorkflowApi {
	return &WorkflowApi{esClient, workflow.NewWorkflowProjector(esClient, cache.NewMapCache())}
}

func (wa *WorkflowApi) Create(workflow *types.WorkflowSpec) (string, error) {
	// TODO validation
	id := uuid.NewV4().String()

	data, err := ptypes.MarshalAny(workflow)
	if err != nil {
		return "", err
	}

	eventId := eventids.NewSubject(WORKFLOW_SUBJECT, id)
	event := events.New(eventId, types.WorkflowEvent_WORKFLOW_CREATED.String(), data)
	err = wa.esClient.Append(event)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (wa *WorkflowApi) Delete(id string) error {
	// TODO validation
	eventId := eventids.NewSubject(WORKFLOW_SUBJECT, id)

	event := events.New(eventId, types.WorkflowEvent_WORKFLOW_CREATED.String(), nil)

	err := wa.esClient.Append(event)
	if err != nil {
		return err
	}
	return nil
}

func (wa *WorkflowApi) Get(id string) (*types.Workflow, error) {
	return wa.Projector.Get(id)
}

// TODO Support queries
// TODO support filtering; not fetching all
// Lists all the ids of all workflows, watched by the projector
func (wa *WorkflowApi) List(query string) ([]string, error) {
	return wa.Projector.List(query)
}
