package parse

import (
	"io"

	"github.com/fission/fission-workflows/pkg/parse/yaml"
	"github.com/fission/fission-workflows/pkg/types"
)

type Parser interface {
	Parse(r io.Reader) (*types.WorkflowSpec, error)
}

func Parse(r io.Reader) (*types.WorkflowSpec, error) {
	return yaml.Parse(r)
}
