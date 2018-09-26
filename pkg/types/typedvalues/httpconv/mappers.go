package httpconv

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"

	"github.com/fission/fission-workflows/pkg/types/typedvalues"
	"github.com/fission/fission-workflows/pkg/util/mediatype"
	"github.com/gogo/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
)

// TODO set original content type as metadata
// TODO support multipart/form-data

var (
	// Common media types
	MediaTypeBytes    = mediatype.MustParse("application/octet-stream")
	MediaTypeJSON     = mediatype.MustParse("application/json")
	MediaTypeProtobuf = mediatype.MustParse("application/protobuf")
	MediaTypeText     = mediatype.MustParse("text/plain")

	// Media type parameter for protobuf message addressing.
	messageTypeParam       = "proto"
	ErrMessageTypeNotFound = errors.New("media type did not contain message type parameter")

	// Default implementations
	jsonMapper     = &JSONMapper{}
	textMapper     = &TextMapper{}
	bytesMapper    = &BytesMapper{}
	protobufMapper = &ProtobufMapper{}
)

type ParserFormatter interface {
	Formatter
	Parser
}

type Parser interface {
	Parse(mt *mediatype.MediaType, reader io.Reader) (*typedvalues.TypedValue, error)
}

type Formatter interface {
	Format(w http.ResponseWriter, body *typedvalues.TypedValue) error
}

type BytesMapper struct {
}

func (p *BytesMapper) Parse(mt *mediatype.MediaType, reader io.Reader) (*typedvalues.TypedValue, error) {
	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return typedvalues.Wrap(bs)
}

func (p *BytesMapper) Format(w http.ResponseWriter, body *typedvalues.TypedValue) error {
	var bs []byte
	switch body.ValueType() {
	case typedvalues.TypeString:
		s, err := typedvalues.UnwrapString(body)
		if err != nil {
			return err
		}
		bs = []byte(s)
	case typedvalues.TypeBytes:
		b, err := typedvalues.UnwrapBytes(body)
		if err != nil {
			return err
		}
		bs = b
	default:
		return typedvalues.ErrUnsupportedType
	}
	_, err := w.Write(bs)
	if err != nil {
		return err
	}

	mediatype.SetContentTypeHeader(MediaTypeBytes, w)
	return nil
}

type JSONMapper struct {
}

func (m *JSONMapper) Format(w http.ResponseWriter, body *typedvalues.TypedValue) error {
	i, err := typedvalues.Unwrap(body)
	if err != nil {
		return err
	}
	bs, err := json.Marshal(i)
	if err != nil {
		return errors.WithStack(err)
	}
	_, err = w.Write(bs)
	if err != nil {
		return err
	}
	mt := MediaTypeJSON.Copy()
	mt.SetParam(messageTypeParam, body.GetValue().GetTypeUrl())
	mediatype.SetContentTypeHeader(mt, w)
	return nil
}

func (m *JSONMapper) Parse(mt *mediatype.MediaType, reader io.Reader) (*typedvalues.TypedValue, error) {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	// We borrow the proto parameter to do type inference
	if msgName, ok := mt.Parameters[messageTypeParam]; ok {
		// try to use type
		tv, err := m.parseJSONWithType(data, msgName)
		if err != nil {
			return nil, err
		}
		return tv, nil
	}

	// Alternatively, we use JSON's dynamic structure
	var i interface{}
	err = json.Unmarshal(data, &i)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return typedvalues.Wrap(i)
}

func (m *JSONMapper) parseJSONWithType(data []byte, msgName string) (*typedvalues.TypedValue, error) {
	msgType := proto.MessageType(msgName)
	if msgType == nil {
		return nil, ErrMessageTypeNotFound
	}

	msg := reflect.New(msgType).Interface()
	err := json.Unmarshal(data, msg)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return typedvalues.Wrap(msg)
}

type TextMapper struct{}

func (m *TextMapper) Format(w http.ResponseWriter, body *typedvalues.TypedValue) error {
	// The TextMapper is permissive, all types that make sense to convert to a string, are supported.
	i, err := typedvalues.Unwrap(body)
	if err != nil {
		return err
	}
	var output string
	switch body.ValueType() {
	case typedvalues.TypeExpression:
		fallthrough
	case typedvalues.TypeString:
		output = i.(string)
	case typedvalues.TypeBytes:
		output = string(i.([]byte))
	case typedvalues.TypeUInt64:
		fallthrough
	case typedvalues.TypeUInt32:
		fallthrough
	case typedvalues.TypeInt32:
		fallthrough
	case typedvalues.TypeInt64:
		fallthrough
	case typedvalues.TypeFloat32:
		fallthrough
	case typedvalues.TypeFloat64:
		output = fmt.Sprintf("%d", i)
	default:
		return typedvalues.ErrUnsupportedType
	}
	_, err = w.Write([]byte(output))
	if err != nil {
		return err
	}
	mediatype.SetContentTypeHeader(MediaTypeText, w)
	return nil
}

func (m *TextMapper) Parse(mt *mediatype.MediaType, reader io.Reader) (*typedvalues.TypedValue, error) {
	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}
	return typedvalues.Wrap(string(bs))
}

type ProtobufMapper struct{}

func (p *ProtobufMapper) Format(w http.ResponseWriter, body *typedvalues.TypedValue) error {
	_, err := w.Write(body.GetValue().Value)
	if err != nil {
		return err
	}

	mt := MediaTypeProtobuf.Copy()
	mt.SetParam(messageTypeParam, body.GetValue().GetTypeUrl())
	mediatype.SetContentTypeHeader(mt, w)
	return nil
}

func (p *ProtobufMapper) Parse(mt *mediatype.MediaType, reader io.Reader) (*typedvalues.TypedValue, error) {
	// Do not check if the MediaType actually is application/protobuf, to allow users to parse mislabeled MediaTypes.

	// Fetch the message type from the media type parameters
	msgName, ok := mt.Parameters[messageTypeParam]
	if !ok {
		return nil, ErrMessageTypeNotFound
	}

	switch mt.Suffix {
	case "json":
		return p.parseJSONWithType(reader, msgName)
	default:
		return p.parseProtoWithType(reader, msgName)
	}
}

func (p *ProtobufMapper) parseJSONWithType(reader io.Reader, msgName string) (*typedvalues.TypedValue, error) {
	msgType := proto.MessageType(msgName)
	if msgType == nil {
		return nil, ErrMessageTypeNotFound
	}

	msg := reflect.New(msgType).Interface().(proto.Message)
	err := jsonpb.Unmarshal(reader, msg)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return typedvalues.Wrap(msg)
}

func (p *ProtobufMapper) parseProtoWithType(reader io.Reader, msgName string) (*typedvalues.TypedValue, error) {
	bs, err := ioutil.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	msgType := proto.MessageType(msgName)
	if msgType == nil {
		return nil, ErrMessageTypeNotFound
	}

	msg := reflect.New(msgType).Interface().(proto.Message)
	err = proto.Unmarshal(bs, msg)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return typedvalues.Wrap(msg)
}
