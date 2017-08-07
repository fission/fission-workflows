package workflow

import (
	"github.com/fission/fission-workflow/pkg/cache"
	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/projector/project"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/sirupsen/logrus"
	"strings"
)

type workflowProjector struct {
	esClient   eventstore.Client
	cache      cache.Cache // TODO ensure concurrent
	updateChan chan *eventstore.Event
}

func NewWorkflowProjector(esClient eventstore.Client, cache cache.Cache) project.WorkflowProjector {
	p := &workflowProjector{
		esClient:   esClient,
		cache:      cache,
		updateChan: make(chan *eventstore.Event),
	}
	go p.Run()
	return p
}

func (wp *workflowProjector) Get(subject string) (*types.Workflow, error) {
	cached := wp.getCache(subject)
	if cached != nil {
		return cached, nil
	}

	events, err := wp.esClient.Get("workflows." + subject) // TODO Fix hardcode subject
	if err != nil {
		return nil, err
	}

	var resultState *types.Workflow
	for _, event := range events {
		updatedState, err := wp.applyUpdate(event)
		if err != nil {
			logrus.Error(err)
		}
		resultState = updatedState
	}

	return resultState, nil
}

func (wp *workflowProjector) Watch(subject string) error {
	_, err := wp.esClient.Subscribe(&eventstore.SubscriptionConfig{
		Subject: subject,
		EventCh: wp.updateChan, // TODO provide clean channel that multiplexes into actual one
	})
	return err
}

func (wp *workflowProjector) Cache() cache.Cache {
	return wp.cache
}

func (wp *workflowProjector) Close() error {
	return nil
}

func (wp *workflowProjector) List(query string) ([]string, error) {

	subjects, err := wp.esClient.Subjects("workflows." + query) // TODO fix this hardcode
	if err != nil {
		return nil, err
	}
	// Strip the subject
	results := make([]string, len(subjects))
	for key, subject := range subjects {
		p := strings.SplitN(subject, ".", 2)
		results[key] = p[1] // TODO fix this hardcode
	}
	return results, nil
}

func (wp *workflowProjector) Run() {
	defer logrus.Debug("WorkflowProjector update loop has shutdown.")
	for event := range wp.updateChan {
		updatedState, err := wp.applyUpdate(event)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"err":          err,
				"updatedState": updatedState,
				"event":        event,
			}).Error("Failed to apply event to state")
		}
	}
}

func (wp *workflowProjector) getCache(subject string) *types.Workflow {
	raw, ok := wp.cache.Get(subject)
	if !ok {
		return nil
	}
	wf, ok := raw.(*types.Workflow)
	if !ok {
		logrus.Warnf("Cache contains invalid wf '%v'. Invalidating key.", raw)
		_ = wp.cache.Delete(subject)
	}
	return wf
}

func (wp *workflowProjector) applyUpdate(event *eventstore.Event) (*types.Workflow, error) {
	logrus.WithField("event", event).Debug("WorkflowProjector handling event.")
	invocationId := event.EventId.Subjects[1] // TODO fix hardcoded lookup

	currentState := wp.getCache(invocationId)
	if currentState == nil {
		currentState = Initial()
	}

	newState, err := Apply(*currentState, event)
	if err != nil {
		// TODO improve error handling (e.g. retry / replay)
		return nil, err
	}

	err = wp.cache.Put(invocationId, newState)
	if err != nil {
		return nil, err
	}
	return newState, nil
}
