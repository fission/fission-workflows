package nats

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/nats-io/go-nats-streaming"
	"github.com/sirupsen/logrus"
)

const (
	subjectActivity          = "_activity"
	mostRecentMsg     uint64 = 0
	firstMsg          uint64 = 1
	rangeFetchTimeout        = time.Duration(1) * time.Minute
)

type eventType int32

const (
	noop eventType = iota
	deleted
)

type subjectEvent struct {
	Subject string    `json:"Subject,omitempty"`
	Type    eventType `json:"type,omitempty"`
}

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

// Msg has a python style element selector (-1 = len(events)-1)
func (cn *Conn) Msg(subject string, seqID uint64) (*stan.Msg, error) {
	msgRange, err := cn.MsgSeqRange(subject, seqID, seqID)
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
		case <-time.After(time.Duration(5) * time.Second):
			return nil, fmt.Errorf("timed out while finding boundary for Subject '%s'", subject)
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

	var result []*stan.Msg
	elementC := make(chan *stan.Msg)
	errC := make(chan error)
	sub, err := cn.Subscribe(subject, func(msg *stan.Msg) {
		defer msg.Ack()

		select {
		case <-time.After(rangeFetchTimeout):
			errC <- errors.New("range fetch timeout")
			close(elementC)
			close(errC)
		case elementC <- msg:
			if msg.Sequence == seqEnd {
				msg.Sub.Close()
				close(elementC)
				close(errC)
			}
		}
	}, leftBoundOptions...)
	if err != nil {
		return nil, err
	}
	subsActive.WithLabelValues("counter").Inc()
	defer func() {
		sub.Close()
		subsActive.WithLabelValues("counter").Dec()
	}()

	for {
		select {
		case err := <-errC:
			return result, err
		case msg := <-elementC:
			if msg != nil {
				result = append(result, msg)
			}
		}
	}
}

// WildcardConn is an abstraction on top of Conn that provides wildcard support
type WildcardConn struct {
	*Conn
}

func NewWildcardConn(conn stan.Conn) *WildcardConn {
	return &WildcardConn{
		Conn: &Conn{conn},
	}
}

func (wc *WildcardConn) Subscribe(wildcardSubject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	if !hasWildcard(wildcardSubject) {
		return wc.Conn.Subscribe(wildcardSubject, cb, opts...)
	}

	ws := &WildcardSub{
		subject: wildcardSubject,
		sources: map[string]stan.Subscription{},
	}

	metaSub, err := wc.Conn.Subscribe(subjectActivity, func(msg *stan.Msg) {
		subjectEvent := &subjectEvent{}
		err := json.Unmarshal(msg.Data, subjectEvent)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"msg":             subjectEvent,
				"wildcardSubject": wildcardSubject,
			}).Warnf("Failed to parse subjectEvent.")
			return
		}
		subject := subjectEvent.Subject

		// Although the activity channel should be specific to one query, recheck if wildcardSubject falls in range of query.
		if !queryMatches(subject, wildcardSubject) {
			return
		}

		logrus.WithFields(logrus.Fields{
			"wildcardSubject": wildcardSubject,
			"event":           subjectEvent,
		}).Debug("NatsClient received activity.")

		// TODO add semaphore for ws.sources manipulation
		switch subjectEvent.Type {
		case noop:
			// Create a new listener if event is of a new subject
			// TODO issue: if resumed previous event will be replayed, check if these are ignored by the event store!
			if _, ok := ws.sources[subject]; !ok {
				sub, err := wc.Subscribe(subject, cb, opts...)
				if err != nil {
					logrus.Errorf("Failed to subscribe to wildcardSubject '%v': %v", subjectEvent, err)
				}
				ws.sources[subject] = sub
				subsActive.WithLabelValues(subjectEvent.Subject[:strings.Index(subjectEvent.Subject, ".")]).Inc()
			}
		case deleted:
			// Delete the current listener of the subject of the event
			if _, ok := ws.sources[subject]; ok {
				err := ws.sources[subject].Unsubscribe()
				if err != nil {
					logrus.Errorf("Failed to close (sub)listener: %v", err)
				}
				subsActive.WithLabelValues(subjectEvent.Subject[:strings.Index(subjectEvent.Subject, ".")]).Dec()
			}
		default:
			panic(fmt.Sprintf("Unknown eventType: %v", subjectEvent))
		}
	}, stan.DeliverAllAvailable())
	if err != nil {
		return nil, err
	}
	ws.activitySub = metaSub
	logrus.Infof("Subscribed to '%s'", wildcardSubject)

	return ws, nil
}

func (wc *WildcardConn) Publish(subject string, data []byte) error {
	err := wc.Conn.Publish(subject, data)
	if err != nil {
		return err
	}

	// Announce Subject activity on notification thread, because of missing wildcards in NATS streaming
	activityEvent := &subjectEvent{
		Subject: subject,
		Type:    noop, // TODO infer from context if noop or closed
	}
	err = wc.publishActivity(activityEvent)
	if err != nil {
		logrus.Warnf("Failed to publish Subject '%s': %v", subject, err)
	}

	return nil
}

func (wc *WildcardConn) publishActivity(activity *subjectEvent) error {
	subjectData, err := json.Marshal(activity)
	if err != nil {
		return err
	}

	err = wc.Conn.Publish(subjectActivity, subjectData)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"Subject": subjectActivity,
		"event":   activity,
	}).Debug("Published activity event to event store.")

	return nil
}

func (wc *WildcardConn) List(matcher fes.StringMatcher) ([]string, error) {

	msgs, err := wc.Conn.MsgSeqRange(subjectActivity, firstMsg, mostRecentMsg)
	if err != nil {
		return nil, err
	}
	subjectCount := map[string]int{}
	for _, msg := range msgs {
		subjectEvent := &subjectEvent{}
		err := json.Unmarshal(msg.Data, subjectEvent)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"msg":             subjectEvent,
				"activitySubject": subjectActivity,
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

	var results []string
	for subject := range subjectCount {
		results = append(results, subject)
	}

	return results, nil
}

// WildcardSub is an abstraction on top of stan.Subscription that provides wildcard support
type WildcardSub struct {
	subject     string
	sources     map[string]stan.Subscription
	activitySub stan.Subscription
}

func (ws *WildcardSub) Unsubscribe() error {
	logrus.Infof("Unsubscribing wildcard subscription for '%v'", ws.subject)
	err := ws.activitySub.Unsubscribe()
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
