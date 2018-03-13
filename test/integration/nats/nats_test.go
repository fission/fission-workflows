package nats

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fes/backend/nats"
	"github.com/fission/fission-workflows/test/integration"
	"github.com/stretchr/testify/assert"
)

var (
	backend fes.Backend
)

// Tests the event store implementation with a live NATS cluster.
// This test will start and stop a NATS streaming cluster by itself.
//
// Prerequisites:
// - Docker

func TestMain(m *testing.M) {
	if testing.Short() {
		fmt.Println("Skipping NATS integration tests...")
		return
	}

	ctx := context.TODO()

	cfg := integration.RunNatsStreaming(ctx)

	natsBackend, err := nats.Connect(cfg)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to cluster: %v", err))
	}
	backend = natsBackend

	status := m.Run()
	os.Exit(status)
}

func TestNatsBackend_GetNonExistent(t *testing.T) {
	key := fes.NewAggregate("nonExistentType", "nonExistentId")

	// check
	events, err := backend.Get(key)
	assert.Error(t, err)
	assert.Empty(t, events)
}

func TestNatsBackend_Append(t *testing.T) {
	key := fes.NewAggregate("someType", "someId")
	event := fes.NewEvent(key, nil)
	err := backend.Append(event)
	assert.NoError(t, err)

	// check
	events, err := backend.Get(key)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	event.Id = "1"
	assert.Equal(t, event, events[0])
}

func TestNatsBackend_List(t *testing.T) {
	subjects, err := backend.List(&fes.ContainsMatcher{})
	assert.NoError(t, err)
	assert.NotEmpty(t, subjects)
}
