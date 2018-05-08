package typedvalues

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrioritizeInputs(t *testing.T) {
	invalidPrioInput := MustParse("bar")
	invalidPrioInput.SetLabel("priority", "NaN")
	regularInput := MustParse("foo")
	regularInput.SetLabel("priority", "1")
	prioInput := MustParse("zzz")
	prioInput.SetLabel("priority", "2")
	highPrioInput := MustParse("important")
	highPrioInput.SetLabel("priority", "3")
	inputs := PrioritizeInputs(map[string]*TypedValue{
		"a": invalidPrioInput,
		"b": regularInput,
		"c": prioInput,
		"d": highPrioInput,
	})
	assert.Equal(t, highPrioInput, inputs[0].Val)
	assert.Equal(t, prioInput, inputs[1].Val)
	assert.Equal(t, regularInput, inputs[2].Val)
	assert.Equal(t, invalidPrioInput, inputs[3].Val)
}
