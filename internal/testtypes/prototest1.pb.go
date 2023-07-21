// Code generated by protoc-gen-gogo and then hand-edited to remove
// explicit proto dependencies so that the avro module doesn't need to depend
// on that package.

// source: prototest1.proto

package testtypes

import (
	"gopkg.in/errgo.v2/fmt/errors"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
// var _ = proto.Marshal
var _ = errors.Newf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
//const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// LabelFor also has a doc comment.
type LabelFor int32

const (
	LabelFor_LABEL_FOR_ZERO  LabelFor = 0
	LabelFor_LABEL_FOR_ONE   LabelFor = 1
	LabelFor_LABEL_FOR_TWO   LabelFor = 2
	LabelFor_LABEL_FOR_THREE LabelFor = 3
)

var LabelFor_name = map[int32]string{
	0: "LABEL_FOR_ZERO",
	1: "LABEL_FOR_ONE",
	2: "LABEL_FOR_TWO",
	3: "LABEL_FOR_THREE",
}

var LabelFor_value = map[string]int32{
	"LABEL_FOR_ZERO":  0,
	"LABEL_FOR_ONE":   1,
	"LABEL_FOR_TWO":   2,
	"LABEL_FOR_THREE": 3,
}

func (x LabelFor) String() string {
	return ProtoEnumName(LabelFor_name, int32(x))
}

func (LabelFor) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_8f1e2929fcb7cefa, []int{0}
}

// MessageA has a doc comment.
type MessageA struct {
	Id                   string   `protobuf:"bytes,1,opt,name=id,proto3" json:"id,omitempty"`
	Label                LabelFor `protobuf:"varint,2,opt,name=label,proto3,enum=arble.foo.v1.LabelFor" json:"label,omitempty"`
	FooUrl               string   `protobuf:"bytes,3,opt,name=foo_url,json=fooUrl,proto3" json:"foo_url,omitempty"`
	Enabled              bool     `protobuf:"varint,4,opt,name=enabled,proto3" json:"enabled,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MessageA) Reset() { *m = MessageA{} }

// This is commented out because otherwise the logging in tests
// calls it, which we don't want.
//
//	func (m *MessageA) String() string {
//		return proto.CompactTextString(m)
//	}
func (*MessageA) ProtoMessage() {}
func (*MessageA) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f1e2929fcb7cefa, []int{0}
}
func (m *MessageA) XXX_Unmarshal(b []byte) error {
	panic("not called")
	//	return xxx_messageInfo_MessageA.Unmarshal(m, b)
}
func (m *MessageA) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	panic("not called")
	//	return xxx_messageInfo_MessageA.Marshal(b, m, deterministic)
}
func (m *MessageA) XXX_Merge(src ProtoMessage) {
	panic("not called")
	//	xxx_messageInfo_MessageA.Merge(m, src)
}
func (m *MessageA) XXX_Size() int {
	panic("not called")
	//	return xxx_messageInfo_MessageA.Size(m)
}
func (m *MessageA) XXX_DiscardUnknown() {
	//	xxx_messageInfo_MessageA.DiscardUnknown(m)
}

//var xxx_messageInfo_MessageA proto.InternalMessageInfo

func (m *MessageA) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *MessageA) GetLabel() LabelFor {
	if m != nil {
		return m.Label
	}
	return LabelFor_LABEL_FOR_ZERO
}

func (m *MessageA) GetFooUrl() string {
	if m != nil {
		return m.FooUrl
	}
	return ""
}

func (m *MessageA) GetEnabled() bool {
	if m != nil {
		return m.Enabled
	}
	return false
}

// UserDisability is used to show whether a disability has been selected by a user.
type MessageB struct {
	Arble                *MessageA `protobuf:"bytes,1,opt,name=arble,proto3" json:"arble,omitempty"`
	Selected             bool      `protobuf:"varint,2,opt,name=selected,proto3" json:"selected,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *MessageB) Reset() { *m = MessageB{} }

// This is commented out because otherwise the logging in tests
// calls it, which we don't want.
//
//	func (m *MessageB) String() string {
//		return proto.CompactTextString(m)
//	}
func (*MessageB) ProtoMessage() {}
func (*MessageB) Descriptor() ([]byte, []int) {
	return fileDescriptor_8f1e2929fcb7cefa, []int{1}
}
func (m *MessageB) XXX_Unmarshal(b []byte) error {
	panic("not called")
}
func (m *MessageB) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	panic("not called")
}
func (m *MessageB) XXX_Merge(src ProtoMessage) {
	panic("not called")
}
func (m *MessageB) XXX_Size() int {
	panic("not called")
}
func (m *MessageB) XXX_DiscardUnknown() {
	panic("not called")
}

//var xxx_messageInfo_MessageB proto.InternalMessageInfo

func (m *MessageB) GetArble() *MessageA {
	if m != nil {
		return m.Arble
	}
	return nil
}

func (m *MessageB) GetSelected() bool {
	if m != nil {
		return m.Selected
	}
	return false
}

//func init() {
//	proto.RegisterEnum("arble.foo.v1.LabelFor", LabelFor_name, LabelFor_value)
//	proto.RegisterType((*MessageA)(nil), "arble.foo.v1.MessageA")
//	proto.RegisterType((*MessageB)(nil), "arble.foo.v1.MessageB")
//}

//func init() { proto.RegisterFile("prototest1.proto", fileDescriptor_8f1e2929fcb7cefa) }

var fileDescriptor_8f1e2929fcb7cefa = []byte{
	// 259 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x5c, 0x90, 0x51, 0x4b, 0xc3, 0x30,
	0x14, 0x85, 0x6d, 0xa7, 0x5b, 0x77, 0xa7, 0xb5, 0x5e, 0x41, 0x8b, 0x4f, 0x65, 0x4f, 0x45, 0x24,
	0x30, 0xfd, 0x05, 0x2b, 0x64, 0xf8, 0x50, 0x2d, 0x84, 0x89, 0xb8, 0x97, 0xd2, 0xda, 0x5b, 0x19,
	0x04, 0x33, 0x92, 0x28, 0x88, 0x7f, 0x5e, 0x9a, 0xd1, 0x39, 0x7d, 0xcb, 0x3d, 0xb9, 0xf7, 0x7c,
	0x87, 0x03, 0xd1, 0x46, 0x2b, 0xab, 0x2c, 0x19, 0x3b, 0x63, 0xee, 0x89, 0xc7, 0x95, 0xae, 0x25,
	0xb1, 0x56, 0x29, 0xf6, 0x39, 0x9b, 0x7e, 0x43, 0xf0, 0x40, 0xc6, 0x54, 0x6f, 0x34, 0xc7, 0x10,
	0xfc, 0x75, 0x13, 0x7b, 0x89, 0x97, 0x8e, 0x85, 0xbf, 0x6e, 0xf0, 0x06, 0x8e, 0x64, 0x55, 0x93,
	0x8c, 0xfd, 0xc4, 0x4b, 0xc3, 0xdb, 0x0b, 0xb6, 0x7f, 0xc9, 0xf2, 0xee, 0x6b, 0xa1, 0xb4, 0xd8,
	0x2e, 0xe1, 0x25, 0x8c, 0x5a, 0xa5, 0xca, 0x0f, 0x2d, 0xe3, 0x81, 0xb3, 0x18, 0xb6, 0x4a, 0x3d,
	0x69, 0x89, 0x31, 0x8c, 0xe8, 0xbd, 0xaa, 0x25, 0x35, 0xf1, 0x61, 0xe2, 0xa5, 0x81, 0xe8, 0xc7,
	0xe9, 0x72, 0x07, 0xcf, 0x3a, 0x98, 0xb3, 0x77, 0xfc, 0xc9, 0x7f, 0x58, 0x9f, 0x51, 0x6c, 0x97,
	0xf0, 0x0a, 0x02, 0x43, 0x92, 0x5e, 0x2d, 0x35, 0x2e, 0x5d, 0x20, 0x76, 0xf3, 0xf5, 0x0b, 0x04,
	0x7d, 0x36, 0x44, 0x08, 0xf3, 0x79, 0xc6, 0xf3, 0x72, 0x51, 0x88, 0x72, 0xc5, 0x45, 0x11, 0x1d,
	0xe0, 0x19, 0x9c, 0xfc, 0x6a, 0xc5, 0x23, 0x8f, 0xbc, 0xbf, 0xd2, 0xf2, 0xb9, 0x88, 0x7c, 0x3c,
	0x87, 0xd3, 0x3d, 0xe9, 0x5e, 0x70, 0x1e, 0x0d, 0xb2, 0xc9, 0x6a, 0xdc, 0x55, 0x69, 0xbf, 0x36,
	0x64, 0xea, 0xa1, 0xeb, 0xf3, 0xee, 0x27, 0x00, 0x00, 0xff, 0xff, 0x28, 0xf9, 0x9e, 0xf2, 0x63,
	0x01, 0x00, 0x00,
}
