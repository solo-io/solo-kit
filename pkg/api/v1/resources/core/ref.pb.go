// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: ref.proto

package core

import (
	fmt "fmt"
	math "math"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/solo-io/protoc-gen-ext/extproto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// A way to reference resources across namespaces
// TODO(ilackarms): make upstreamname and secretref into ResourceRefs
type ResourceRef struct {
	Name      string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	Namespace string `protobuf:"bytes,2,opt,name=namespace,proto3" json:"namespace,omitempty"`
}

func (m *ResourceRef) Reset()         { *m = ResourceRef{} }
func (m *ResourceRef) String() string { return proto.CompactTextString(m) }
func (*ResourceRef) ProtoMessage()    {}
func (*ResourceRef) Descriptor() ([]byte, []int) {
	return fileDescriptor_65d958559ea81b29, []int{0}
}
func (m *ResourceRef) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ResourceRef.Unmarshal(m, b)
}
func (m *ResourceRef) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ResourceRef.Marshal(b, m, deterministic)
}
func (m *ResourceRef) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ResourceRef.Merge(m, src)
}
func (m *ResourceRef) XXX_Size() int {
	return xxx_messageInfo_ResourceRef.Size(m)
}
func (m *ResourceRef) XXX_DiscardUnknown() {
	xxx_messageInfo_ResourceRef.DiscardUnknown(m)
}

var xxx_messageInfo_ResourceRef proto.InternalMessageInfo

func (m *ResourceRef) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ResourceRef) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
}

func init() {
	proto.RegisterType((*ResourceRef)(nil), "core.solo.io.ResourceRef")
}

func init() { proto.RegisterFile("ref.proto", fileDescriptor_65d958559ea81b29) }

var fileDescriptor_65d958559ea81b29 = []byte{
	// 193 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xe2, 0x2c, 0x4a, 0x4d, 0xd3,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x49, 0xce, 0x2f, 0x4a, 0xd5, 0x2b, 0xce, 0xcf, 0xc9,
	0xd7, 0xcb, 0xcc, 0x97, 0x12, 0x49, 0xcf, 0x4f, 0xcf, 0x07, 0x4b, 0xe8, 0x83, 0x58, 0x10, 0x35,
	0x52, 0x42, 0xa9, 0x15, 0x25, 0x10, 0xc1, 0xd4, 0x8a, 0x12, 0x88, 0x98, 0x92, 0x2f, 0x17, 0x77,
	0x50, 0x6a, 0x71, 0x7e, 0x69, 0x51, 0x72, 0x6a, 0x50, 0x6a, 0x9a, 0x90, 0x10, 0x17, 0x4b, 0x5e,
	0x62, 0x6e, 0xaa, 0x04, 0xa3, 0x02, 0xa3, 0x06, 0x67, 0x10, 0x98, 0x2d, 0x24, 0xc3, 0xc5, 0x09,
	0xa2, 0x8b, 0x0b, 0x12, 0x93, 0x53, 0x25, 0x98, 0xc0, 0x12, 0x08, 0x01, 0x2b, 0x9e, 0x0b, 0x0b,
	0xe5, 0x19, 0x26, 0x2c, 0x92, 0x67, 0x98, 0xb1, 0x48, 0x9e, 0xc1, 0xc9, 0xee, 0x87, 0x13, 0xe3,
	0x8a, 0x47, 0x72, 0x8c, 0x51, 0xa6, 0xe9, 0x99, 0x25, 0x19, 0xa5, 0x49, 0x7a, 0xc9, 0xf9, 0xb9,
	0xfa, 0x20, 0x57, 0xe9, 0x66, 0xe6, 0x43, 0xe8, 0xec, 0xcc, 0x12, 0xfd, 0x82, 0xec, 0x74, 0xfd,
	0xc4, 0x82, 0x4c, 0xfd, 0x32, 0x43, 0xfd, 0x22, 0xa8, 0xe5, 0xc5, 0xfa, 0x20, 0x0f, 0x24, 0xb1,
	0x81, 0x5d, 0x65, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0xe4, 0x82, 0xa5, 0x43, 0xda, 0x00, 0x00,
	0x00,
}

func (this *ResourceRef) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ResourceRef)
	if !ok {
		that2, ok := that.(ResourceRef)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.Name != that1.Name {
		return false
	}
	if this.Namespace != that1.Namespace {
		return false
	}
	return true
}
