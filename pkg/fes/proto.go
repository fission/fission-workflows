package fes

import (
	"time"

	"github.com/fission/fission-workflow/pkg/util/labels"
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

	// TODO make more efficient
	return labels.SetLabels{
		"aggregate.id":   m.Aggregate.Id,
		"aggregate.type": m.Aggregate.Type,
		"parent.type":    parent.Type,
		"parent.id":      parent.Id,
		"event.id":       m.Id,
		"event.type":     m.Type,
	}
}
