package projector

import (
	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/util/pubsub"
)

type Projection interface {
	// It does not make assumptions about the sequence of events.
	Apply(state *types.WorkflowInvocation, event ...*eventstore.Event) error

	Initial() *types.WorkflowInvocation
}

// Per object type view only!!!
type Projector interface {
	pubsub.Publisher

	// Get projection from cache or attempt to replay it.
	Get(subject string) (*types.WorkflowInvocation, error)

	// Replays events, if it already exists, it is invalidated and replayed
	Fetch(subject string) error
}
