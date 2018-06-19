package nats

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/nats-io/go-nats"
	"github.com/nats-io/go-nats-streaming"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/sirupsen/logrus"
)

const (
	defaultClient = "fes"
)

var (
	ErrInvalidAggregate = errors.New("invalid aggregate")

	subsActive = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: "fes",
		Subsystem: "nats",
		Name:      "subs_active",
		Help:      "Number of active subscriptions to NATS subjects.",
	}, []string{"subType"})

	eventsAppended = prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "fes",
		Subsystem: "nats",
		Name:      "events_appended_total",
		Help:      "Count of appended events (including any internal events).",
	}, []string{"eventType"})

	eventDelay = prometheus.NewSummary(prometheus.SummaryOpts{
		Namespace: "fes",
		Subsystem: "nats",
		Name:      "event_propagation_delay",
		Help:      "Delay between event publish and receive by the subscribers.",
	})
)

func init() {
	prometheus.MustRegister(subsActive, eventsAppended, eventDelay)
}

// EventStore is a NATS-based implementation of the EventStore interface.
type EventStore struct {
	pubsub.Publisher
	conn   *WildcardConn
	sub    map[fes.Aggregate]stan.Subscription
	Config Config
}

type Config struct {
	Cluster string
	Client  string
	URL     string // e.g. nats://localhost:9300
}

func NewEventStore(conn *WildcardConn, cfg Config) *EventStore {
	return &EventStore{
		Publisher: pubsub.NewPublisher(),
		conn:      conn,
		sub:       map[fes.Aggregate]stan.Subscription{},
		Config:    cfg,
	}
}

// Connect to a NATS cluster using the config.
func Connect(cfg Config) (*EventStore, error) {
	if cfg.Client == "" {
		cfg.Client = defaultClient
	}
	if cfg.URL == "" {
		cfg.URL = nats.DefaultURL
	}
	url := stan.NatsURL(cfg.URL)
	conn, err := stan.Connect(cfg.Cluster, cfg.Client, url)
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

// Watch a aggregate type for new events. The events are emitted over the publisher interface.
func (es *EventStore) Watch(aggregate fes.Aggregate) error {
	subject := fmt.Sprintf("%s.>", aggregate.Type)
	sub, err := es.conn.Subscribe(subject, func(msg *stan.Msg) {
		event, err := toEvent(msg)
		if err != nil {
			logrus.Error(err)
			return
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
			return
		}

		// Record the time it took for the event to be propagated from publisher to subscriber.
		ts, _ := ptypes.Timestamp(event.Timestamp)
		eventDelay.Observe(float64(time.Now().Sub(ts).Nanoseconds()))

	}, stan.DeliverAllAvailable())
	if err != nil {
		return err
	}

	logrus.Infof("Backend client watches:' %s'", subject)
	es.sub[aggregate] = sub
	return nil
}

func (es *EventStore) Close() error {
	err := es.conn.Close()
	if err != nil {
		return err
	}
	return nil
}

// Append publishes (and persists) an event on the NATS message queue
func (es *EventStore) Append(event *fes.Event) error {
	// TODO make generic / configurable whether to fold event into parent's Subject
	subject := toSubject(*event.Aggregate)
	if event.Parent != nil {
		subject = toSubject(*event.Parent)
	}
	data, err := proto.Marshal(event)
	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"aggregate":    event.Aggregate.Format(),
		"parent":       event.Parent.Format(),
		"nats.subject": subject,
	}).Infof("Event added: %v", event.Type)

	err = es.conn.Publish(subject, data)
	if err != nil {
		return err
	}
	eventsAppended.WithLabelValues(event.Type).Inc()
	eventsAppended.WithLabelValues("control").Inc()
	return nil
}

// Get returns all events related to a specific aggregate
func (es *EventStore) Get(aggregate fes.Aggregate) ([]*fes.Event, error) {
	if !fes.ValidateAggregate(&aggregate) {
		return nil, ErrInvalidAggregate
	}
	subject := toSubject(aggregate)

	// TODO check if subject exists in NATS (MsgSeqRange takes a long time otherwise)

	msgs, err := es.conn.MsgSeqRange(subject, firstMsg, mostRecentMsg)
	if err != nil {
		return nil, err
	}
	var results []*fes.Event
	for k := range msgs {
		event, err := toEvent(msgs[k])
		if err != nil {
			return nil, err
		}
		results = append(results, event)
	}

	return results, nil
}

// List returns all entities of which the subject matches the StringMatcher
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

func toSubject(a fes.Aggregate) string {
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
