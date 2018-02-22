package protobuf

import (
	"io"
	"io/ioutil"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/gogo/protobuf/proto"
)

// Parser implements the parse.Parser interface to parse workflow specs from a protobuf encoding.
type Parser struct {
}

// Parse parses a workflow spec from the provided reader, assuming that the protobuf encoding is used.
//
// It will return an error if (1) it fails to read the reader, or if (2) it fails to unmarshall the bytes
// into the WorkflowSpec.
func (p *Parser) Parse(r io.Reader) (*types.WorkflowSpec, error) {
	bs, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var wf *types.WorkflowSpec
	err = proto.Unmarshal(bs, wf)
	if err != nil {
		return nil, err
	}

	return wf, nil
}
