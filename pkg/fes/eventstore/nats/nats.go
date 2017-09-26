package nats

import (
	"fmt"
	"time"

	"strings"

	"encoding/json"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/nats-io/go-nats-streaming"
	"github.com/sirupsen/logrus"
)

// Conn is a wrapper of 'stan.Conn' struct to augment the API with bounded subscriptions and channel-based subscriptions
type Conn struct {
	stan.Conn
}

func NewConn(conn stan.Conn) *Conn {
	return &Conn{conn}
}

func (cn *Conn) SubscribeChan(subject string, msgChan chan *stan.Msg, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return cn.Subscribe(subject, func(msg *stan.Msg) {
		msgChan <- msg
	}, opts...)
}

const (
	SUBJECT_ACTIVITY = "_activity"
)

const MOST_RECENT_MSG uint64 = 0
const FIRST_MSG uint64 = 1

// Msg has a python style element selector (-1 = len(events)-1)
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

func (cn *Conn) MsgSeqRange(subject string, seqStart uint64, seqEnd uint64) ([]*stan.Msg, error) {
	// Find boundary if 0
	if seqEnd == 0 {
		rightBound := make(chan uint64)

		leftSub, err := cn.Subscribe(subject, func(msg *stan.Msg) {
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
			return nil, fmt.Errorf("timed out while finding boundary for subject '%s'", subject)
		}
	}

	if seqStart > seqEnd {
		return nil, fmt.Errorf("seqStart '%v' can not be larger than seqEnd '%v'", seqStart, seqEnd)
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
	sub, err := cn.Subscribe(subject, func(msg *stan.Msg) {
		defer msg.Ack()
		// TODO add a timeout here
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

// WildcardConn is an abstraction on top of Conn that provides wildcard support
type WildcardConn struct {
	*Conn
	activitySub stan.Subscription
}

func NewWildcardConn(conn stan.Conn) *WildcardConn {
	return &WildcardConn{
		Conn: &Conn{conn},
	}
}

func (wc *WildcardConn) Subscribe(subject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	if !hasWildcard(subject) {
		return wc.Conn.Subscribe(subject, cb, opts...)
	}

	ws := &WildcardSub{
		sources: map[string]stan.Subscription{},
	}
	if wc.activitySub != nil {
		wc.activitySub.Close()
	}
	metaSub, err := wc.Conn.Subscribe(SUBJECT_ACTIVITY, func(msg *stan.Msg) {
		subjectEvent := &SubjectEvent{}
		err := json.Unmarshal(msg.Data, subjectEvent)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"msg":     subjectEvent,
				"subject": subject,
			}).Warnf("Failed to parse subjectEvent.")
			return
		}
		logrus.WithFields(logrus.Fields{
			"subject": subject,
			"event":   subjectEvent,
		}).Debug("NatsClient received activity.")

		// Although the activity channel should be specific to one query, recheck if subject falls in range of query.
		if !queryMatches(subject, subjectEvent.Subject) {
			return
		}

		switch subjectEvent.Type {
		case ACTIVITY_CREATED:
			if _, ok := ws.sources[subject]; !ok {

				sub, err := wc.Subscribe(subject, cb, opts...)
				if err != nil {
					logrus.Errorf("Failed to subscribe to subject '%v': %v", subjectEvent, err)
				}
				ws.sources[subject] = sub
			}
		default:
			// TODO notify subscription that subject has been closed and close channel
			panic(fmt.Sprintf("Unknown ActivityEvent: %v", subjectEvent))
		}
	}, stan.DeliverAllAvailable())
	if err != nil {
		return nil, err
	}
	wc.activitySub = metaSub

	return ws, nil
}

func (wc *WildcardConn) Publish(subject string, data []byte) error {
	err := wc.Conn.Publish(subject, data)
	if err != nil {
		return err
	}

	// Announce subject activity on notification thread, because of missing wildcards in NATS streaming
	activityEvent := &SubjectEvent{
		Subject: subject,
		Type:    ACTIVITY_CREATED, // TODO infer from context if created or closed
	}
	err = wc.publishActivity(activityEvent)
	if err != nil {
		logrus.Warnf("Failed to publish subject '%s': %v", subject, err)
	}

	return nil
}

func (wc *WildcardConn) publishActivity(activity *SubjectEvent) error {
	subjectData, err := json.Marshal(activity)
	if err != nil {
		return err
	}

	err = wc.Conn.Publish(SUBJECT_ACTIVITY, subjectData)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"subject": SUBJECT_ACTIVITY,
		"event":   activity,
	}).Debug("Published activity event to event store.")

	return nil
}

func (wc *WildcardConn) List(matcher fes.StringMatcher) ([]string, error) {

	msgs, err := wc.Conn.MsgSeqRange(SUBJECT_ACTIVITY, FIRST_MSG, MOST_RECENT_MSG)
	if err != nil {
		return nil, err
	}
	subjectCount := map[string]int{}
	for _, msg := range msgs {
		subjectEvent := &SubjectEvent{}
		err := json.Unmarshal(msg.Data, subjectEvent)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"msg":             subjectEvent,
				"activitySubject": SUBJECT_ACTIVITY,
			}).Warnf("Failed to parse subjectEvent.")
			continue
		}

		subject := subjectEvent.Subject
		if matcher.Match(subject) {
			count := 1
			if c, ok := subjectCount[subject]; ok {
				count += c
			}
			subjectCount[subject] = count
		}
	}

	results := []string{}
	for subject := range subjectCount {
		results = append(results, subject)
	}

	return results, nil
}

type WildcardSub struct {
	sources map[string]stan.Subscription
}

func (ws *WildcardSub) Unsubscribe() error {
	var err error
	for id, source := range ws.sources {
		err = source.Unsubscribe()
		delete(ws.sources, id)
	}
	return err
}

func (ws *WildcardSub) Close() error {
	return ws.Unsubscribe()
}

func hasWildcard(subject string) bool {
	return strings.ContainsAny(subject, "*>")
}

func queryMatches(subject string, query string) bool {
	subjectParts := strings.Split(subject, ".")
	queryParts := strings.Split(query, ".")

	for key, part := range subjectParts {
		if part == "" {
			return false
		}

		if len(query) < key {
			return false
		}

		if queryParts[key] == ">" {
			return true
		}

		if queryParts[key] == "*" {
			continue
		}

		if queryParts[key] != part {
			return false
		}
	}
	return true
}

type ActivityEvent int32

const (
	ACTIVITY_CREATED ActivityEvent = iota
	ACTIVITY_DELETED
)

type SubjectEvent struct {
	Subject string        `json:"subject,omitempty"`
	Type    ActivityEvent `json:"type,omitempty"`
}
