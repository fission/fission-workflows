package pubsub

import (
	"io"
	"sync"

	"github.com/fission/fission-workflow/pkg/util/labels"
)

// A simple PubSub implementation
const (
	DEFAULT_SUB_CAP = 10
)

type SubscriptionOptions struct {
	Buf           int
	LabelSelector labels.Selector
}

type Subscription struct {
	SubscriptionOptions
	Ch chan interface{}
}

type Msg struct {
	Labels  labels.Labels
	Payload interface{}
}

type Publisher interface {
	io.Closer
	Subscribe(opts ...SubscriptionOptions) *Subscription
	Unsubscribe(sub *Subscription) error
	Publish(msg *Msg) error
}

func NewPublisher() Publisher {
	return &publisher{
		subs: []*Subscription{},
	}
}

type publisher struct {
	subs []*Subscription
	lock sync.Mutex
}

func (pu *publisher) Unsubscribe(sub *Subscription) error {
	pu.lock.Lock()
	defer pu.lock.Unlock()
	close(sub.Ch)
	updatedSubs := []*Subscription{}
	for _, s := range pu.subs {
		if sub != s {
			updatedSubs = append(updatedSubs, s)
		}
	}
	return nil
}

func (pu *publisher) Subscribe(opts ...SubscriptionOptions) *Subscription {
	pu.lock.Lock()
	defer pu.lock.Unlock()
	var subOpts SubscriptionOptions
	if len(opts) > 0 {
		subOpts = opts[0]
	}

	if subOpts.Buf <= 0 {
		subOpts.Buf = DEFAULT_SUB_CAP
	}

	sub := &Subscription{
		Ch:                  make(chan interface{}, subOpts.Buf),
		SubscriptionOptions: subOpts,
	}

	pu.subs = append(pu.subs, sub)
	return sub
}

func (pu *publisher) Publish(msg *Msg) error {
	pu.lock.Lock()
	defer pu.lock.Unlock()
	for _, sub := range pu.subs {
		if sub.LabelSelector != nil && !sub.LabelSelector.Matches(msg.Labels) {
			continue
		}
		select {
		case sub.Ch <- msg:
			// OK
		default:
			// Drop message if subscribers channel is full
			// Future: allow subscribers to specify in options what should happen when their channel is full.
		}
	}
	return nil
}

func (pu *publisher) Close() error {
	var err error
	for _, sub := range pu.subs {
		err = pu.Unsubscribe(sub)
	}
	return err
}
