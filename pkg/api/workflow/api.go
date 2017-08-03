package workflow

import (
	"github.com/fission/fission-workflow/pkg/cache"
	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/eventstore/eventids"
	"github.com/fission/fission-workflow/pkg/eventstore/events"
	"github.com/fission/fission-workflow/pkg/projector/project"
	"github.com/fission/fission-workflow/pkg/projector/project/workflow"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/util"
	"github.com/golang/protobuf/ptypes"
	"github.com/sirupsen/logrus"
)

type Api struct {
	esClient  eventstore.Client
	Projector project.WorkflowProjector
	Parser    *Parser
}

func NewApi(esClient eventstore.Client, parser *Parser) *Api {
	projector := workflow.NewWorkflowProjector(esClient, cache.NewMapCache()) // TODO move to arguments
	return &Api{esClient, projector, parser}
}

func (wa *Api) Create(workflow *types.WorkflowSpec) (string, error) {

	id := util.Uid()

	data, err := ptypes.MarshalAny(workflow)
	if err != nil {
		return "", err
	}

	eventId := eventids.NewSubject(types.SUBJECT_WORKFLOW, id)
	event := events.New(eventId, types.WorkflowEvent_WORKFLOW_CREATED.String(), data)
	err = wa.esClient.Append(event)
	if err != nil {
		return "", err
	}

	// TODO move this to controller or seperate service or fission function, in order to make it more reliable
	// TODO more FT
	parsed, err := wa.Parser.Parse(workflow)
	if err != nil {
		logrus.WithField("id", eventId).Errorf("Failed to parse workflow: %v", err)
	}

	parsedData, err := ptypes.MarshalAny(parsed)
	if err != nil {
		return "", err
	}

	parsedEvent := events.New(eventId, types.WorkflowEvent_WORKFLOW_PARSED.String(), parsedData)
	err = wa.esClient.Append(parsedEvent)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (wa *Api) Delete(id string) error {

	eventId := eventids.NewSubject(types.SUBJECT_WORKFLOW, id)

	event := events.New(eventId, types.WorkflowEvent_WORKFLOW_CREATED.String(), nil)

	err := wa.esClient.Append(event)
	if err != nil {
		return err
	}
	return nil
}

func (wa *Api) Get(id string) (*types.Workflow, error) {
	return wa.Projector.Get(id)
}

// TODO Support queries
// TODO support filtering; not fetching all
// Lists all the ids of all workflows, watched by the projector
func (wa *Api) List(query string) ([]string, error) {
	return wa.Projector.List(query)
}
