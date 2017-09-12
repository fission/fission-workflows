package functionref

import (
	"testing"

	"github.com/fission/fission"
	"github.com/stretchr/testify/assert"
)

func TestSelfParser_Parse(t *testing.T) {
	subject := fission.FunctionReference{
		Type: "foo",
		Name: "bar",
	}
	parser := &SelfParser{}
	ref, err := parser.Parse(subject)
	assert.NoError(t, err)
	assert.Equal(t, subject, ref)
}

func TestSelfParser_Parse_Failed(t *testing.T) {
	subject := "foo"
	parser := &SelfParser{}
	_, err := parser.Parse(subject)
	assert.Error(t, err)
}

func TestNameParserFormatter_Parse(t *testing.T) {
	subject := "foo"
	parser := &NameParserFormatter{}
	ref, err := parser.Parse(subject)
	assert.NoError(t, err)
	assert.Equal(t, fission.FunctionReference{
		Type: fission.FunctionReferenceTypeFunctionName,
		Name: subject,
	}, ref)
}

func TestNameParserFormatter_Parse_Failed(t *testing.T) {
	subject := fission.FunctionReference{
		Type: "foo",
		Name: "bar",
	}
	parser := &NameParserFormatter{}
	_, err := parser.Parse(subject)
	assert.Error(t, err)
}

func TestNameParserFormatter_Format(t *testing.T) {
	subject := fission.FunctionReference{
		Type: fission.FunctionReferenceTypeFunctionName,
		Name: "bar",
	}
	parser := &NameParserFormatter{}
	formatted, err := parser.Format(subject)
	assert.NoError(t, err)
	assert.Equal(t, subject.Name, formatted)
}

func TestNameParserFormatter_Format_Failed(t *testing.T) {
	subject := fission.FunctionReference{
		Type: "someOtherType",
		Name: "bar",
	}
	parser := &NameParserFormatter{}
	_, err := parser.Format(subject)
	assert.Error(t, err)
}
