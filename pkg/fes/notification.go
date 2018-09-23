package fes

import (
	"github.com/fission/fission-workflows/pkg/util/pubsub"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
)

const (
	PubSubLabelEventID        = "event.id"
	PubSubLabelEventType      = "event.type"
	PubSubLabelAggregateType  = "aggregate.type"
	PubSubLabelAggregateID    = "aggregate.id"
	DefaultNotificationBuffer = 64
)

type Notification struct {
	*pubsub.EmptyMsg
	Payload   Entity
	EventType string
	SpanCtx   opentracing.SpanContext
}

func NewNotification(entity Entity, event *Event) *Notification {
	var spanCtx opentracing.SpanContext
	if event.Metadata != nil {
		sctx, err := opentracing.GlobalTracer().Extract(opentracing.TextMap, opentracing.TextMapCarrier(event.Metadata))
		if err != nil && err != opentracing.ErrSpanContextNotFound {
			logrus.Warnf("failed to extract opentracing tracer from event: %v", err)
		}
		spanCtx = sctx
	}
	return &Notification{
		EmptyMsg:  pubsub.NewEmptyMsg(event.Labels(), event.CreatedAt()),
		Payload:   entity,
		EventType: event.Type,
		SpanCtx:   spanCtx,
	}
}
