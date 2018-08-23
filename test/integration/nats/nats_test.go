package nats

import (
	"fmt"
	"os"
	"testing"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fes/backend/nats"
	fesnats "github.com/fission/fission-workflows/pkg/fes/backend/nats"
	"github.com/fission/fission-workflows/pkg/util"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"gopkg.in/ory-am/dockertest.v3"
)

var (
	backend fes.Backend
)

// Tests the event store implementation with a live NATS cluster.
// This test will start and stop a NATS streaming cluster by itself.

func TestMain(m *testing.M) {
	if testing.Short() {
		log.Info("Short test; skipping NATS integration tests.")
		return
	}
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	id := util.UID()
	clusterId := fmt.Sprintf("fission-workflows-tests-%s", id)
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{

		Repository:   "nats-streaming",
		Tag:          "0.8.0-beta",
		Cmd:          []string{"-cid", clusterId, "-p", fmt.Sprintf("%d", 4222)},
		ExposedPorts: []string{"4222"},
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	cleanup := func() {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}
	defer cleanup()

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	if err := pool.Retry(func() error {
		cfg := fesnats.Config{
			Cluster: clusterId,
			Client:  fmt.Sprintf("client-%s", id),
			URL:     fmt.Sprintf("nats://%s:%s", "0.0.0.0", resource.GetPort("4222/tcp")),
		}

		var err error
		backend, err = nats.Connect(cfg)
		if err != nil {
			return fmt.Errorf("failed to connect to cluster: %v", err)
		}
		return nil
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	log.Info("Setup done; running tests")
	status := m.Run()
	log.Info("Cleaning up test message queue")

	// You can't defer this because os.Exit doesn't care for defer
	cleanup()
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
	dummyEvent := &fes.DummyEvent{Msg: "dummy"}
	event, err := fes.NewEvent(key, dummyEvent)
	assert.NoError(t, err)
	err = backend.Append(event)
	assert.NoError(t, err)

	// check
	events, err := backend.Get(key)
	assert.NoError(t, err)
	assert.Len(t, events, 1)
	assert.Equal(t, event.GetType(), events[0].GetType())
	assert.Equal(t, event.GetTimestamp().GetNanos(), events[0].GetTimestamp().GetNanos())
	data, err := fes.UnmarshalEventData(events[0])
	assert.NoError(t, err)
	assert.Equal(t, dummyEvent, data)
}

func TestNatsBackend_List(t *testing.T) {
	subjects, err := backend.List(fes.ContainsMatcher(""))
	assert.NoError(t, err)
	assert.NotEmpty(t, subjects)
}
