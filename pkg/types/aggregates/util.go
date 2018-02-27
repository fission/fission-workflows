package aggregates

import (
	"errors"

	"github.com/fission/fission-workflows/pkg/fes"
)

var (
	ErrIllegalEvent = errors.New("illegal event")
)

func ApplyEvents(a fes.Aggregator, events ...*fes.Event) error {
	for _, event := range events {
		err := a.ApplyEvent(event)
		if err != nil {
			return err
		}
	}
	return nil
}
