package projector

import (
	"time"

	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/types"
)

type Projection interface {
	// It does not make assumptions about the sequence of events.
	Apply(state *types.WorkflowInvocationContainer, event ...*eventstore.Event) error

	Initial() *types.WorkflowInvocationContainer
}

// Per object type view only!!!
type Projector interface {
	// Get projection from cache or attempt to replay it.
	Get(subject string) (*types.WorkflowInvocationContainer, error)

	// Replays events, if it already exists, it is invalidated and replayed
	Fetch(subject string) error

	// Suscribe to updates in this projector
	Subscribe(chan *InvocationNotification) error
}

// In order to avoid leaking eventstore details
type InvocationNotification struct {
	Id   string
	Data *types.WorkflowInvocationContainer
	Type types.InvocationEvent
	Time *time.Time
}
