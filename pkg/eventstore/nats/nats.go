package nats

import (
	"fmt"
	"time"

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

func (cn *Conn) Publish(subject string, data []byte) error {
	return cn.conn.Publish(subject, data)
}

func (cn *Conn) PublishAsync(subject string, data []byte, ah stan.AckHandler) (string, error) {
	return cn.conn.PublishAsync(subject, data, ah)
}

func (cn *Conn) Subscribe(subject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return cn.conn.Subscribe(subject, cb, opts...)
}

func (cn *Conn) QueueSubscribe(subject, qgroup string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return cn.conn.QueueSubscribe(subject, qgroup, cb, opts...)
}

func (cn *Conn) Close() error {
	return cn.conn.Close()
}

func (cn *Conn) NatsConn() *nats.Conn {
	return cn.conn.NatsConn()
}

// Augmented functions

func (cn *Conn) SubscribeChan(subject string, msgChan chan *stan.Msg, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return cn.conn.Subscribe(subject, func(msg *stan.Msg) {
		msgChan <- msg
	}, opts...)
}

// Python style element selector (-1 = len(events)-1)

const MOST_RECENT_MSG uint64 = 0
const FIRST_MSG uint64 = 1

// 0 == current
func (cn *Conn) Msg(subject string, seqId uint64) (*stan.Msg, error) {
	msgRange, err := cn.MsgSeqRange(subject, seqId, seqId)
	if err != nil {
		return nil, err
	}
	if len(msgRange) == 0 {
		return nil, nil
	}
	return msgRange[0], nil
}

// TODO use channel as output to enable backpressure
// 0 == 1
func (cn *Conn) MsgSeqRange(subject string, seqStart uint64, seqEnd uint64) ([]*stan.Msg, error) {
	// Find boundary if 0
	if seqEnd == 0 {
		rightBound := make(chan uint64)

		leftSub, err := cn.conn.Subscribe(subject, func(msg *stan.Msg) {
			rightBound <- msg.Sequence
			msg.Sub.Close()
			msg.Ack()
		}, stan.MaxInflight(1), stan.SetManualAckMode(), stan.StartWithLastReceived())
		if err != nil {
			return nil, err
		}
		defer leftSub.Close()
		select {
		case seqEnd = <-rightBound:
		case <-time.After(time.Duration(10) * time.Second):
			return nil, fmt.Errorf("MsgSeqRange timed out while finding boundary for subject '%s'", subject)
		}
	}

	if seqStart > seqEnd {
		return nil, fmt.Errorf("seqStart '%v' can not be larger than seqEnd '%v'.", seqStart, seqEnd)
	}

	// Subscribe until boundary
	leftBoundOptions := []stan.SubscriptionOption{stan.MaxInflight(1), stan.SetManualAckMode()}
	if seqStart == 0 {
		leftBoundOptions = append(leftBoundOptions, stan.StartWithLastReceived())
	} else {
		leftBoundOptions = append(leftBoundOptions, stan.StartAtSequence(seqStart))
	}

	result := []*stan.Msg{}
	c := make(chan *stan.Msg)
	sub, err := cn.conn.Subscribe(subject, func(msg *stan.Msg) {
		defer msg.Ack()
		// TODO maybe add a timeout here too?
		c <- msg
		if msg.Sequence == seqEnd {
			msg.Sub.Close()
			close(c)
		}
	}, leftBoundOptions...)
	if err != nil {
		return nil, err
	}
	defer sub.Close()

	for msg := range c {
		result = append(result, msg)
	}

	return result, nil
}
