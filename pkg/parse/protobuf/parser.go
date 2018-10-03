package protobuf

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"

	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
)

var DefaultParser = &Parser{}

func Parse(r io.Reader) (*types.WorkflowSpec, error) {
	return DefaultParser.Parse(r)
}

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
	wf := &types.WorkflowSpec{}
	protoErr := proto.Unmarshal(bs, wf)
	if protoErr != nil {
		// Fallback: it might be protobuf serialized into json.
		jsonErr := jsonpb.Unmarshal(bytes.NewReader(bs), wf)
		if jsonErr != nil {
			return nil, fmt.Errorf("failed to parse protobuf-encoded workflow (proto: %v, jsonpb: %v)", protoErr, jsonErr)
		}
	}
	return wf, nil
}
