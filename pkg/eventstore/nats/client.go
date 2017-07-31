package nats

import (
	"strings"

	"fmt"

	"github.com/fission/fission-workflow/pkg/eventstore"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats-streaming"
	"github.com/sirupsen/logrus"
)

// TODO use logrus with fields
type Client struct {
	conn          *Conn
	subscriptions []*Subscription
}

const (
	SUBJECT_ACTIVITY = "_activity"
)

func New(conn *Conn) eventstore.Client {
	return &Client{conn, []*Subscription{}}
}

// Wildcards are allowed in the subject
// TODO address other config fields other than subject and eventCh
func (nc *Client) Subscribe(config *eventstore.SubscriptionConfig) (eventstore.Subscription, error) {
	query := config.Subject
	if len(query) == 0 {
		return nil, fmt.Errorf("Subscribe requires a subject.")
	}

	if config.EventCh == nil {
		return nil, fmt.Errorf("Subscribe requires a event channel to emit the events to.")
	}

	sub := &Subscription{
		config:  config,
		sources: map[string]stan.Subscription{},
	}

	if !hasWildcard(query) {
		stanSub, err := nc.subscribeSingle(&eventstore.SubscriptionConfig{
			Subject: query,
			EventCh: config.EventCh,
			// TODO address other config params as well
		})
		if err != nil {
			return nil, err
		}
		sub.sources[query] = stanSub
		return sub, nil
	}

	activitySubject := toActivitySubject(query)

	metaSub, err := nc.conn.Subscribe(activitySubject, func(msg *stan.Msg) {
		subjectEvent := &eventstore.SubjectEvent{}
		err := proto.Unmarshal(msg.Data, subjectEvent)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"msg":             subjectEvent,
				"query":           query,
				"activitySubject": activitySubject,
			}).Warnf("Failed to parse subjectEvent.")
			return
		}
		logrus.WithFields(logrus.Fields{
			"subject": activitySubject,
			"event":   subjectEvent,
		}).Debug("NatsClient received activity.")

		// Although the activity channel should be specific to one query, recheck if subject falls in range of query.
		if !queryMatches(subjectEvent.GetSubject(), query) {
			logrus.WithFields(logrus.Fields{
				"activitySubject":  activitySubject,
				"subscribeSubject": query,
				"subjectEvent":     subjectEvent,
			}).Debug("Ignoring activity event, because it does not match subscription subject.")
			return
		}

		switch subjectEvent.GetType() {
		case eventstore.SubjectEvent_CREATED:
			if _, ok := sub.sources[subjectEvent.GetSubject()]; !ok {
				stanSub, err := nc.subscribeSingle(&eventstore.SubscriptionConfig{
					Subject: subjectEvent.GetSubject(),
					EventCh: config.EventCh,
					// TODO address other config params as well
				})
				if err != nil {
					logrus.Errorf("Failed to subscribe to subject '%v': %v", subjectEvent, err)
				}
				sub.sources[subjectEvent.GetSubject()] = stanSub
			}
		default:
			// TODO notify subscription that subject has been closed, close channel, whatever
			panic(fmt.Sprintf("Unknown SubjectEvent: %v", subjectEvent))
		}
	}, stan.DeliverAllAvailable())
	if err != nil {
		panic(err)
		return nil, err
	}

	sub.sources[activitySubject] = metaSub
	nc.subscriptions = append(nc.subscriptions, sub)

	logrus.WithFields(logrus.Fields{
		"config":  config,
		"subject": activitySubject,
	}).Debug("Subscribed to activity subject")

	return sub, nil
}

func (nc *Client) subscribeSingle(config *eventstore.SubscriptionConfig) (stan.Subscription, error) {
	if hasWildcard(config.Subject) {
		logrus.Panicf("subscribeSingle does not support wildcards in subject '%s'", config.Subject)
	}

	logrus.WithFields(logrus.Fields{
		"config":  config,
		"subject": config.Subject,
	}).Debug("Subscribed to subject.")

	return nc.conn.Subscribe(config.Subject, func(msg *stan.Msg) {
		event, err := unmarshalMsg(msg)
		if err != nil {
			logrus.Errorf("Failed to retrieve event from msg '%v'", msg)
		}
		config.EventCh <- event
	}, stan.DeliverAllAvailable())
}

func (nc *Client) Get(subject string) ([]*eventstore.Event, error) {
	if hasWildcard(subject) {
		logrus.Panicf("subscribeSingle does not support wildcards in subject '%s'", subject)
	}

	logrus.WithField("subject", subject).Debug("GET events from event store")

	msgs, err := nc.conn.MsgSeqRange(subject, FIRST_MSG, MOST_RECENT_MSG)
	if err != nil {
		return nil, err
	}
	results := []*eventstore.Event{}
	for _, msg := range msgs {
		event, err := unmarshalMsg(msg)
		if err != nil {
			return nil, err
		}
		results = append(results, event)
	}

	return results, nil
}

func (nc *Client) Subjects(query string) ([]string, error) {

	logrus.WithField("query", query).Debug("LIST subjects from event store")

	activitySubject := toActivitySubject(query)

	msgs, err := nc.conn.MsgSeqRange(activitySubject, FIRST_MSG, MOST_RECENT_MSG)
	if err != nil {
		return nil, err
	}
	subjectCount := map[string]int{}
	for _, msg := range msgs {
		subjectEvent := &eventstore.SubjectEvent{}
		err := proto.Unmarshal(msg.Data, subjectEvent)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"msg":             subjectEvent,
				"query":           query,
				"activitySubject": activitySubject,
			}).Warnf("Failed to parse subjectEvent.")
			continue
		}

		if queryMatches(subjectEvent.GetSubject(), query) {
			count := 1
			if c, ok := subjectCount[subjectEvent.GetSubject()]; ok {
				count += c
			}
			subjectCount[subjectEvent.GetSubject()] = count
		}
	}

	results := []string{}
	for subject := range subjectCount {
		results = append(results, subject)
	}

	return results, nil
}

func (nc *Client) Append(event *eventstore.Event) error {
	invokeSubject := createSubject(event.GetEventId())
	data, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	err = nc.conn.Publish(invokeSubject, data)
	if err != nil {
		return err
	}

	// Announce subject activity on notification thread, because of missing wildcards in NATS streaming
	activitySubject := toActivitySubject(invokeSubject)
	activityEvent := &eventstore.SubjectEvent{
		Subject: invokeSubject,
		Type:    eventstore.SubjectEvent_CREATED, // TODO infer from context if created or closed
	}
	err = nc.publishActivity(activityEvent)
	if err != nil {
		logrus.Warnf("Failed to publish subject '%s' to activity subject '%s': %v", invokeSubject,
			activitySubject, err)
	}
	logrus.WithFields(logrus.Fields{
		"subject":         invokeSubject,
		"event":           event,
		"activitySubject": activitySubject,
		"activityEvent":   activityEvent,
	}).Info("PUBLISH event to event store.")

	return nil
}

func (nc *Client) publishActivity(activity *eventstore.SubjectEvent) error {
	subjectData, err := proto.Marshal(activity)
	if err != nil {
		return err
	}

	subject := toActivitySubject(activity.GetSubject())
	err = nc.conn.Publish(subject, subjectData)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"subject": subject,
		"event":   activity,
	}).Debug("Published activity event to event store.")

	return nil
}

// Replay re-emits events for subject for subscription
func (nc *Client) Replay(subject string, context eventstore.Subscription) error {
	sub, ok := context.(*Subscription)
	if !ok {
		return fmt.Errorf("Subscription '%v' has to be of type Subscription to be replayed", context)
	}

	source, ok := sub.sources[subject]
	if !ok {
		return fmt.Errorf("Subject '%s' does not belong to the current subscription '%v'", subject, context)
	}

	if err := source.Close(); err != nil {
		return fmt.Errorf("Failed to close existing source '%s', due to: %v", subject, err)
	}

	stanSub, err := nc.subscribeSingle(&eventstore.SubscriptionConfig{
		Subject: subject,
		EventCh: context.Config().EventCh,
	})
	if err != nil {
		return err
	}

	sub.sources[subject] = stanSub

	return nil
}

func (nc *Client) Close() error {
	for _, sub := range nc.subscriptions {
		if err := sub.Close(); err != nil {
			return fmt.Errorf("Failed to close subscription for '%s' due to: %v", sub, err)
		}
	}

	return nc.conn.Close()
}

/*

	Subscription

*/
// TODO make concurrency-proof
type Subscription struct {
	config  *eventstore.SubscriptionConfig
	sources map[string]stan.Subscription
}

func (ns *Subscription) Close() error {
	for subject, sub := range ns.sources {
		if err := sub.Close(); err != nil {
			return fmt.Errorf("Failed to close subscription for '%s' due to: %v", subject, err)
		}
	}
	close(ns.config.EventCh)
	return nil
}

func (ns *Subscription) Config() *eventstore.SubscriptionConfig {
	return ns.config
}

/*

	Utility functions

*/
func toActivitySubject(subject string) string {
	// For now, just broadcast on a common subject
	return SUBJECT_ACTIVITY
}

func unmarshalMsg(msg *stan.Msg) (*eventstore.Event, error) {
	e := &eventstore.Event{}
	err := proto.Unmarshal(msg.Data, e)
	if err != nil {
		return nil, err
	}

	e.EventId.Id = fmt.Sprintf("%s", msg.Sequence)
	return e, nil
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

func createSubject(event *eventstore.EventID) string {
	return strings.Join(event.GetSubjects(), ".")
}
