package backoff

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestBackoff_DefaultBackoffAlgorithm(t *testing.T) {
	b := Backoff()

	assert.Equal(t, DefaultBackoffAlgorithm, b.Algorithm)
	assert.Equal(t, 1, b.Attempts)

	b2 := Backoff(b)

	assert.Equal(t, DefaultBackoffAlgorithm, b2.Algorithm)
	assert.Equal(t, 2, b2.Attempts)
	assert.True(t, b2.LockedUntil.After(b.LockedUntil))
	assert.True(t, b2.Lockout.Nanoseconds() > b.Lockout.Nanoseconds())
}