package project

import (
	"io"
	"time"

	"github.com/fission/fission-workflow/pkg/cache"
	"github.com/fission/fission-workflow/pkg/types"
)

type WorkflowProjector interface {
	io.Closer
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
	io.Closer
	// Get projection from cache or attempt to replay it.
	Get(subject string) (*types.WorkflowInvocation, error)

	Cache() cache.Cache

	// Replays events, if it already exists, it is invalidated and replayed
	// Populates cache
	Watch(query string) error

	// Lists all subjects that fit the query
	List(query string) ([]string, error)

	// Suscribe to updates on subjects watched by this projector
	Subscribe(updateCh chan *InvocationNotification) error
}

// TODO generify
// In order to avoid leaking eventstore details
type InvocationNotification struct {
	Id   string
	Data *types.WorkflowInvocation
	Type types.InvocationEvent
	Time time.Time
}
