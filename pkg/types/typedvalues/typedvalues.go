package typedvalues

import (
	"errors"
	"fmt"
	"regexp"

	"github.com/gogo/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/wrappers"
)

var (
	TypeBool       = "bool"
	TypeNumber     = "number"
	TypeNil        = "nil"
	TypeString     = "string"
	TypeBytes      = "bytes"
	TypeExpression = "expression"
	TypeTask       = "task"
	TypeWorkflow   = "workflow"
	TypeMap        = "map"
	TypeList       = "list"
)

var (
	expressionRe = regexp.MustCompile("^\\{(.*)\\}$")
)

// TODO add caching to formatting and parsing

func (m *TypedValue) ValueType() string {
	return m.Value.TypeUrl
}

func (m *TypedValue) Interface() interface{} {
	return MustFormat(m)
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

//
// TypedValue
//

// Prints a short description of the Value
func (tv TypedValue) Short() string {
	// var val string
	// if len(tv.Value) > typedValueShortMaxLen {
	// 	val = fmt.Sprintf("%s[..%d..]", tv.Value[:typedValueShortMaxLen], len(tv.Value)-typedValueShortMaxLen)
	// } else {
	// 	val = fmt.Sprintf("%s", tv.Value)
	// }
	//
	// return fmt.Sprintf("<Type=\"%s\", Val=\"%v\">", tv.Type, strings.Replace(val, "\n", "", -1))
	return ""
}

func (tv *TypedValue) SetLabel(k string, v string) *TypedValue {
	// if tv == nil {
	// 	return tv
	// }
	// if tv.Labels == nil {
	// 	tv.Labels = map[string]string{}
	// }
	// tv.Labels[k] = v
	//
	return tv
}

func (tv *TypedValue) GetLabel(k string) (string, bool) {
	// if tv == nil {
	// 	return "", false
	// }
	//
	// if tv.Labels == nil {
	// 	tv.Labels = map[string]string{}
	// }
	// v, ok := tv.Labels[k]

	return "", true
}

func Format(tv *TypedValue) (interface{}, error) {
	var dynamic ptypes.DynamicAny
	err := ptypes.UnmarshalAny(tv.Value, &dynamic)
	if err != nil {
		return nil, err
	}

	var i interface{}
	switch t := dynamic.Message.(type) {
	case *MapValue:
		mapValue := make(map[string]interface{}, len(t.Value))
		for k, v := range t.Value {
			entry, err := Format(v)
			if err != nil {
				return TypedValue{}, fmt.Errorf("map[%s]: %v", k, err)
			}
			mapValue[k] = entry
		}
		i = mapValue
	case *ArrayValue:
		arrayValue := make([]interface{}, len(t.Value))
		for k, v := range t.Value {
			entry, err := Format(v)
			if err != nil {
				return TypedValue{}, fmt.Errorf("[%d]: %v", k, err)
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
		// Message does not have to be unwrapped
		i = t
	}

	return i, nil
}

func Parse(val interface{}) (*TypedValue, error) {
	var msg proto.Message
	switch t := val.(type) {
	case map[string]interface{}:
		values := make(map[string]*TypedValue, len(t))
		for k, v := range t {
			tv, err := Parse(v)
			if err != nil {
				return nil, fmt.Errorf("map[%s]: %v", k, err)
			}
			values[k] = tv
		}
		msg = &MapValue{Value: values}
	case []interface{}:
		values := make([]*TypedValue, len(t))
		for i, v := range t {
			tv, err := Parse(v)
			if err != nil {
				return nil, fmt.Errorf("array[%d]: %v", i, err)
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
		return nil, fmt.Errorf("cannot parse unsupported %T", t)
	}
	marshaled, err := ptypes.MarshalAny(msg)
	if err != nil {
		return nil, err
	}
	// TODO cache already if safe (copied/cloned/primitive)
	return &TypedValue{
		Value: marshaled,
	}, nil
}

func MustParse(val interface{}) *TypedValue {
	tv, err := Parse(val)
	if err != nil {
		panic(err)
	}
	return tv
}

func MustFormat(tv *TypedValue) interface{} {
	i, err := Format(tv)
	if err != nil {
		panic(err)
	}
	return i
}

func FormatString(tv *TypedValue) (string, error) {
	i, err := Format(tv)
	if err != nil {
		return "", err
	}

	s, ok := i.(string)
	if !ok {
		return "", errors.New("not string")
	}
	return s, nil
}

func FormatBool(tv *TypedValue) (bool, error) {
	i, err := Format(tv)
	if err != nil {
		return false, err
	}

	s, ok := i.(bool)
	if !ok {
		return false, errors.New("not string")
	}
	return s, nil
}

func FormatNumber(tv *TypedValue) (float64, error) {
	panic("todo")
}
