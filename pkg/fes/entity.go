package fes

import (
	"errors"
	"reflect"
)

// BaseEntity is a helper to implement most of the Entity interface.
//
// Structs using this struct will only need to implement the following methods:
// - ApplyEvent(event)
type BaseEntity struct {
	aggregate Aggregate
	// parent is a pointer to the wrapper of this mixin, to allow for reflection-based aggregation.
	parent Entity
}

func (am *BaseEntity) Aggregate() Aggregate {
	if am == nil {
		return Aggregate{}
	}
	return am.aggregate
}

// UpdateState mutates the current Entity to the new provided Entity.
//
// By default it uses reflection to update the fields. For improved performance override this method with a
// aggregate-specific one.
func (am *BaseEntity) UpdateState(newState Entity) error {
	if newState.Aggregate() != am.Aggregate() {
		return errors.New("invalid newState")
	}

	n := reflect.Indirect(reflect.ValueOf(newState))
	old := reflect.Indirect(reflect.ValueOf(am.parent))

	for i := 0; i < old.NumField(); i++ {
		updatedField := n.Field(i)

		field := old.Field(i)
		if field.IsValid() {
			if field.CanSet() {
				field.Set(updatedField)
			}
		}
	}
	return nil
}

func (am BaseEntity) CopyBaseEntity(self Entity) *BaseEntity {
	return &BaseEntity{
		aggregate: am.aggregate,
		parent:    self,
	}
}

func NewBaseEntity(thiz Entity, aggregate Aggregate) *BaseEntity {
	return &BaseEntity{
		aggregate: aggregate,
		parent:    thiz,
	}
}
