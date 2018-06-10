// Package pubsub is a simple, label-based, thread-safe PubSub implementation.
package pubsub

import (
	"io"
	"sync"
	"time"

	"github.com/fission/fission-workflows/pkg/util/labels"
)

const (
	defaultSubscriptionBuffer = 10
)

type Msg interface {
	Labels() labels.Labels
	CreatedAt() time.Time
}

type Publisher interface {
	io.Closer
	Subscribe(opts ...SubscriptionOptions) *Subscription
	Unsubscribe(sub *Subscription) error
	Publish(msg Msg) error
}

// SubscriptionOptions allow subscribers to customize the type and behaviour of the subscription.
type SubscriptionOptions struct {
	// Buffer is the size of the buffer kept for incoming messages. In case the buffer is full, subsequent messages will
	// be dropped. The default size of the buffer is 10.
	Buffer int

	// LabelMatcher allows subscribers to narrow the selection of messages that they are notified of.
	// By default there is no matcher; the subscriber will receive all messages published by the publisher.
	LabelMatcher labels.Matcher
}

type Subscription struct {
	SubscriptionOptions
	Ch chan Msg
}

type EmptyMsg struct {
	labels    labels.Labels
	createdAt time.Time
}

func (gm *EmptyMsg) Labels() labels.Labels {
	return gm.labels
}

func (gm *EmptyMsg) CreatedAt() time.Time {
	return gm.createdAt
}

type GenericMsg struct {
	*EmptyMsg
	payload interface{}
}

func NewEmptyMsg(lbls labels.Labels, createdAt time.Time) *EmptyMsg {
	return &EmptyMsg{
		labels:    lbls,
		createdAt: createdAt,
	}
}

func NewGenericMsg(lbls labels.Labels, createdAt time.Time, payload interface{}) *GenericMsg {
	return &GenericMsg{
		EmptyMsg: NewEmptyMsg(lbls, createdAt),
		payload:  payload,
	}
}

func (pm *GenericMsg) Payload() interface{} {
	return pm.payload
}

func NewPublisher() *DefaultPublisher {
	return &DefaultPublisher{
		subs: []*Subscription{},
	}
}

type DefaultPublisher struct {
	subs []*Subscription
	lock sync.Mutex
}

func (pu *DefaultPublisher) Unsubscribe(sub *Subscription) error {
	pu.lock.Lock()
	defer pu.lock.Unlock()
	close(sub.Ch)
	var updatedSubs []*Subscription
	for _, s := range pu.subs {
		if sub != s {
			updatedSubs = append(updatedSubs, s)
		}
	}
	return nil
}

func (pu *DefaultPublisher) Subscribe(opts ...SubscriptionOptions) *Subscription {
	pu.lock.Lock()
	defer pu.lock.Unlock()
	var subOpts SubscriptionOptions
	if len(opts) > 0 {
		subOpts = opts[0]
	}

	if subOpts.Buffer <= 0 {
		subOpts.Buffer = defaultSubscriptionBuffer
	}

	sub := &Subscription{
		Ch:                  make(chan Msg, subOpts.Buffer),
		SubscriptionOptions: subOpts,
	}

	pu.subs = append(pu.subs, sub)
	return sub
}

func (pu *DefaultPublisher) Publish(msg Msg) error {
	pu.lock.Lock()
	defer pu.lock.Unlock()
	for _, sub := range pu.subs {
		if sub.LabelMatcher != nil && !sub.LabelMatcher.Matches(msg.Labels()) {
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

func (pu *DefaultPublisher) Close() error {
	var err error
	for _, sub := range pu.subs {
		err = pu.Unsubscribe(sub)
	}
	return err
}
