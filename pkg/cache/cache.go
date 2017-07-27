package cache

type Cache interface {
	Put(key string, data interface{}) error
	Get(Key string) (interface{}, bool)
	Delete(key string) error
	List() map[string]interface{}
}

type NoCache struct{}

func (nc *NoCache) Put(key string, data interface{}) error {
	return nil
}

func (nc *NoCache) Get(Key string) (interface{}, bool) {
	return nil, false
}

func (nc *NoCache) Delete(key string) error {
	return nil
}

func (nc *NoCache) List() map[string]interface{} {
	return map[string]interface{}{}
}

type MapCache struct {
	store map[string]interface{}
}

func NewMapCache() *MapCache {
	return &MapCache{
		store: map[string]interface{}{},
	}
}

func (mc *MapCache) Put(key string, data interface{}) error {
	mc.store[key] = data
	return nil
}

func (mc *MapCache) Get(key string) (interface{}, bool) {
	val, exists := mc.store[key]
	return val, exists
}

func (mc *MapCache) Delete(key string) error {
	delete(mc.store, key)
	return nil
}

func (mc *MapCache) List() map[string]interface{} {
	return mc.store
}
