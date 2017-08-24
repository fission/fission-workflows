package project

import (
	"github.com/fission/fission-workflow/pkg/cache"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/util/pubsub"
)

type WorkflowProjector interface {
	pubsub.Publisher

	// Get projection from cache or attempt to replay it.
	Get(subject string) (*types.Workflow, error)

	Cache() cache.Cache

	// Replays events, if it already exists, it is invalidated and replayed
	// Populates cache
	Watch(query string) error

	// Lists all subjects that fit the query
	List(query string) ([]string, error)
}

type InvocationProjector interface {
	pubsub.Publisher

	// Get projection from cache or attempt to replay it.
	Get(subject string) (*types.WorkflowInvocation, error)

	Cache() cache.Cache

	// Replays events, if it already exists, it is invalidated and replayed
	// Populates cache
	// TODO could be moved to constructor
	Watch(query string) error

	// Lists all subjects that fit the query
	List(query string) ([]string, error)
}
