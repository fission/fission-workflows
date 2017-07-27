package nats

import (
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
)

// Wrapper of 'stan' package to augment the API
type Conn struct {
	conn stan.Conn
}

func NewConn(conn stan.Conn) *Conn {
	return &Conn{conn}
}

func (sm *Conn) Publish(subject string, data []byte) error {
	return sm.conn.Publish(subject, data)
}

func (sm *Conn) PublishAsync(subject string, data []byte, ah stan.AckHandler) (string, error) {
	return sm.conn.PublishAsync(subject, data, ah)
}

func (sm *Conn) Subscribe(subject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return sm.conn.Subscribe(subject, cb, opts...)
}

func (sm *Conn) QueueSubscribe(subject, qgroup string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return sm.conn.QueueSubscribe(subject, qgroup, cb, opts...)
}

func (sm *Conn) Close() error {
	return sm.conn.Close()
}

func (sm *Conn) NatsConn() *nats.Conn {
	return sm.conn.NatsConn()
}

// Augmented functions

func (sm *Conn) SubscribeChan(subject string, msgChan chan *stan.Msg, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return sm.conn.Subscribe(subject, func(msg *stan.Msg) {
		msgChan <- msg
	}, opts...)
}
