package fes

import (
	"time"

	"github.com/fission/fission-workflow/pkg/util/labels"
	"github.com/fission/fission-workflow/pkg/util/labels/kubelabels"
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
	return kubelabels.New(map[string]string{
		"aggregate.id":   m.Aggregate.Id,
		"aggregate.type": m.Aggregate.Type,
	})
}
