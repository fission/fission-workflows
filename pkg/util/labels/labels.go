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
