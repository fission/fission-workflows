package workflow

import (
	"fmt"

	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/eventstore/eventids"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/events"
	"github.com/fission/fission-workflow/pkg/util"
	"github.com/golang/protobuf/ptypes"
)

type Api struct {
	esClient eventstore.Client
	Parser   *Parser
}

func NewApi(esClient eventstore.Client, parser *Parser) *Api {
	return &Api{esClient, parser}
}

func (wa *Api) Create(workflow *types.WorkflowSpec) (string, error) {

	id := util.Uid()

	data, err := ptypes.MarshalAny(workflow)
	if err != nil {
		return "", err
	}

	eventId := eventids.NewSubject(types.SUBJECT_WORKFLOW, id)
	event := eventstore.NewEvent(eventId, events.Workflow_WORKFLOW_CREATED.String(), data)
	err = wa.esClient.Append(event)
	if err != nil {
		return "", err
	}

	// TODO move this to controller or seperate service or fission function, in order to make it more reliable
	// TODO more FT
	parsed, err := wa.Parser.Parse(workflow)
	if err != nil {
		return "", fmt.Errorf("Failed to parse workflow: %v", err)
	}

	parsedData, err := ptypes.MarshalAny(parsed)
	if err != nil {
		return "", err
	}

	parsedEvent := eventstore.NewEvent(eventId, events.Workflow_WORKFLOW_PARSED.String(), parsedData)
	err = wa.esClient.Append(parsedEvent)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (wa *Api) Delete(id string) error {

	eventId := eventids.NewSubject(types.SUBJECT_WORKFLOW, id)

	event := eventstore.NewEvent(eventId, events.Workflow_WORKFLOW_DELETED.String(), nil)

	err := wa.esClient.Append(event)
	if err != nil {
		return err
	}
	return nil
}
