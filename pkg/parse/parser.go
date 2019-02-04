package parse

import (
	"errors"
	"io"

	"github.com/fission/fission-workflows/pkg/parse/protobuf"
	"github.com/fission/fission-workflows/pkg/parse/yaml"
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/sirupsen/logrus"
)

var (
	DefaultParser = NewMetaParser(map[string]Parser{
		"yaml": yaml.DefaultParser,
		"pb":   protobuf.DefaultParser,
	})
)

type Parser interface {
	Parse(r io.Reader) (*types.WorkflowSpec, error)
}

func Parse(r io.Reader) (*types.WorkflowSpec, error) {
	return DefaultParser.Parse(r)
}

func ParseWith(r io.Reader, parsers ...string) (*types.WorkflowSpec, error) {
	return DefaultParser.ParseWith(r, parsers...)
}

func Supports(parser string) bool {
	return DefaultParser.Supports(parser)
}

func Parsers() []string {
	return DefaultParser.Parsers()
}

type MetaParser struct {
	parsers map[string]Parser
}

func NewMetaParser(parsers map[string]Parser) *MetaParser {
	return &MetaParser{
		parsers: parsers,
	}
}

func (mp *MetaParser) Parse(r io.Reader) (*types.WorkflowSpec, error) {
	return mp.ParseWith(r, mp.Parsers()...)
}

func (mp *MetaParser) ParseWith(r io.Reader, parsers ...string) (*types.WorkflowSpec, error) {
	if parsers == nil {
		return nil, errors.New("no parsers provided")
	}
	var result *types.WorkflowSpec
	for _, name := range parsers {
		p, ok := mp.parsers[name]
		if !ok {
			continue
		}
		wf, err := p.Parse(r)
		if err != nil {
			logrus.WithField("parser", name).Warnf("parser failed: %v", err)
			wf = nil
			continue
		}
		result = wf
		break
	}
	var err error
	if result == nil {
		err = errors.New("failed to parse workflow")
	}
	return result, err
}

func (mp *MetaParser) Supports(s string) bool {
	_, ok := mp.parsers[s]
	return ok
}

func (mp *MetaParser) Parsers() []string {
	ps := make([]string, len(mp.parsers))
	var i int
	for name := range mp.parsers {
		ps[i] = name
		i++
	}
	return ps
}
