package labels

type Labels interface {
	// Has returns whether the provided label exists.
	Has(label string) (exists bool)

	// Get returns the value for the provided label.
	Get(label string) (value string)
}

type Selector interface {
	Matches(labels Labels) bool
}

type SetLabels map[string]string

func (l SetLabels) Has(label string) (exists bool) {
	_, ok := l[label]
	return ok
}

func (l SetLabels) Get(label string) (value string) {
	return l[label]
}

type inSelector struct {
	Key    string
	Values []string
}

func InSelector(key string, values ...string) *inSelector {
	return &inSelector{
		Key:    key,
		Values: values,
	}
}

func (s *inSelector) Matches(labels Labels) bool {
	if !labels.Has(s.Key) {
		return false
	}

	val := labels.Get(s.Key)
	for _, v := range s.Values {
		if val == v {
			return true
		}
	}

	return false
}

type andSelector struct {
	Selectors []Selector // And
}

func AndSelector(reqs ...Selector) *andSelector {
	return &andSelector{reqs}
}

func (s *andSelector) Matches(labels Labels) bool {
	for _, req := range s.Selectors {
		valid := req.Matches(labels)
		if !valid {
			return false
		}
	}

	return true
}

type orSelector struct {
	Selectors []Selector // And
}

func OrSelector(reqs ...Selector) *orSelector {
	return &orSelector{reqs}
}

func (s *orSelector) Matches(labels Labels) bool {
	for _, req := range s.Selectors {
		valid := req.Matches(labels)
		if valid {
			return true
		}
	}

	return false
}
