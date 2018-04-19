package typedvalues

import (
	"regexp"

	"github.com/fission/fission-workflows/pkg/types"
)

const (
	TypeExpression ValueType = "expression"
)

var (
	ExpressionRe = regexp.MustCompile("\\{.*\\}")
)

type ExpressionParserFormatter struct{}

func (pf *ExpressionParserFormatter) Accepts() []ValueType {
	return []ValueType{
		TypeExpression,
	}
}

func (pf *ExpressionParserFormatter) Parse(ctx Parser, i interface{}) (*types.TypedValue, error) {
	s, ok := i.(string)
	if !ok {
		return nil, TypedValueErr{
			src: i,
			err: ErrUnsupportedType,
		}
	}

	return ParseExpression(s)
}

func (pf *ExpressionParserFormatter) Format(ctx Formatter, v *types.TypedValue) (interface{}, error) {
	return FormatExpression(v)
}

func ParseExpression(s string) (*types.TypedValue, error) {
	if !ExpressionRe.MatchString(s) {
		return nil, TypedValueErr{
			src: s,
			err: ErrUnsupportedType,
		}
	}

	return &types.TypedValue{
		Type:  string(TypeExpression),
		Value: []byte(s),
	}, nil
}

func FormatExpression(v *types.TypedValue) (string, error) {
	if ValueType(v.Type) != TypeExpression {
		return "", TypedValueErr{
			src: v,
			err: ErrUnsupportedType,
		}
	}
	return string(v.Value), nil
}
