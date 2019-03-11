package cache

import (
	"testing"

	"github.com/fission/fission-workflows/pkg/fes"
	"github.com/fission/fission-workflows/pkg/fes/testutil"
	"github.com/stretchr/testify/assert"
)

func setupLoadingCache() (*LoadingCache, *testutil.Cache, *testutil.Backend) {
	backingCache := testutil.NewCache()
	backend := testutil.NewBackend()
	cache := NewLoadingCache(backingCache, backend, testutil.Projector)
	return cache, backingCache, backend
}

func TestLoadingCache_GetAggregateStoreAndCache(t *testing.T) {
	cache, _, eventStore := setupLoadingCache()
	key := fes.Aggregate{Type: testutil.MockEntityType, Id: "1"}
	target := "abcdef"
	events := testutil.ToDummyEvents(key, target)
	for _, event := range events {
		err := eventStore.Append(event)
		assert.NoError(t, err)
	}

	// First fetch: load from event store
	e, err := cache.GetAggregate(key)
	assert.NoError(t, err)
	storedEntity, ok := e.(*testutil.MockEntity)
	assert.True(t, ok)
	assert.Equal(t, target, storedEntity.S)

	// Second fetch: load from cache (ensure no event store fallback by clearing the event store).
	eventStore.Reset()
	e, err = cache.GetAggregate(key)
	assert.NoError(t, err)
	cachedEntity, ok := e.(*testutil.MockEntity)
	assert.True(t, ok)
	assert.Equal(t, target, cachedEntity.S)
}

func TestLRUCache(t *testing.T) {
	cache := NewLRUCache(2)
	e1, _ := testutil.Projector.NewProjection(fes.Aggregate{Type: "test", Id: "1"})
	e2, _ := testutil.Projector.NewProjection(fes.Aggregate{Type: "test", Id: "2"})
	e3, _ := testutil.Projector.NewProjection(fes.Aggregate{Type: "test", Id: "3"})
	err := cache.Put(e1)
	assert.NoError(t, err)
	err = cache.Put(e2)
	assert.NoError(t, err)
	err = cache.Put(e3)
	assert.NoError(t, err)

	assert.Equal(t, len(cache.List()), 2)
	_, err = cache.GetAggregate(fes.GetAggregate(e1))
	assert.True(t, fes.ErrEntityNotFound.Is(err))
	c2, err := cache.GetAggregate(fes.GetAggregate(e2))
	assert.EqualValues(t, e2, c2)
	c3, err := cache.GetAggregate(fes.GetAggregate(e3))
	assert.EqualValues(t, e3, c3)
}
