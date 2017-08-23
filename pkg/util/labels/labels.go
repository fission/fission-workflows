package labels

import (
	"fmt"
)

type Labels interface {
	fmt.Stringer
}

type Selector interface {
	Matches(labels Labels) bool
}
