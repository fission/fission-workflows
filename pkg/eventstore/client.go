package eventstore

import (
	"io"
)

// Client V2
type Client interface {
	io.Closer
	// Replays all events for subject, and continues to watch afterwards.
	Subscribe(config *SubscriptionConfig) (Subscription, error)

	Get(subject string) ([]*Event, error) // Stan.StartAtSequence // Convenience function, uses subscribe in the background

	Append(event *Event) error

	// To replay existing: subscription.close() -> subscribe(subscription.Config())
	Replay(subject string, subscription Subscription) error
}

type Subscription interface {
	io.Closer
	Config() *SubscriptionConfig
}

// TODO Make concurrency-proof
type SubscriptionConfig struct {
	Subject string
	EventCh chan *Event
	//StartAt  string // Enum: FROM(0)...NOW
	//FinishAt string // Enum: FROM(0)...NOW
}
