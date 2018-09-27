package typedvalues

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTypesGenerated(t *testing.T) {
	for i, v := range Types {
		t.Run(fmt.Sprintf("Types[%d]", i), func(t *testing.T) {
			fmt.Printf("Types[%d] = %s\n", i, v)
			assert.NotEmpty(t, v)
		})
	}
}
