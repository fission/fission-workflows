package functionref

import (
	"errors"

	"github.com/fission/fission"
)

var (
	ErrInvalidType  = errors.New("invalid type")
	ErrParseFailed  = errors.New("failed to parse")
	ErrFormatFailed = errors.New("failed to format")
)

type Parser interface {
	Parse(ref interface{}) (fission.FunctionReference, error)
}

type Formatter interface {
	Format(ref fission.FunctionReference) (string, error)
}

var DefaultParser Parser = &ComposedParser{[]Parser{
	&SelfParser{},
	&NameParserFormatter{},
}}

var DefaultFormatter Formatter = &ComposedFormatter{[]Formatter{
	&NameParserFormatter{},
}}

func Format(ref fission.FunctionReference) (string, error) {
	return DefaultFormatter.Format(ref)
}

func Parse(ref interface{}) (fission.FunctionReference, error) {
	return DefaultParser.Parse(ref)
}

type ComposedParser struct {
	Parsers []Parser
}

func (cp *ComposedParser) Parse(ref interface{}) (fission.FunctionReference, error) {
	for _, parser := range cp.Parsers {
		result, err := parser.Parse(ref)
		if err == nil {
			return result, nil
		}
	}
	return fission.FunctionReference{}, ErrParseFailed
}

type ComposedFormatter struct {
	Formatters []Formatter
}

func (cf *ComposedFormatter) Format(ref fission.FunctionReference) (string, error) {
	for _, formatter := range cf.Formatters {
		result, err := formatter.Format(ref)
		if err == nil {
			return result, nil
		}
	}
	return "", ErrFormatFailed
}

type SelfParser struct{}

func (sp *SelfParser) Parse(ref interface{}) (fission.FunctionReference, error) {
	fnRef, ok := ref.(fission.FunctionReference)
	if !ok {
		return fission.FunctionReference{}, ErrInvalidType
	}
	return fnRef, nil
}

type NameParserFormatter struct{}

func (np *NameParserFormatter) Format(ref fission.FunctionReference) (string, error) {
	if ref.Type != fission.FunctionReferenceTypeFunctionName {
		return "", ErrInvalidType
	}

	return ref.Name, nil
}

func (np *NameParserFormatter) Parse(ref interface{}) (fission.FunctionReference, error) {
	refString, ok := ref.(string)
	if !ok {
		return fission.FunctionReference{}, ErrInvalidType
	}

	return fission.FunctionReference{
		Type: fission.FunctionReferenceTypeFunctionName,
		Name: refString,
	}, nil
}
