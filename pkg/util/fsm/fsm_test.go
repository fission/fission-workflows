package fsm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFsm_NodeExists(t *testing.T) {
	machine := New(1, []Transition{
		{"a", 1, 2},
	})

	assert.Equal(t, machine.NodeExists(1), true)
	assert.Equal(t, machine.NodeExists(2), true)
	assert.Equal(t, machine.NodeExists(3), false)
	assert.Equal(t, machine.NodeExists("a"), false)
}

func TestFsm_FindTransition(t *testing.T) {
	machine := New(1, []Transition{
		{"a", 1, 2},
	})

	assert.Equal(t, machine.GetTransition(1, 2).Event, "a")
	assert.Empty(t, machine.GetTransition(1, 4))
	assert.Empty(t, machine.GetTransition(4, 1))
	assert.Empty(t, machine.GetTransition(3, 3))
}

func TestFsm_NewInstance(t *testing.T) {
	machine := New(1, []Transition{
		{"a", 1, 2},
	})

	i1 := machine.NewInstance()
	assert.Equal(t, i1.Current(), 1)

	i2 := machine.NewInstance("b")
	assert.Equal(t, i2.Current(), "b")
}

func TestFsmInstance_Evaluate(t *testing.T) {
	machine := New(1, []Transition{
		{"a", 1, 2},
	})
	i := machine.NewInstance()

	i.Evaluate("unknown")
	err := i.Evaluate("a")

	assert.NoError(t, err)
	assert.Equal(t, i.Current(), 2)
}

func TestFsmInstance_CanTransitionTo(t *testing.T) {
	machine := New(1, []Transition{
		{"a", 1, 2},
	})
	i := machine.NewInstance()

	assert.Equal(t, i.CanTransitionTo(2), true)
	assert.Equal(t, i.CanTransitionTo(3), false)
	assert.Equal(t, i.CanTransitionTo("a"), false)
}
