package backoff

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInstance_C(t *testing.T) {
	i := Instance{
		MaxRetries:        5,
		BackoffPolicy:     ExponentialBackoff,
		BaseRetryDuration: 10 * time.Millisecond,
	}
	start := time.Now()
	var expectedAttempt int
	for attempt := range i.C(context.TODO()) {
		assert.Equal(t, expectedAttempt, attempt)
		expectedAttempt++
	}
	assert.Equal(t, i.MaxRetries, expectedAttempt)
	log.Println("Total time backed off:", time.Now().Sub(start))
	end := time.Now()
	assert.True(t, end.Sub(start) > 200*time.Millisecond)
	assert.True(t, end.Sub(start) < 400*time.Millisecond)
}
