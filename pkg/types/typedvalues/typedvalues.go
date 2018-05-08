package typedvalues

import (
	"errors"
	"fmt"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/sirupsen/logrus"
)

type TypedValue = types.TypedValue

var (
	// TODO unsupported -> non-fatal
	ErrUnsupportedType = errors.New("unsupported type")  // Error to indicate parserFormatter cannot handle type
	ErrValueConversion = errors.New("failed to convert") // Error to indicate parserFormatter internal error
)

type TypedValueErr struct {
	src interface{}
	err error
}

func (tve TypedValueErr) Error() string {
	return fmt.Sprintf("%v (%+v)", tve.err.Error(), tve.src)
}

type ValueType = string

type Parser interface {
	Parse(ctx Parser, i interface{}) (*types.TypedValue, error)
}

type Formatter interface {
	Format(ctx Formatter, v *types.TypedValue) (interface{}, error)
}

type ParserFormatter interface {
	Accepts() []ValueType
	Parser
	Formatter
}

var DefaultParserFormatter = newDefaultParserFormatter()

func Parse(i interface{}) (*types.TypedValue, error) {
	return DefaultParserFormatter.Parse(DefaultParserFormatter, i)
}

func Format(v *types.TypedValue) (interface{}, error) {
	return DefaultParserFormatter.Format(DefaultParserFormatter, v)
}

// MustParse transforms the value into a TypedValue or panics.
func MustParse(i interface{}) *types.TypedValue {
	tv, err := Parse(i)
	if err != nil {
		panic(err)
	}
	return tv
}

// MustParse transforms the TypedValue into a value or panics.
func MustFormat(tv *types.TypedValue) interface{} {
	i, err := Format(tv)
	if err != nil {
		panic(err)
	}
	return i
}

// ComposedParserFormatter is used to group multiple ParserFormatters together
type ComposedParserFormatter struct {
	formatters map[ValueType][]Formatter
	parsers    []Parser // TODO also limit to types for parsers
	supported  []ValueType
}

func (pf *ComposedParserFormatter) Accepts() []ValueType {
	return pf.supported
}

func (pf *ComposedParserFormatter) Parse(ctx Parser, i interface{}) (*types.TypedValue, error) {
	for _, parser := range pf.parsers {
		logrus.Debugf("Trying to parse with: %T", parser)
		tv, err := parser.Parse(ctx, i)
		if err != nil {
			logrus.Debugf("Parser error %t", err)
			if isErrUnsupported(err) {
				continue
			} else {
				return nil, err
			}
		}
		logrus.Debugf("Parser success: %t", tv)
		return tv, nil
	}
	logrus.Debugf("No parsers for %t", i)
	return nil, TypedValueErr{
		src: i,
		err: ErrUnsupportedType,
	}
}

func (pf *ComposedParserFormatter) Format(ctx Formatter, v *types.TypedValue) (interface{}, error) {
	if v == nil {
		return nil, nil
	}
	formatters, ok := pf.formatters[ValueType(v.Type)]
	if !ok {
		logrus.Debugf("No known formatter for: %T", v)
		return 0, TypedValueErr{
			src: v,
			err: ErrUnsupportedType,
		}
	}
	logrus.Debugf("Formatter options for %v: %T", v.Type, formatters)
	for _, formatter := range formatters {
		logrus.Debugf("Trying to format with: %T", formatter)
		tv, err := formatter.Format(ctx, v)
		if err != nil {
			logrus.Debugf("Formatter error %t", err)
			if isErrUnsupported(err) {
				continue
			} else {
				return nil, err
			}
		}
		logrus.Debugf("Formatter success: %t", tv)
		return tv, nil
	}
	return nil, TypedValueErr{
		src: v,
		err: ErrUnsupportedType,
	}
}

func NewComposedParserFormatter(pfs []ParserFormatter) *ComposedParserFormatter {
	formatters := map[ValueType][]Formatter{}
	var parsers []Parser

	for _, v := range pfs {
		parsers = append(parsers, v)
		for _, t := range v.Accepts() {
			vts := formatters[t]
			if vts == nil {
				vts = []Formatter{}
			}
			vts = append(vts, v)
			formatters[t] = vts
		}
	}

	var ts []ValueType
	for k := range formatters {
		ts = append(ts, k)
	}

	return &ComposedParserFormatter{
		formatters: formatters,
		parsers:    parsers,
		supported:  ts,
	}
}

func isErrUnsupported(err error) bool {
	if tverr, ok := err.(TypedValueErr); ok {
		return tverr.err == ErrUnsupportedType
	}
	return false
}

func newDefaultParserFormatter() ParserFormatter {
	return NewComposedParserFormatter([]ParserFormatter{
		&IdentityParserFormatter{},
		&MapParserFormatter{},
		&ListParserFormatter{},
		&ControlFlowParserFormatter{},
		&ExpressionParserFormatter{},
		&BoolParserFormatter{},
		&NumberParserFormatter{},
		&StringParserFormatter{},
		&NilParserFormatter{},
		&BytesParserFormatter{},
	})
}
