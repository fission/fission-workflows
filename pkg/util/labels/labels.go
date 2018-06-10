// Package labels provides storing, fetching and matching based on labels.
package labels

// Labels is the interface for fetching labels.
type Labels interface {
	// Get returns the value for the provided label.
	Get(label string) (value string, exists bool)
}

// Matcher is the interface for matching objects based on their labels.
type Matcher interface {
	Matches(labels Labels) bool
}

// Set is a set of labels, in which a key can only occur once.
type Set map[string]string

// Get fetches a label based on the key. If the label exists the value will be returned.
//
// If the set is nil, Get will simply return that the label does not exist either.
func (l Set) Get(label string) (value string, exists bool) {
	if l == nil {
		return "", false
	}
	value, exists = l[label]
	return value, exists
}

// Set stores a label with the associated value.
//
// If the set is nil, a set will be created.
// If a label with the same name exists already in the set, it will be overwritten.
func (l *Set) Set(label, value string) {
	if l == nil {
		*l = map[string]string{}
	}
	(*l)[label] = value
}

// InMatcher matches all labels where the label keys equals the key, and (optionally)
// the value has to be in the set of accepted values.
type InMatcher struct {
	Key    string
	Values []string
}

// In matches all labels where the label keys equals the key, and (optionally)
// the value has to be in the set of accepted values.
func In(key string, values ...string) InMatcher {
	return InMatcher{
		Key:    key,
		Values: values,
	}
}

// Matches returns true if the labels are selected by one of the matchers in the inMatches.
func (s InMatcher) Matches(labels Labels) bool {
	val, ok := labels.Get(s.Key)
	if !ok {
		return false
	}
	for _, v := range s.Values {
		if val == v {
			return true
		}
	}

	return false
}

// AndMatcher is a composite matcher that selects all labels that are selected by all matchers.
type AndMatcher struct {
	Matchers []Matcher
}

// And returns a composite matcher that selects all labels that are selected by all matchers.
func And(ms ...Matcher) AndMatcher {
	return AndMatcher{ms}
}

// Matches returns true if the labels are selected by one of the matchers in the andMatchers.
func (s AndMatcher) Matches(labels Labels) bool {
	for _, req := range s.Matchers {
		if !req.Matches(labels) {
			return false
		}
	}
	return true
}

// OrMatcher is a composite matcher that selects all labels that are selected by one of the matchers.
type OrMatcher struct {
	matchers []Matcher
}

// Or returns a composite matcher that selects all labels that are selected by one of the matchers.
func Or(ms ...Matcher) OrMatcher {
	return OrMatcher{ms}
}

// Matches returns true if the labels are selected by one of the matchers in the orMatchers.
func (s OrMatcher) Matches(labels Labels) bool {
	for _, req := range s.matchers {
		if req.Matches(labels) {
			return true
		}
	}
	return false
}
