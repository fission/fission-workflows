// package store provides typed, centralized access to the event-sourced workflow and invocation models
package store

import (
	"errors"
	"fmt"

	"github.com/fission/fission-workflows/pkg/api/aggregates"
	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/fission/fission-workflows/pkg/util/labels"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
)

type Workflows struct {
	fes.CacheReader // Currently needed for pubsub publisher interface, should be exposed here
}

func NewWorkflowsStore(workflows fes.CacheReader) *Workflows {
	return &Workflows{
		workflows,
	}
}

// GetWorkflow returns an event-sourced workflow.
// If an error occurred the error is returned, if no workflow was found both return values are nil.
func (s *Workflows) GetWorkflow(workflowID string) (*types.Workflow, error) {
	key := fes.NewAggregate(aggregates.TypeWorkflow, workflowID)
	entity, err := s.GetAggregate(key)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, nil
	}

	wf, ok := entity.(*aggregates.Workflow)
	if !ok {
		panic(fmt.Sprintf("aggregate type mismatch for key %s (expected: %T, got %T)", key.Format(),
			&aggregates.Workflow{}, wf))
	}

	return wf.Workflow, nil
}

// GetWorkflowNotifications returns a subscription to the updates of the workflow cache.
// Returns nil if the cache does not support pubsub.
//
// Future: Currently this assumes the presence of a pubsub.Publisher interface in the cache.
// In the future we can fallback to pull-based mechanisms
func (s *Workflows) GetWorkflowUpdates() *WorkflowSubscription {
	selector := labels.In(fes.PubSubLabelAggregateType, aggregates.TypeWorkflow)
	invokePub, ok := s.CacheReader.(pubsub.Publisher)
	if !ok {
		return nil
	}

	return &WorkflowSubscription{
		Subscription: invokePub.Subscribe(pubsub.SubscriptionOptions{
			Buffer:       fes.DefaultNotificationBuffer,
			LabelMatcher: selector,
		}),
		closeFn: func() error {
			return invokePub.Close()
		},
	}
}

type Invocations struct {
	fes.CacheReader
}

func NewInvocationStore(invocations fes.CacheReader) *Invocations {
	return &Invocations{
		invocations,
	}
}

// GetInvocation returns an event-sourced invocation.
// If an error occurred the error is returned, if no invocation was found both return values are nil.
func (s *Invocations) GetInvocation(invocationID string) (*types.WorkflowInvocation, error) {
	key := fes.NewAggregate(aggregates.TypeWorkflowInvocation, invocationID)
	entity, err := s.GetAggregate(key)
	if err != nil {
		return nil, err
	}
	if entity == nil {
		return nil, nil
	}

	wfi, ok := entity.(*aggregates.WorkflowInvocation)
	if !ok {
		panic(fmt.Sprintf("aggregate type mismatch for key %s (expected: %T, got %T - %v)", key.Format(),
			&aggregates.WorkflowInvocation{}, wfi, wfi))
	}

	return wfi.WorkflowInvocation, nil
}

// GetInvocationSubscription returns a subscription to the updates of the invocation cache.
// Returns nil if the cache does not support pubsub.
//
// Future: Currently this assumes the presence of a pubsub.Publisher interface in the cache.
// In the future we can fallback to pull-based mechanisms
func (s *Invocations) GetInvocationUpdates() *InvocationSubscription {
	selector := labels.In(fes.PubSubLabelAggregateType, aggregates.TypeWorkflowInvocation, aggregates.TypeTaskInvocation)
	invokePub, ok := s.CacheReader.(pubsub.Publisher)
	if !ok {
		return nil
	}

	return &InvocationSubscription{
		Subscription: invokePub.Subscribe(pubsub.SubscriptionOptions{
			Buffer:       fes.DefaultNotificationBuffer,
			LabelMatcher: selector,
		}),
		closeFn: func() error {
			return invokePub.Close()
		},
	}
}

type WorkflowSubscription struct {
	*pubsub.Subscription
	closeFn func() error
}

func (sub *WorkflowSubscription) ToNotification(msg pubsub.Msg) (*fes.Notification, error) {
	update, ok := msg.(*fes.Notification)
	if !ok {
		return nil, errors.New("received message is not a notification")
	}
	return update, nil
}

func (sub *WorkflowSubscription) Close() error {
	if sub.closeFn == nil {
		return nil
	}
	return sub.closeFn()
}

type InvocationSubscription struct {
	*pubsub.Subscription
	closeFn func() error
}

func (sub *InvocationSubscription) ToNotification(msg pubsub.Msg) (*fes.Notification, error) {
	update, ok := msg.(*fes.Notification)
	if !ok {
		return nil, errors.New("received message is not a notification")
	}
	return update, nil
}

func (sub *InvocationSubscription) Close() error {
	if sub.closeFn == nil {
		return nil
	}
	return sub.closeFn()
}

func ParseNotificationToWorkflow(update *fes.Notification) (*types.Workflow, error) {
	entity, ok := update.Payload.(*aggregates.Workflow)
	if !ok {
		return nil, errors.New("received message does not include workflow invocation as payload")
	}
	return entity.Workflow, nil
}

func ParseNotificationToInvocation(update *fes.Notification) (*types.WorkflowInvocation, error) {
	entity, ok := update.Payload.(*aggregates.WorkflowInvocation)
	if !ok {
		return nil, errors.New("received message does not include workflow invocation as payload")
	}
	return entity.WorkflowInvocation, nil
}
