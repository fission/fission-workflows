package typedvalues

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fission/fission-workflows/pkg/types"
)

const (
	FORMAT_RESERVED = "reserved"
	TYPE_EXPRESSION = "expr"
	TYPE_RAW        = "raw"
	TYPE_NIL        = "empty"
)

func Expr(expr string) *types.TypedValue {
	return &types.TypedValue{
		Type:  FormatType(TYPE_EXPRESSION),
		Value: []byte(expr),
	}
}

func IsExpression(value *types.TypedValue) bool {
	return value.Type == TYPE_EXPRESSION
}

// RawParserFormatter converts []byte values to TypedValue, without any formatting or parsing.
type RawParserFormatter struct{}

func (dp *RawParserFormatter) Parse(i interface{}, allowedTypes ...string) (*types.TypedValue, error) {
	b, ok := i.([]byte)
	if !ok {
		return nil, errors.New("provided value is not of type '[]byte'")
	}

	return &types.TypedValue{
		Type:  TYPE_RAW,
		Value: b,
	}, nil
}

func (dp *RawParserFormatter) Format(v *types.TypedValue) (interface{}, error) {
	// Ignore any type checking here, as the value is always a []byte
	return v.Value, nil
}

// ExprParserFormatter parses and formats expressions to and from valid expression-strings
type ExprParserFormatter struct{}

func (ep *ExprParserFormatter) Parse(i interface{}, allowedTypes ...string) (*types.TypedValue, error) {
	s, ok := i.(string)

	if !ok {
		return nil, errors.New("provided value is not of type 'string'")
	}
	// Normalize
	ns := strings.TrimSpace(s)

	// Check if the string is an expression
	if !strings.HasPrefix(ns, "{") || !strings.HasSuffix(ns, "}") {
		return nil, errors.New("provided value is not of type 'expression string'")
	}

	return &types.TypedValue{
		Type:  TYPE_EXPRESSION,
		Value: []byte(ns),
	}, nil
}

func (ep *ExprParserFormatter) Format(v *types.TypedValue) (interface{}, error) {
	if IsFormat(v.Type, TYPE_EXPRESSION) {
		return nil, fmt.Errorf("value '%v' is not of type 'expr'", v)
	}

	return string(v.Value), nil
}

// ComposedParserFormatter is used to group multiple ParserFormatters together (e.g. RefParserFormatter +
// JsonParserFormatter + XmlParserFormatter)
type ComposedParserFormatter struct {
	pfs        map[string]ParserFormatter // Language : ParserFormatter
	priorities []string
}

func NewComposedParserFormatter(pfs map[string]ParserFormatter, order ...string) *ComposedParserFormatter {
	keys := map[string]interface{}{}
	priorities := []string{}
	// Filter out non-existent and duplicate keys from order
	for _, p := range order {
		_, seen := keys[p]
		if _, ok := pfs[p]; ok && !seen {
			priorities = append(priorities, p)
			keys[p] = nil
		}
	}

	// Ensure that all keys are present in the order list
	for k := range pfs {
		if _, seen := keys[k]; !seen {
			priorities = append(priorities, k)
		}
	}

	return &ComposedParserFormatter{
		pfs:        pfs,
		priorities: priorities,
	}
}

func (cp *ComposedParserFormatter) Parse(i interface{}, allowedTypes ...string) (result *types.TypedValue, err error) {
	if tv, ok := i.(*types.TypedValue); ok {
		return tv, nil
	}

	if len(allowedTypes) == 0 {
		allowedTypes = cp.priorities
	}

	for _, p := range allowedTypes {
		parser, ok := cp.pfs[p]
		if !ok {
			continue
		}
		result, err = parser.Parse(i)
		if err == nil && result != nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to parse %T value '%v': %v", i, i, err)
	}
	return result, nil
}

func (cp *ComposedParserFormatter) Format(v *types.TypedValue) (interface{}, error) {
	if v == nil {
		return nil, nil
	}

	formatter, ok := cp.pfs[v.Type]
	if !ok {
		return nil, fmt.Errorf("TypedValue '%v' has unknown type '%v'", v, v.Type)
	}
	return formatter.Format(v)
}

type NilParserFormatter struct {
}

func (NilParserFormatter) Parse(i interface{}, allowedTypes ...string) (*types.TypedValue, error) {
	if i == nil {
		return &types.TypedValue{
			Type: TYPE_NIL,
		}, nil
	}
	return nil, errors.New("value not nil")
}

func (NilParserFormatter) Format(v *types.TypedValue) (interface{}, error) {
	return nil, nil
}
