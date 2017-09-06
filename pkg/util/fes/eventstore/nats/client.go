package nats

import (
	"fmt"

	"github.com/fission/fission-workflow/pkg/util/fes"
	"github.com/fission/fission-workflow/pkg/util/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/nats-io/go-nats-streaming"
	"github.com/sirupsen/logrus"
)

type EventStore struct {
	pubsub.Publisher
	conn *WildcardConn
	sub  map[fes.Aggregate]stan.Subscription
}

func NewEventStore(conn *WildcardConn) *EventStore {
	return &EventStore{
		Publisher: pubsub.NewPublisher(),
		conn:      conn,
	}
}

// Watch a aggregate type
func (es *EventStore) Watch(aggregate fes.Aggregate) error {
	sub, err := es.conn.Subscribe(aggregate.Type, func(msg *stan.Msg) {
		event, err := toEvent(msg)
		if err != nil {
			logrus.Error(err)
		}
		err = es.Publisher.Publish(event)
		if err != nil {
			logrus.Error(err)
		}
	}, stan.DeliverAllAvailable())
	if err != nil {
		return err
	}
	es.sub[aggregate] = sub
	return nil
}

func (es *EventStore) Close() error {
	return es.conn.Close()
}

func (es *EventStore) HandleEvent(event *fes.Event) error {
	subject := toSubject(event.Aggregate)
	data, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	return es.conn.Publish(subject, data)
}

func (es *EventStore) Get(aggregate *fes.Aggregate) ([]*fes.Event, error) {
	//logrus.WithField("subject", aggregateType).Debug("GET events from event store")
	subject := toSubject(aggregate)

	msgs, err := es.conn.MsgSeqRange(subject, FIRST_MSG, MOST_RECENT_MSG)
	if err != nil {
		return nil, err
	}
	results := []*fes.Event{}
	for _, msg := range msgs {
		event, err := toEvent(msg)
		if err != nil {
			return nil, err
		}
		results = append(results, event)
	}

	return results, nil
}

func (es *EventStore) List(matcher fes.StringMatcher) ([]string, error) {
	return es.conn.List(matcher)
}

func toSubject(a *fes.Aggregate) string {
	return fmt.Sprintf("%s.%s", a.Type, a.Id)
}

func toEvent(msg *stan.Msg) (*fes.Event, error) {
	e := &fes.Event{}
	err := proto.Unmarshal(msg.Data, e)
	if err != nil {
		return nil, err
	}

	e.Id = fmt.Sprintf("%s", msg.Sequence)
	return e, nil
}
