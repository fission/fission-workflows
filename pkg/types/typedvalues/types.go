package typedvalues

import (
	"strings"

	"github.com/fission/fission-workflow/pkg/types"
)

type Parser interface {
	Parse(i interface{}) (*types.TypedValue, error) // TODO allow hint of type
}

type Formatter interface {
	Format(v *types.TypedValue) (interface{}, error)
}

type ParserFormatter interface {
	Parser
	Formatter
}

// Splits valueTypes of format '<language>/<type>' into (format, type)
func ParseType(valueType string) (format string, subType string) {
	parts := strings.SplitN(valueType, "/", 2)

	if len(parts) == 0 {
		return "", ""
	}

	if len(parts) == 1 {
		switch parts[0] {
		case TYPE_EXPRESSION:
			fallthrough
		case TYPE_RAW:
			return FORMAT_RESERVED, parts[0]
		default:
			return parts[0], ""
		}
	}

	return parts[0], parts[1]
}

func FormatType(parts ...string) string {
	return strings.Join(parts[:1], "/")
}

func isFormat(targetValueType string, format string) bool {
	f, _ := ParseType(targetValueType)
	return strings.EqualFold(f, format)
}

func NewDefaultParserFormatter() ParserFormatter {
	return NewComposedParserFormatter(map[string]ParserFormatter{
		FormatType(FORMAT_JSON) : &JsonParserFormatter{},
		FormatType(TYPE_EXPRESSION) : &ExprParserFormatter{},
		FormatType(TYPE_RAW) : &RawParserFormatter{},
	})
}
