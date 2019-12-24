// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: solo-kit/test/mocks/api/v1/more_mock_resources.proto

package v1

import (
	bytes "bytes"
	fmt "fmt"
	math "math"

	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/solo-io/protoc-gen-ext/extproto"
	core "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
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

//
//Description of the AnotherMockResource
type AnotherMockResource struct {
	Metadata core.Metadata `protobuf:"bytes,1,opt,name=metadata,proto3" json:"metadata"`
	Status   core.Status   `protobuf:"bytes,6,opt,name=status,proto3" json:"status"`
	// comments that go above the basic field in our docs
	BasicField           string   `protobuf:"bytes,2,opt,name=basic_field,json=basicField,proto3" json:"basic_field,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AnotherMockResource) Reset()         { *m = AnotherMockResource{} }
func (m *AnotherMockResource) String() string { return proto.CompactTextString(m) }
func (*AnotherMockResource) ProtoMessage()    {}
func (*AnotherMockResource) Descriptor() ([]byte, []int) {
	return fileDescriptor_166414021e79b557, []int{0}
}
func (m *AnotherMockResource) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AnotherMockResource.Unmarshal(m, b)
}
func (m *AnotherMockResource) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AnotherMockResource.Marshal(b, m, deterministic)
}
func (m *AnotherMockResource) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AnotherMockResource.Merge(m, src)
}
func (m *AnotherMockResource) XXX_Size() int {
	return xxx_messageInfo_AnotherMockResource.Size(m)
}
func (m *AnotherMockResource) XXX_DiscardUnknown() {
	xxx_messageInfo_AnotherMockResource.DiscardUnknown(m)
}

var xxx_messageInfo_AnotherMockResource proto.InternalMessageInfo

func (m *AnotherMockResource) GetMetadata() core.Metadata {
	if m != nil {
		return m.Metadata
	}
	return core.Metadata{}
}

func (m *AnotherMockResource) GetStatus() core.Status {
	if m != nil {
		return m.Status
	}
	return core.Status{}
}

func (m *AnotherMockResource) GetBasicField() string {
	if m != nil {
		return m.BasicField
	}
	return ""
}

type ClusterResource struct {
	Metadata core.Metadata `protobuf:"bytes,1,opt,name=metadata,proto3" json:"metadata"`
	Status   core.Status   `protobuf:"bytes,6,opt,name=status,proto3" json:"status"`
	// comments that go above the basic field in our docs
	BasicField           string   `protobuf:"bytes,2,opt,name=basic_field,json=basicField,proto3" json:"basic_field,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *ClusterResource) Reset()         { *m = ClusterResource{} }
func (m *ClusterResource) String() string { return proto.CompactTextString(m) }
func (*ClusterResource) ProtoMessage()    {}
func (*ClusterResource) Descriptor() ([]byte, []int) {
	return fileDescriptor_166414021e79b557, []int{1}
}
func (m *ClusterResource) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ClusterResource.Unmarshal(m, b)
}
func (m *ClusterResource) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ClusterResource.Marshal(b, m, deterministic)
}
func (m *ClusterResource) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ClusterResource.Merge(m, src)
}
func (m *ClusterResource) XXX_Size() int {
	return xxx_messageInfo_ClusterResource.Size(m)
}
func (m *ClusterResource) XXX_DiscardUnknown() {
	xxx_messageInfo_ClusterResource.DiscardUnknown(m)
}

var xxx_messageInfo_ClusterResource proto.InternalMessageInfo

func (m *ClusterResource) GetMetadata() core.Metadata {
	if m != nil {
		return m.Metadata
	}
	return core.Metadata{}
}

func (m *ClusterResource) GetStatus() core.Status {
	if m != nil {
		return m.Status
	}
	return core.Status{}
}

func (m *ClusterResource) GetBasicField() string {
	if m != nil {
		return m.BasicField
	}
	return ""
}

func init() {
	proto.RegisterType((*AnotherMockResource)(nil), "testing.solo.io.AnotherMockResource")
	proto.RegisterType((*ClusterResource)(nil), "testing.solo.io.ClusterResource")
}

func init() {
	proto.RegisterFile("solo-kit/test/mocks/api/v1/more_mock_resources.proto", fileDescriptor_166414021e79b557)
}

var fileDescriptor_166414021e79b557 = []byte{
	// 347 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x52, 0xcd, 0x4e, 0xea, 0x40,
	0x14, 0xbe, 0xc3, 0xed, 0x25, 0xdc, 0x61, 0x81, 0x19, 0x09, 0x69, 0x88, 0x01, 0x82, 0x31, 0xc1,
	0x85, 0x9d, 0x20, 0xc6, 0x18, 0x77, 0x62, 0xe2, 0x8e, 0x0d, 0xee, 0xdc, 0x90, 0x61, 0x18, 0xcb,
	0x84, 0x96, 0x43, 0x66, 0xa6, 0x84, 0x35, 0x4f, 0xe3, 0x23, 0xf8, 0x08, 0x26, 0xc6, 0x57, 0x70,
	0xe1, 0x1b, 0xb0, 0x70, 0x6f, 0xa6, 0xd3, 0x36, 0xd1, 0xe8, 0xde, 0x5d, 0xfb, 0xfd, 0xf4, 0x7c,
	0xdf, 0x39, 0xc5, 0x67, 0x1a, 0x22, 0x38, 0x59, 0x48, 0x43, 0x8d, 0xd0, 0x86, 0xc6, 0xc0, 0x17,
	0x9a, 0xb2, 0x95, 0xa4, 0xeb, 0x3e, 0x8d, 0x41, 0x89, 0x89, 0x45, 0x26, 0x4a, 0x68, 0x48, 0x14,
	0x17, 0x3a, 0x58, 0x29, 0x30, 0x40, 0x6a, 0x56, 0x2c, 0x97, 0x61, 0x60, 0xdd, 0x81, 0x84, 0x66,
	0x3d, 0x84, 0x10, 0x52, 0x8e, 0xda, 0x27, 0x27, 0x6b, 0x12, 0xb1, 0x31, 0x0e, 0x14, 0x1b, 0x93,
	0x61, 0xad, 0x62, 0x60, 0x3e, 0x45, 0x18, 0x36, 0x63, 0x86, 0x65, 0xfc, 0xc1, 0x57, 0x5e, 0x1b,
	0x66, 0x12, 0xfd, 0x93, 0x3b, 0x7f, 0x77, 0x7c, 0xf7, 0x05, 0xe1, 0xfd, 0xab, 0x25, 0x98, 0xb9,
	0x50, 0x23, 0xe0, 0x8b, 0x71, 0x96, 0x9b, 0x5c, 0xe0, 0x4a, 0x3e, 0xc7, 0x47, 0x1d, 0xd4, 0xab,
	0x9e, 0x36, 0x02, 0x0e, 0x4a, 0xe4, 0x05, 0x82, 0x51, 0xc6, 0x0e, 0xbd, 0xa7, 0xd7, 0xf6, 0x9f,
	0x71, 0xa1, 0x26, 0xe7, 0xb8, 0xec, 0x12, 0xf8, 0xe5, 0xd4, 0x57, 0xff, 0xec, 0xbb, 0x4d, 0xb9,
	0x61, 0xe5, 0xf1, 0xdd, 0x43, 0xa9, 0x33, 0x53, 0x93, 0x36, 0xae, 0x4e, 0x99, 0x96, 0x7c, 0x72,
	0x2f, 0x45, 0x34, 0xf3, 0x4b, 0x1d, 0xd4, 0xfb, 0x3f, 0xc6, 0x29, 0x74, 0x63, 0x91, 0xcb, 0xc3,
	0xed, 0xce, 0xfb, 0x87, 0xff, 0xb2, 0x58, 0x6d, 0x77, 0x5e, 0x83, 0xd4, 0x99, 0x8b, 0x6d, 0xf7,
	0x5d, 0xac, 0xbb, 0xfb, 0x8c, 0x70, 0xed, 0x3a, 0x4a, 0xb4, 0x11, 0xea, 0x37, 0x77, 0x39, 0x72,
	0x5d, 0x78, 0x64, 0xbb, 0x10, 0xb2, 0xc7, 0x5d, 0xe4, 0xa2, 0xc7, 0x76, 0xe7, 0x95, 0x7c, 0x34,
	0x1c, 0xd8, 0x2f, 0x3f, 0xbc, 0xb5, 0xd0, 0xdd, 0x71, 0x28, 0xcd, 0x3c, 0x99, 0x06, 0x1c, 0x62,
	0x77, 0x42, 0x09, 0xf4, 0xbb, 0x3f, 0x71, 0xdd, 0x9f, 0x96, 0xd3, 0xcb, 0x0e, 0x3e, 0x02, 0x00,
	0x00, 0xff, 0xff, 0x66, 0xd8, 0x5b, 0xb3, 0xaa, 0x02, 0x00, 0x00,
}

func (this *AnotherMockResource) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*AnotherMockResource)
	if !ok {
		that2, ok := that.(AnotherMockResource)
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
	if !this.Metadata.Equal(&that1.Metadata) {
		return false
	}
	if !this.Status.Equal(&that1.Status) {
		return false
	}
	if this.BasicField != that1.BasicField {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *ClusterResource) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ClusterResource)
	if !ok {
		that2, ok := that.(ClusterResource)
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
	if !this.Metadata.Equal(&that1.Metadata) {
		return false
	}
	if !this.Status.Equal(&that1.Status) {
		return false
	}
	if this.BasicField != that1.BasicField {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
