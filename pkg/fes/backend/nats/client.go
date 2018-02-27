package nats

import (
	"fmt"
	"strings"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/golang/protobuf/proto"
	nats "github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/sirupsen/logrus"
)

const (
	defaultClient = "fes"
)

type EventStore struct {
	pubsub.Publisher
	conn   *WildcardConn
	sub    map[fes.Aggregate]stan.Subscription
	Config Config
}

type Config struct {
	//Cluster: clusterId,
	//Client:  "someClient",
	//Url:     fmt.Sprintf("nats://%s:%d", address, port),
	Cluster string
	Client  string

	// Example: nats://localhost:9300
	Url string
}

func NewEventStore(conn *WildcardConn, cfg Config) *EventStore {
	return &EventStore{
		Publisher: pubsub.NewPublisher(),
		conn:      conn,
		sub:       map[fes.Aggregate]stan.Subscription{},
		Config:    cfg,
	}
}

func Connect(cfg Config) (*EventStore, error) {
	if cfg.Client == "" {
		cfg.Client = defaultClient
	}
	if cfg.Url == "" {
		cfg.Url = nats.DefaultURL
	}
	natsUrl := stan.NatsURL(cfg.Url)
	conn, err := stan.Connect(cfg.Cluster, cfg.Client, natsUrl)
	if err != nil {
		return nil, err
	}
	wconn := NewWildcardConn(conn)
	logrus.WithField("cluster", cfg.Cluster).
		WithField("url", "!redacted!").
		WithField("client", cfg.Client).
		Info("connected to NATS")

	return NewEventStore(wconn, cfg), nil
}

// Watch a aggregate type
func (es *EventStore) Watch(aggregate fes.Aggregate) error {
	subject := fmt.Sprintf("%s.>", aggregate.Type)
	sub, err := es.conn.Subscribe(subject, func(msg *stan.Msg) {
		event, err := toEvent(msg)
		if err != nil {
			logrus.Error(err)
		}

		logrus.WithFields(logrus.Fields{
			"aggregate.type": event.Aggregate.Type,
			"aggregate.id":   event.Aggregate.Id,
			"event.type":     event.Type,
			"event.id":       event.Id,
			"nats.Subject":   msg.Subject,
		}).Debug("Publishing aggregate event to subscribers.")

		err = es.Publisher.Publish(event)
		if err != nil {
			logrus.Error(err)
		}
	}, stan.DeliverAllAvailable())
	if err != nil {
		return err
	}

	logrus.Infof("Backend client watches:' %s'", subject)
	es.sub[aggregate] = sub
	return nil
}

func (es *EventStore) Close() error {
	return es.conn.Close()
}

func (es *EventStore) Append(event *fes.Event) error {
	// TODO make generic / configurable whether to fold event into parent's Subject
	subject := toSubject(event.Aggregate)
	if event.Parent != nil {
		subject = toSubject(event.Parent)
	}
	data, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"event.id":       event.Id,
		"event.type":     event.Type,
		"aggregate.id":   event.Aggregate.Id,
		"aggregate.type": event.Aggregate.Type,
		"nats.Subject":   subject,
	}).Info("Backend client appending event.")

	return es.conn.Publish(subject, data)
}

func (es *EventStore) Get(aggregate *fes.Aggregate) ([]*fes.Event, error) {
	subject := toSubject(aggregate)

	msgs, err := es.conn.MsgSeqRange(subject, firstMsg, mostRecentMsg)
	if err != nil {
		return nil, err
	}
	var results []*fes.Event
	for _, msg := range msgs {
		event, err := toEvent(msg)
		if err != nil {
			return nil, err
		}
		results = append(results, event)
	}

	return results, nil
}

func (es *EventStore) List(matcher fes.StringMatcher) ([]fes.Aggregate, error) {
	subjects, err := es.conn.List(matcher)
	if err != nil {
		return nil, err
	}
	var results []fes.Aggregate
	for _, subject := range subjects {
		a := toAggregate(subject)
		results = append(results, *a)
	}

	return results, nil
}

func toAggregate(subject string) *fes.Aggregate {
	parts := strings.SplitN(subject, ".", 2)
	if len(parts) < 2 {
		return nil
	}
	return &fes.Aggregate{
		Type: parts[0],
		Id:   parts[1],
	}
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

	e.Id = fmt.Sprintf("%d", msg.Sequence)
	return e, nil
}
