package fes

import "errors"

func ValidateAggregate(aggregate *Aggregate) error {
	if aggregate == nil {
		return ErrInvalidAggregate.WithError(errors.New("aggregate is nil"))
	}

	if len(aggregate.Id) == 0 {
		return ErrInvalidAggregate.WithAggregate(aggregate).WithError(errors.New("aggregate does not contain an id"))
	}

	if len(aggregate.Type) == 0 {
		return ErrInvalidAggregate.WithAggregate(aggregate).WithError(errors.New("aggregate does not contain a type"))
	}

	return nil
}

// ValidateEvent validates the event struct.
//
// Note:
// - It does not parse or check the event data, except that it checks that it is not nil
// - It does not check the event ID, since events that have not been persisted do not have an ID assigned yet.
func ValidateEvent(event *Event) error {
	if event == nil {
		return errors.New("event is nil")
	}
	if len(event.Type) == 0 {
		return errors.New("event has no event type")
	}
	if err := ValidateAggregate(event.Aggregate); err != nil {
		return err
	}
	if event.Parent != nil {
		if err := ValidateAggregate(event.Parent); err != nil {
			return err
		}
	}
	if event.Timestamp == nil {
		return errors.New("event has no timestamp")
	}
	if event.Data == nil {
		return errors.New("event has no data")
	}
	return nil
}

func ValidateEntity(entity Entity) error {
	if entity == nil {
		return errors.New("entity is nil")
	}
	key := entity.Aggregate()
	return ValidateAggregate(&key)
}
