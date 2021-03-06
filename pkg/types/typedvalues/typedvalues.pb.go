// Code generated by protoc-gen-go. DO NOT EDIT.
// source: pkg/types/typedvalues/typedvalues.proto

/*
Package typedvalues is a generated protocol buffer package.

It is generated from these files:
	pkg/types/typedvalues/typedvalues.proto

It has these top-level messages:
	TypedValue
	Expression
	MapValue
	ArrayValue
	NilValue
*/
package typedvalues

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import google_protobuf "github.com/golang/protobuf/ptypes/any"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// TypedValue is used to serialize, deserialize, transfer data values across the workflow engine.
//
// It consists partly copy of protobuf's Any, to avoid protobuf requirement of a protobuf-based type.
type TypedValue struct {
	// Value holds the actual value in a serialized form.
	Value *google_protobuf.Any `protobuf:"bytes,1,opt,name=value" json:"value,omitempty"`
	// Labels hold metadata about the value. It is used for example to store origins of data, past transformations,
	// and information needed by serialization processes.
	Metadata map[string]string `protobuf:"bytes,3,rep,name=metadata" json:"metadata,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *TypedValue) Reset()                    { *m = TypedValue{} }
func (m *TypedValue) String() string            { return proto.CompactTextString(m) }
func (*TypedValue) ProtoMessage()               {}
func (*TypedValue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *TypedValue) GetValue() *google_protobuf.Any {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *TypedValue) GetMetadata() map[string]string {
	if m != nil {
		return m.Metadata
	}
	return nil
}

type Expression struct {
	Value string `protobuf:"bytes,1,opt,name=value" json:"value,omitempty"`
}

func (m *Expression) Reset()                    { *m = Expression{} }
func (m *Expression) String() string            { return proto.CompactTextString(m) }
func (*Expression) ProtoMessage()               {}
func (*Expression) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Expression) GetValue() string {
	if m != nil {
		return m.Value
	}
	return ""
}

type MapValue struct {
	Value map[string]*TypedValue `protobuf:"bytes,1,rep,name=value" json:"value,omitempty" protobuf_key:"bytes,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *MapValue) Reset()                    { *m = MapValue{} }
func (m *MapValue) String() string            { return proto.CompactTextString(m) }
func (*MapValue) ProtoMessage()               {}
func (*MapValue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *MapValue) GetValue() map[string]*TypedValue {
	if m != nil {
		return m.Value
	}
	return nil
}

type ArrayValue struct {
	Value []*TypedValue `protobuf:"bytes,1,rep,name=value" json:"value,omitempty"`
}

func (m *ArrayValue) Reset()                    { *m = ArrayValue{} }
func (m *ArrayValue) String() string            { return proto.CompactTextString(m) }
func (*ArrayValue) ProtoMessage()               {}
func (*ArrayValue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *ArrayValue) GetValue() []*TypedValue {
	if m != nil {
		return m.Value
	}
	return nil
}

type NilValue struct {
}

func (m *NilValue) Reset()                    { *m = NilValue{} }
func (m *NilValue) String() string            { return proto.CompactTextString(m) }
func (*NilValue) ProtoMessage()               {}
func (*NilValue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func init() {
	proto.RegisterType((*TypedValue)(nil), "fission.workflows.types.TypedValue")
	proto.RegisterType((*Expression)(nil), "fission.workflows.types.Expression")
	proto.RegisterType((*MapValue)(nil), "fission.workflows.types.MapValue")
	proto.RegisterType((*ArrayValue)(nil), "fission.workflows.types.ArrayValue")
	proto.RegisterType((*NilValue)(nil), "fission.workflows.types.NilValue")
}

func init() { proto.RegisterFile("pkg/types/typedvalues/typedvalues.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 302 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0x52, 0x2f, 0xc8, 0x4e, 0xd7,
	0x2f, 0xa9, 0x2c, 0x48, 0x2d, 0x06, 0x93, 0x29, 0x65, 0x89, 0x39, 0xa5, 0xa8, 0x6c, 0xbd, 0x82,
	0xa2, 0xfc, 0x92, 0x7c, 0x21, 0xf1, 0xb4, 0xcc, 0xe2, 0xe2, 0xcc, 0xfc, 0x3c, 0xbd, 0xf2, 0xfc,
	0xa2, 0xec, 0xb4, 0x9c, 0xfc, 0xf2, 0x62, 0x3d, 0xb0, 0x36, 0x29, 0xc9, 0xf4, 0xfc, 0xfc, 0xf4,
	0x9c, 0x54, 0x7d, 0xb0, 0xb2, 0xa4, 0xd2, 0x34, 0xfd, 0xc4, 0xbc, 0x4a, 0x88, 0x1e, 0xa5, 0x23,
	0x8c, 0x5c, 0x5c, 0x21, 0x20, 0x93, 0xc2, 0x40, 0x26, 0x09, 0x69, 0x71, 0xb1, 0x82, 0x8d, 0x94,
	0x60, 0x54, 0x60, 0xd4, 0xe0, 0x36, 0x12, 0xd1, 0x83, 0xe8, 0xd4, 0x83, 0xe9, 0xd4, 0x73, 0xcc,
	0xab, 0x0c, 0x82, 0x28, 0x11, 0xf2, 0xe5, 0xe2, 0xc8, 0x4d, 0x2d, 0x49, 0x4c, 0x49, 0x2c, 0x49,
	0x94, 0x60, 0x56, 0x60, 0xd6, 0xe0, 0x36, 0x32, 0xd4, 0xc3, 0xe1, 0x02, 0x3d, 0x84, 0x15, 0x7a,
	0xbe, 0x50, 0x3d, 0xae, 0x79, 0x25, 0x45, 0x95, 0x41, 0x70, 0x23, 0xa4, 0xac, 0xb9, 0x78, 0x51,
	0xa4, 0x84, 0x04, 0xb8, 0x98, 0xb3, 0x53, 0x2b, 0xc1, 0x2e, 0xe1, 0x0c, 0x02, 0x31, 0x85, 0x44,
	0x60, 0xae, 0x63, 0x02, 0x8b, 0x41, 0x38, 0x56, 0x4c, 0x16, 0x8c, 0x4a, 0x4a, 0x5c, 0x5c, 0xae,
	0x15, 0x05, 0x45, 0xa9, 0x60, 0xdb, 0x11, 0xea, 0x18, 0x91, 0xd4, 0x29, 0xad, 0x65, 0xe4, 0xe2,
	0xf0, 0x4d, 0x2c, 0x80, 0x78, 0xd4, 0x09, 0xa1, 0x04, 0xe4, 0x72, 0x1d, 0x9c, 0x2e, 0x87, 0xe9,
	0xd0, 0x03, 0x93, 0x10, 0x47, 0x43, 0xb4, 0x4a, 0xc5, 0x72, 0x71, 0x21, 0x04, 0xb1, 0x38, 0xd7,
	0x12, 0xd9, 0xb9, 0xdc, 0x46, 0xca, 0x44, 0x84, 0x0e, 0xb2, 0x9f, 0xdc, 0xb9, 0xb8, 0x1c, 0x8b,
	0x8a, 0x12, 0x2b, 0x21, 0x0e, 0xb6, 0x44, 0x75, 0x30, 0x09, 0x86, 0x29, 0x71, 0x71, 0x71, 0xf8,
	0x65, 0xe6, 0x80, 0x85, 0x9c, 0x78, 0xa3, 0xb8, 0x91, 0x12, 0x4e, 0x12, 0x1b, 0x38, 0x62, 0x8d,
	0x01, 0x01, 0x00, 0x00, 0xff, 0xff, 0x80, 0xbc, 0x9e, 0x1f, 0x64, 0x02, 0x00, 0x00,
}
