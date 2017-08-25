package typedvalues

import (
	"errors"
	"fmt"
	"strings"

	"github.com/fission/fission-workflow/pkg/types"
)

const (
	FORMAT_RESERVED = "reserved"
	TYPE_EXPRESSION = "expr"
	TYPE_RAW        = "raw"
)

func Expr(expr string) *types.TypedValue {
	return &types.TypedValue{
		Type: FormatType(TYPE_EXPRESSION),
		Value: []byte(expr),
	}
}

func IsExpression(value *types.TypedValue) bool {
	return value.Type == TYPE_EXPRESSION
}

// RawParserFormatter converts []byte values to TypedValue, without any formatting or parsing.
type RawParserFormatter struct{}

func (dp *RawParserFormatter) Parse(i interface{}) (*types.TypedValue, error) {
	b, ok := i.([]byte)
	if !ok {
		return nil, errors.New("Provided value is not of type '[]byte'")
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

func (ExprParserFormatter) Parse(i interface{}) (*types.TypedValue, error) {
	s, ok := i.(string)

	if !ok {
		return nil, errors.New("Provided value is not of type 'string'")
	}
	// Normalize
	ns := strings.TrimSpace(s)

	// Check if the string is an expression
	if !strings.HasPrefix(ns, "$") { // TODO add support for expressions other than selectors
		return nil, errors.New("Provided value is not of type 'expression string'")
	}

	return &types.TypedValue{
		Type:  TYPE_EXPRESSION,
		Value: []byte(ns),
	}, nil
}

func (ExprParserFormatter) Format(v *types.TypedValue) (interface{}, error) {
	if isFormat(v.Type, TYPE_EXPRESSION) {
		return nil, fmt.Errorf("Value '%v' is not of type 'expr'", v)
	}

	return string(v.Value), nil
}

// Used to group multiple ParserFormatters together (e.g. RefParserFormatter + JsonParserFormatter + XmlParserFormatter)
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

func (cp *ComposedParserFormatter) Parse(i interface{}) (result *types.TypedValue, err error) {
	if tv, ok := i.(*types.TypedValue); ok {
		return tv, nil
	}

	for _, p := range cp.priorities {
		result, err = cp.pfs[p].Parse(i)
		if err == nil && result != nil {
			break
		}
	}
	if err != nil {
		return nil, fmt.Errorf("Failed to parse value '%v'", i)
	}
	return result, nil
}

func (cp *ComposedParserFormatter) Format(v *types.TypedValue) (interface{}, error) {
	formatter, ok := cp.pfs[v.Type]
	if !ok {
		return nil, fmt.Errorf("TypedValue '%v' has unknown type '%v'", v, v.Type)
	}
	return formatter.Format(v)
}
