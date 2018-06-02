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

type InSelector struct {
	Key    string
	Values []string
}

func In(key string, values ...string) *InSelector {
	return &InSelector{
		Key:    key,
		Values: values,
	}
}

func (s *InSelector) Matches(labels Labels) bool {
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

type AndSelector struct {
	Selectors []Selector
}

func And(reqs ...Selector) *AndSelector {
	return &AndSelector{reqs}
}

func (s *AndSelector) Matches(labels Labels) bool {
	for _, req := range s.Selectors {
		valid := req.Matches(labels)
		if !valid {
			return false
		}
	}

	return true
}

type OrSelector struct {
	Selectors []Selector
}

func Or(reqs ...Selector) *OrSelector {
	return &OrSelector{reqs}
}

func (s *OrSelector) Matches(labels Labels) bool {
	for _, req := range s.Selectors {
		valid := req.Matches(labels)
		if valid {
			return true
		}
	}

	return false
}
