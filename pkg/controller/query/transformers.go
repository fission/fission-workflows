package query

import (
	"fmt"
	"strings"

	"reflect"

	"github.com/fission/fission-workflow/pkg/types"
	"github.com/fission/fission-workflow/pkg/types/typedvalues"
)

type DataTransformer interface {
	Apply(in ...*types.TypedValue) (*types.TypedValue, error) // TODO explicitly only allow literals / resolved value
}

// Built-in data transformers
type TransformerUppercase struct {
	pf typedvalues.ParserFormatter
}

func (tf *TransformerUppercase) Apply(in *types.TypedValue) (*types.TypedValue, error) {
	_, t := typedvalues.ParseType(in.Type)
	if t != typedvalues.TYPE_STRING {
		return nil, fmt.Errorf("Could not uppercase value of unsupported type '%v'", in)
	}

	i, err := tf.pf.Format(in)
	if err != nil {
		return nil, err
	}

	s, ok := i.(string)
	if !ok {
		return nil, fmt.Errorf("TypedValue '%v' did not parse to expected type 'string', but '%s'", in, reflect.TypeOf(s))
	}

	us := strings.ToUpper(s)
	return tf.pf.Parse(us)
}
