package expr

import (
	"sync"
)

// TODO Keep old states (but prune if OOM)
// TODO provide garbage collector
type Store struct {
	entries  sync.Map // map[string]interface{}
	resolver Resolver
}

func NewStore() *Store {
	return &Store{
		entries:  sync.Map{},
		resolver: DefaultResolver,
	}
}

func (rs *Store) Set(id string, data interface{}) {
	rs.entries.Store(id, data)
}

func (rs *Store) Delete(id string) {
	rs.entries.Delete(id)
}

func (rs *Store) Get(id string) (interface{}, bool) {
	return rs.entries.Load(id)
}

func (rs *Store) Update(id string, updater func(entry interface{}) interface{}) {
	entry, ok := rs.Get(id)
	if ok {
		rs.Set(id, updater(entry))
	}
}

// Range calls f sequentially for each key and value present in the map.
// If f returns false, range stops the iteration.
func (rs *Store) Range(fn func(key string, value interface{}) bool) {
	rs.entries.Range(func(key, value interface{}) bool {
		return fn(key.(string), value)
	})
}
