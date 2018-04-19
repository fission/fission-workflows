package typedvalues

import (
	"github.com/fission/fission-workflows/pkg/types"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
)

const (
	TypeBool   ValueType = "bool"
	TypeNumber ValueType = "number"
	TypeNil    ValueType = "nil"
	TypeString ValueType = "string"
	TypeBytes  ValueType = "bytes"
)

func IsPrimitive(v ValueType) bool {
	return v == TypeBool || v == TypeNumber || v == TypeString || v == TypeNil || v == TypeBytes
}

// var DefaultPrimitiveParserFormatter

type BoolParserFormatter struct{}

func (pf *BoolParserFormatter) Accepts() []ValueType {
	return []ValueType{
		TypeBool,
	}
}

func (pf *BoolParserFormatter) Parse(ctx Parser, i interface{}) (*types.TypedValue, error) {
	b, ok := i.(bool)
	if !ok {
		return nil, TypedValueErr{
			src: i,
			err: ErrUnsupportedType,
		}
	}

	return ParseBool(b), nil
}

func (pf *BoolParserFormatter) Format(ctx Formatter, v *types.TypedValue) (interface{}, error) {
	return FormatBool(v)
}

func ParseBool(b bool) *types.TypedValue {
	var v byte
	if b {
		v = 1
	}
	return &types.TypedValue{
		Type:  string(TypeBool),
		Value: []byte{v},
	}
}

func FormatBool(v *types.TypedValue) (bool, error) {
	err := verifyTypedValue(v, TypeBool)
	if err != nil {
		return false, err
	}
	if v.Value == nil || len(v.Value) < 1 {
		return false, TypedValueErr{
			src: v,
			err: ErrValueConversion,
		}
	}
	return v.Value[0] == 1, nil
}

type NumberParserFormatter struct {
}

func (pf *NumberParserFormatter) Accepts() []ValueType {
	return []ValueType{
		TypeNumber,
	}
}

func (pf *NumberParserFormatter) Parse(ctx Parser, i interface{}) (*types.TypedValue, error) {
	var f float64
	// TODO annotate typedvalue with additional type info
	switch t := i.(type) {
	case float32:
		f = float64(t)
	case float64:
		f = t
	case int64:
		f = float64(t)
	case int:
		f = float64(t)
	default:
		return nil, TypedValueErr{
			src: i,
			err: ErrUnsupportedType,
		}
	}
	return ParseNumber(f), nil
}

func (pf *NumberParserFormatter) Format(ctx Formatter, v *types.TypedValue) (interface{}, error) {
	return FormatNumber(v)
}

// TODO support utility conversions for other number types
func ParseNumber(f float64) *types.TypedValue {
	w := &wrappers.DoubleValue{Value: f}
	bs, err := proto.Marshal(w)
	if err != nil {
		panic(err)
	}
	return &types.TypedValue{
		Type:  string(TypeNumber),
		Value: bs,
	}
}

func FormatNumber(v *types.TypedValue) (float64, error) {
	if ValueType(v.Type) != TypeNumber {
		return 0, TypedValueErr{
			src: v,
			err: ErrUnsupportedType,
		}
	}
	if v.Value == nil {
		return 0, TypedValueErr{
			src: v,
			err: ErrValueConversion,
		}
	}
	w := &wrappers.DoubleValue{}
	err := proto.Unmarshal(v.Value, w)
	if err != nil {
		return 0, TypedValueErr{
			src: v,
			err: err,
		}
	}

	return w.Value, nil
}

type NilParserFormatter struct {
}

func (fp *NilParserFormatter) Accepts() []ValueType {
	return []ValueType{
		TypeNil,
	}
}

func (fp *NilParserFormatter) Parse(ctx Parser, i interface{}) (*types.TypedValue, error) {
	if i != nil {
		return nil, TypedValueErr{
			src: i,
			err: ErrUnsupportedType,
		}
	}
	return ParseNil(), nil
}

func (fp *NilParserFormatter) Format(ctx Formatter, v *types.TypedValue) (interface{}, error) {
	return nil, FormatNil(v)
}

func ParseNil() *types.TypedValue {
	return &types.TypedValue{
		Type:  string(TypeNil),
		Value: nil,
	}
}

func FormatNil(v *types.TypedValue) error {
	if ValueType(v.Type) != TypeNil {
		return TypedValueErr{
			src: v,
			err: ErrValueConversion,
		}
	}
	return nil
}

type StringParserFormatter struct{}

func (pf *StringParserFormatter) Accepts() []ValueType {
	return []ValueType{
		TypeString,
	}
}

func (pf *StringParserFormatter) Parse(ctx Parser, i interface{}) (*types.TypedValue, error) {
	s, ok := i.(string)
	if !ok {
		return nil, TypedValueErr{
			src: i,
			err: ErrUnsupportedType,
		}
	}

	return ParseString(s), nil
}

func (pf *StringParserFormatter) Format(ctx Formatter, v *types.TypedValue) (interface{}, error) {
	return FormatString(v)
}

func ParseString(s string) *types.TypedValue {
	return &types.TypedValue{
		Type:  string(TypeString),
		Value: []byte(s),
	}
}

func FormatString(v *types.TypedValue) (string, error) {
	if ValueType(v.Type) != TypeString {
		return "", TypedValueErr{
			src: v,
			err: ErrUnsupportedType,
		}
	}
	if v.Value == nil {
		return "", TypedValueErr{
			src: v,
			err: ErrValueConversion,
		}
	}
	return string(v.Value), nil
}

type BytesParserFormatter struct{}

func (pf *BytesParserFormatter) Accepts() []ValueType {
	return []ValueType{
		TypeBytes,
	}
}

func (pf *BytesParserFormatter) Parse(ctx Parser, i interface{}) (*types.TypedValue, error) {
	bs, ok := i.([]byte)
	if !ok {
		return nil, TypedValueErr{
			src: i,
			err: ErrUnsupportedType,
		}
	}

	return ParseBytes(bs), nil
}

func (pf *BytesParserFormatter) Format(ctx Formatter, v *types.TypedValue) (interface{}, error) {
	return FormatBytes(v)
}

func ParseBytes(bs []byte) *types.TypedValue {
	return &types.TypedValue{
		Type:  string(TypeBytes),
		Value: bs,
	}
}

func FormatBytes(v *types.TypedValue) ([]byte, error) {
	if ValueType(v.Type) != TypeBytes {
		return nil, TypedValueErr{
			src: v,
			err: ErrUnsupportedType,
		}
	}
	return v.Value, nil
}

type IdentityParserFormatter struct{}

func (pf *IdentityParserFormatter) Accepts() []ValueType {
	return []ValueType{}
}

func (pf *IdentityParserFormatter) Parse(ctx Parser, i interface{}) (*types.TypedValue, error) {
	if tv, ok := i.(*types.TypedValue); ok {
		return tv, nil
	}
	return nil, TypedValueErr{
		src: i,
		err: ErrUnsupportedType,
	}
}

func (pf *IdentityParserFormatter) Format(ctx Formatter, v *types.TypedValue) (interface{}, error) {
	return v, nil
}
