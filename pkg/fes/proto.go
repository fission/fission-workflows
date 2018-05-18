package fes

import (
	"time"

	"github.com/fission/fission-workflows/pkg/util/labels"
	"github.com/golang/protobuf/ptypes"
)

// Amendments to Protobuf models

func (m *Event) CreatedAt() time.Time {
	t, err := ptypes.Timestamp(m.Timestamp)
	if err != nil {
		panic(err)
	}
	return t
}

func (m *Event) Labels() labels.Labels {

	parent := m.Parent
	if parent == nil {
		parent = &Aggregate{}
	}

	// TODO could be created using reflection
	// TODO cache somewhere to avoid rebuilding on every request
	return labels.Set{
		"aggregate.id":   m.Aggregate.Id,
		"aggregate.type": m.Aggregate.Type,
		"parent.type":    parent.Type,
		"parent.id":      parent.Id,
		"event.id":       m.Id,
		"event.type":     m.Type,
	}
}

func (m *Event) BelongsTo(parent Aggregator) bool {
	a := parent.Aggregate()
	return *m.Aggregate != a && *m.Parent != a
}

func (m *Aggregate) Format() string {
	return m.GetType() + "/" + m.GetId()
}
