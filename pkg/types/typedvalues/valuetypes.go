package typedvalues

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes/wrappers"
)

var (
	TypeBool       string
	TypeInt32      string
	TypeInt64      string
	TypeUInt32     string
	TypeUInt64     string
	TypeFloat32    string
	TypeFloat64    string
	TypeString     string
	TypeBytes      string
	TypeNil        string
	TypeExpression string
	TypeMap        string
	TypeList       string
	TypeNumber     []string
	Types          []string
)

// Note: ensure that this file is lexically after the generated Protobufs because of package initialization order.
func init() {
	TypeBool = proto.MessageName(&wrappers.BoolValue{})
	TypeInt32 = proto.MessageName(&wrappers.Int32Value{})
	TypeInt64 = proto.MessageName(&wrappers.Int64Value{})
	TypeUInt32 = proto.MessageName(&wrappers.UInt32Value{})
	TypeUInt64 = proto.MessageName(&wrappers.UInt64Value{})
	TypeFloat32 = proto.MessageName(&wrappers.FloatValue{})
	TypeFloat64 = proto.MessageName(&wrappers.DoubleValue{})
	TypeString = proto.MessageName(&wrappers.StringValue{})
	TypeBytes = proto.MessageName(&wrappers.BytesValue{})
	TypeNil = proto.MessageName(&NilValue{})
	TypeExpression = proto.MessageName(&Expression{})
	TypeMap = proto.MessageName(&MapValue{})
	TypeList = proto.MessageName(&ArrayValue{})
	TypeNumber = []string{
		TypeFloat64,
		TypeFloat32,
		TypeInt32,
		TypeInt64,
		TypeUInt32,
		TypeUInt64,
	}
	Types = []string{
		TypeBool,
		TypeInt32,
		TypeInt64,
		TypeUInt32,
		TypeUInt64,
		TypeFloat32,
		TypeFloat64,
		TypeString,
		TypeBytes,
		TypeNil,
		TypeExpression,
		TypeMap,
		TypeList,
	}
}
