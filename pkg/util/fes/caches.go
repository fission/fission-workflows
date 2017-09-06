package fes

// Cache
type MapCache struct {
	contents map[string]map[string]Aggregator // Map: AggregateType -> AggregateId -> entity
}

func NewMapCache() CacheReaderWriter {
	return &MapCache{
		contents: map[string]map[string]Aggregator{},
	}
}

func (rc *MapCache) Get(entity Aggregator) error {
	ref := entity.Aggregate()
	err := validateAggregate(ref)
	if err != nil {
		return err
	}

	aType, ok := rc.contents[ref.Type]
	if !ok {
		return nil
	}

	cached, ok := aType[ref.Id]
	if !ok {
		return nil
	}

	return entity.UpdateState(cached)
}

func (rc *MapCache) Put(entity Aggregator) error {
	ref := entity.Aggregate()
	err := validateAggregate(ref)
	if err != nil {
		return err
	}

	if _, ok := rc.contents[ref.Type]; !ok {
		rc.contents[ref.Type] = map[string]Aggregator{}
	}

	rc.contents[ref.Type][ref.Id] = entity
	return nil
}
