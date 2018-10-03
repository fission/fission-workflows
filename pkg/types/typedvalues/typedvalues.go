// package typedvalues provides a data container for annotating, interpreting, and transferring arbitrary data.
//
// It revolves around the TypedValue struct type. Users typically serialize generic (though not entirely generic yet)
// Golang data to a TypedValue in order to serialize and transfer it. Users can set and get annotations from the
// TypedValue, which is for example used to preserve important headers from a HTTP request from which the entity
// was parsed.
//
// The package relies heavily on Protobuf. Besides the primitive types, it supports any entity implementing the
// proto.Message interface. Internally the TypedValue uses the Protobuf Golang implementation for serializing and
// deserializing the TypedValue.
//
// In Workflows TypedValues are used for: serialization, allowing it to store the data in the same format as the
// workflow structures; storing metadata of task outputs, by annotating the TypedValues; and data evaluation,
// by parsing and formatting task inputs and outputs into structured data (where possible).
package typedvalues

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	"github.com/golang/protobuf/ptypes/wrappers"
	"github.com/pkg/errors"
)

const (
	TypeUrlPrefix = "types.fission.io/"
)

var (
	expressionRe            = regexp.MustCompile("^\\{(.*)\\}$")
	ErrIllegalTypeAssertion = errors.New("illegal type assertion")
	ErrUnsupportedType      = errors.New("unsupported type")
)

// TODO add caching to formatting and parsing

func (m *TypedValue) ValueType() string {
	if m == nil || m.Value == nil {
		return ""
	}
	typeUrl := m.Value.TypeUrl
	pos := strings.Index(typeUrl, TypeUrlPrefix)
	if pos == 0 {
		return typeUrl[len(TypeUrlPrefix):]
	}
	return typeUrl
}

func (m *TypedValue) Interface() interface{} {
	if m == nil || m.Value == nil {
		return nil
	}
	return MustUnwrap(m)
}

func (m *TypedValue) Float64() float64 {
	val := m.Interface()
	switch t := val.(type) {
	case int32:
		return float64(t)
	case int64:
		return float64(t)
	case uint32:
		return float64(t)
	case uint64:
		return float64(t)
	case float32:
		return float64(t)
	case float64:
		return float64(t)
	default:
		panic("TypedValue: not a number")
	}
}

// Short prints a short description of the Value
func (m *TypedValue) Short() string {
	if m == nil {
		return ""
	}
	return fmt.Sprintf("<Type=\"%s\", Val=\"%v\">", m.ValueType(), len(m.GetValue().GetValue()))
}

func (m *TypedValue) SetMetadata(k string, v string) *TypedValue {
	if m == nil {
		return m
	}
	if m.Metadata == nil {
		m.Metadata = map[string]string{}
	}
	m.Metadata[k] = v

	return m
}

func (m *TypedValue) GetMetadataValue(k string) (string, bool) {
	if m == nil {
		return "", false
	}

	if m.Metadata == nil {
		m.Metadata = map[string]string{}
	}
	v, ok := m.Metadata[k]
	return v, ok
}

func (m *TypedValue) Equals(other *TypedValue) bool {
	return proto.Equal(m, other)
}

func Unwrap(tv *TypedValue) (interface{}, error) {
	if tv == nil {
		return nil, nil
	}

	msg, err := UnwrapProto(tv)
	if err != nil {
		return nil, err
	}

	var i interface{}
	switch t := msg.(type) {
	case *MapValue:
		mapValue := make(map[string]interface{}, len(t.Value))
		for k, v := range t.Value {
			entry, err := Unwrap(v)
			if err != nil {
				return TypedValue{}, errors.Wrapf(err, "failed to format map[%s]", k)
			}
			mapValue[k] = entry
		}
		i = mapValue
	case *ArrayValue:
		arrayValue := make([]interface{}, len(t.Value))
		for k, v := range t.Value {
			entry, err := Unwrap(v)
			if err != nil {
				return TypedValue{}, errors.Wrapf(err, "failed to format array[%d]", k)
			}
			arrayValue[k] = entry
		}
		i = arrayValue
	case *Expression:
		i = t.Value
	case *wrappers.BoolValue:
		i = t.Value
	case *wrappers.FloatValue:
		i = t.Value
	case *wrappers.DoubleValue:
		i = t.Value
	case *wrappers.Int32Value:
		i = t.Value
	case *wrappers.Int64Value:
		i = t.Value
	case *wrappers.UInt64Value:
		i = t.Value
	case *wrappers.UInt32Value:
		i = t.Value
	case *wrappers.StringValue:
		i = t.Value
	case *wrappers.BytesValue:
		i = t.Value
	case *NilValue:
		i = nil
	default:
		// Message does not have to be unwrapped(?)
		i = t
	}

	return i, nil
}

func Wrap(val interface{}) (*TypedValue, error) {
	var msg proto.Message
	switch t := val.(type) {
	case *TypedValue:
		return t, nil
	case []*TypedValue:
		msg = &ArrayValue{Value: t}
	case map[string]*TypedValue:
		msg = &MapValue{Value: t}
	case map[string]interface{}:
		values := make(map[string]*TypedValue, len(t))
		for k, v := range t {
			tv, err := Wrap(v)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse map[%s]", k)
			}
			values[k] = tv
		}
		msg = &MapValue{Value: values}
	case []interface{}:
		values := make([]*TypedValue, len(t))
		for i, v := range t {
			tv, err := Wrap(v)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse array[%d]", i)
			}
			values[i] = tv
		}
		msg = &ArrayValue{Value: values}
	case proto.Message:
		msg = t
	case string:
		if expressionRe.MatchString(t) {
			msg = &Expression{Value: t}
		} else {
			msg = &wrappers.StringValue{Value: t}
		}
	case bool:
		msg = &wrappers.BoolValue{Value: t}
	case float32:
		msg = &wrappers.FloatValue{Value: t}
	case float64:
		msg = &wrappers.DoubleValue{Value: t}
	case int64:
		msg = &wrappers.Int64Value{Value: t}
	case int:
		msg = &wrappers.Int32Value{Value: int32(t)}
	case int32:
		msg = &wrappers.Int32Value{Value: t}
	case uint32:
		msg = &wrappers.UInt32Value{Value: t}
	case uint64:
		msg = &wrappers.UInt64Value{Value: t}
	case []byte:
		msg = &wrappers.BytesValue{Value: t}
	case nil:
		msg = &NilValue{}
	default:
		return nil, errors.Wrapf(ErrUnsupportedType, "parse %T", t)
	}
	marshaled, err := marshalAny(msg)
	if err != nil {
		return nil, err
	}
	// TODO cache already if safe (copied/cloned/primitive)
	return &TypedValue{
		Value: marshaled,
	}, nil
}

func MustWrap(val interface{}) *TypedValue {
	tv, err := Wrap(val)
	if err != nil {
		panic(err)
	}
	return tv
}

func MustUnwrap(tv *TypedValue) interface{} {
	i, err := Unwrap(tv)
	if err != nil {
		panic(err)
	}
	return i
}

func UnwrapString(tv *TypedValue) (string, error) {
	i, err := Unwrap(tv)
	if err != nil {
		return "", err
	}

	if s, ok := i.(string); ok {
		return s, nil
	}

	if bs, ok := i.([]byte); ok {
		return string(bs), nil
	}

	return "", errors.Wrapf(ErrIllegalTypeAssertion, "failed to unwrap %s to string", tv.ValueType())
}

func UnwrapProto(tv *TypedValue) (proto.Message, error) {
	var dynamic ptypes.DynamicAny
	err := ptypes.UnmarshalAny(tv.Value, &dynamic)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return dynamic.Message, nil
}

func UnwrapBytes(tv *TypedValue) ([]byte, error) {
	i, err := Unwrap(tv)
	if err != nil {
		return nil, err
	}

	if d, ok := i.([]byte); ok {
		return d, nil
	}
	if s, ok := i.(string); ok {
		return []byte(s), nil
	}

	return nil, errors.Wrapf(ErrIllegalTypeAssertion, "failed to unwrap %s to bytes", tv.ValueType())
}

func UnwrapBool(tv *TypedValue) (bool, error) {
	i, err := Unwrap(tv)
	if err != nil {
		return false, err
	}

	s, ok := i.(bool)
	if !ok {
		return false, errors.Wrapf(ErrIllegalTypeAssertion, "failed to unwrap %s to bool", tv.ValueType())
	}
	return s, nil
}

func UnwrapArray(tv *TypedValue) ([]interface{}, error) {
	i, err := Unwrap(tv)
	if err != nil {
		return nil, err
	}

	s, ok := i.([]interface{})
	if !ok {
		return nil, errors.Wrapf(ErrIllegalTypeAssertion, "failed to unwrap %s to array", tv.ValueType())
	}
	return s, nil
}

func UnwrapTypedValueArray(tv *TypedValue) ([]*TypedValue, error) {
	arrayWrapper := &ArrayValue{}
	err := ptypes.UnmarshalAny(tv.Value, arrayWrapper)
	if err != nil {
		return nil, errors.Wrapf(ErrIllegalTypeAssertion, "failed to unwrap %s to TypedValue-array", tv.ValueType())
	}

	arrayValue := make([]*TypedValue, len(arrayWrapper.Value))
	for k, v := range arrayWrapper.Value {
		arrayValue[k] = v
	}
	return arrayValue, nil
}

func UnwrapTypedValueMap(tv *TypedValue) (map[string]*TypedValue, error) {
	mapWrapper := &MapValue{}
	err := ptypes.UnmarshalAny(tv.Value, mapWrapper)
	if err != nil {
		return nil, errors.Wrapf(ErrIllegalTypeAssertion, "failed to unwrap %s to TypedValue-map", tv.ValueType())
	}

	mapValue := make(map[string]*TypedValue, len(mapWrapper.Value))
	for k, v := range mapWrapper.Value {
		mapValue[k] = v
	}
	return mapValue, nil
}

func UnwrapMapTypedValue(tvs map[string]*TypedValue) (map[string]interface{}, error) {
	tv, err := Wrap(tvs)
	if err != nil {
		return nil, err
	}
	return UnwrapMap(tv)
}

func WrapMapTypedValue(tvs map[string]interface{}) (map[string]*TypedValue, error) {
	entries := make(map[string]*TypedValue, len(tvs))
	for k, v := range tvs {
		e, err := Wrap(v)
		if err != nil {
			return nil, err
		}
		entries[k] = e
	}
	return entries, nil
}

func MustWrapMapTypedValue(tvs map[string]interface{}) map[string]*TypedValue {
	tv, err := WrapMapTypedValue(tvs)
	if err != nil {
		panic(err)
	}
	return tv
}

func UnwrapMap(tv *TypedValue) (map[string]interface{}, error) {
	i, err := Unwrap(tv)
	if err != nil {
		return nil, err
	}

	s, ok := i.(map[string]interface{})
	if !ok {
		return nil, errors.Wrapf(ErrIllegalTypeAssertion, "failed to unwrap %s to map", tv.ValueType())
	}
	return s, nil
}

// TODO reduce verbosity of these numberic implementations
func UnwrapInt64(tv *TypedValue) (int64, error) {
	i, err := Unwrap(tv)
	if err != nil {
		return 0, err
	}

	switch t := i.(type) {
	case int64:
		return int64(t), nil
	case int32:
		return int64(t), nil
	case uint32:
		return int64(t), nil
	case uint64:
		return int64(t), nil
	case float32:
		return int64(t), nil
	case float64:
		return int64(t), nil
	default:
		return 0, errors.Wrapf(ErrIllegalTypeAssertion, "failed to unwrap %s to int64", tv.ValueType())
	}
}

func UnwrapFloat64(tv *TypedValue) (float64, error) {
	i, err := Unwrap(tv)
	if err != nil {
		return 0, err
	}

	switch t := i.(type) {
	case int64:
		return float64(t), nil
	case int32:
		return float64(t), nil
	case uint32:
		return float64(t), nil
	case uint64:
		return float64(t), nil
	case float32:
		return float64(t), nil
	case float64:
		return float64(t), nil
	default:
		return 0, errors.Wrapf(ErrIllegalTypeAssertion, "failed to unwrap %s to float64", tv.ValueType())
	}
}

func UnwrapExpression(tv *TypedValue) (string, error) {
	s, err := UnwrapString(tv)
	if err != nil {
		return "", err
	}

	if !IsExpression(s) {
		return "", errors.Wrapf(ErrIllegalTypeAssertion, "failed to unwrap %s to expression", tv.ValueType())
	}
	return s, nil
}

func RemoveExpressionDelimiters(expr string) string {
	return expressionRe.ReplaceAllString(expr, "$1")
}

func IsExpression(s string) bool {
	return expressionRe.MatchString(s)
}

// marshalAny takes the protocol buffer and encodes it into google.protobuf.Any, without prepending the Google API URL.
func marshalAny(pb proto.Message) (*any.Any, error) {
	value, err := proto.Marshal(pb)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &any.Any{TypeUrl: TypeUrlPrefix + proto.MessageName(pb), Value: value}, nil
}
