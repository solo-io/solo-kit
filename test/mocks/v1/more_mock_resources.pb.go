// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/test/mocks/api/v1/more_mock_resources.proto

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
	return fileDescriptor_3005c9d7690ad701, []int{0}
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
	return fileDescriptor_3005c9d7690ad701, []int{1}
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
	proto.RegisterFile("github.com/solo-io/solo-kit/test/mocks/api/v1/more_mock_resources.proto", fileDescriptor_3005c9d7690ad701)
}

var fileDescriptor_3005c9d7690ad701 = []byte{
	// 346 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x92, 0x31, 0x4f, 0xfa, 0x40,
	0x18, 0xc6, 0xff, 0xc7, 0xbf, 0xa2, 0x1e, 0x03, 0xe6, 0x24, 0xa4, 0x61, 0x10, 0x82, 0x31, 0xc1,
	0xc1, 0x5e, 0xc0, 0x98, 0x18, 0x37, 0x6b, 0xa2, 0x13, 0x0b, 0x6e, 0x2e, 0xe4, 0x38, 0xce, 0x72,
	0xa1, 0xe5, 0x25, 0x77, 0x57, 0xc2, 0xcc, 0xa7, 0xf1, 0xa3, 0x18, 0xe3, 0x67, 0x70, 0xf0, 0x1b,
	0xb0, 0x31, 0x9a, 0xeb, 0xb5, 0x24, 0x2e, 0x06, 0x37, 0xa7, 0xb6, 0xcf, 0xfb, 0xfc, 0xf2, 0x3e,
	0x4f, 0xde, 0xe2, 0x87, 0x48, 0x9a, 0x49, 0x3a, 0x0a, 0x38, 0x24, 0x54, 0x43, 0x0c, 0x17, 0x12,
	0xdc, 0x73, 0x2a, 0x0d, 0x35, 0x42, 0x1b, 0x9a, 0x00, 0x9f, 0x6a, 0xca, 0xe6, 0x92, 0x2e, 0xba,
	0x34, 0x01, 0x25, 0x86, 0x56, 0x19, 0x2a, 0xa1, 0x21, 0x55, 0x5c, 0xe8, 0x60, 0xae, 0xc0, 0x00,
	0xa9, 0x5a, 0xb3, 0x9c, 0x45, 0x81, 0xa5, 0x03, 0x09, 0x8d, 0x5a, 0x04, 0x11, 0x64, 0x33, 0x6a,
	0xdf, 0x9c, 0xad, 0x41, 0xc4, 0xd2, 0x38, 0x51, 0x2c, 0x4d, 0xae, 0x75, 0x7f, 0xca, 0x50, 0x2c,
	0x16, 0x86, 0x8d, 0x99, 0x61, 0x39, 0x42, 0x77, 0x40, 0xb4, 0x61, 0x26, 0xd5, 0xbf, 0xd8, 0x51,
	0x7c, 0x3b, 0xa4, 0xfd, 0x8e, 0xf0, 0xf1, 0xed, 0x0c, 0xcc, 0x44, 0xa8, 0x3e, 0xf0, 0xe9, 0x20,
	0x2f, 0x4c, 0xae, 0xf1, 0x41, 0x91, 0xc6, 0x47, 0x2d, 0xd4, 0xa9, 0xf4, 0xea, 0x01, 0x07, 0x25,
	0x8a, 0xe6, 0x41, 0x3f, 0x9f, 0x86, 0xde, 0xeb, 0x47, 0xf3, 0xdf, 0x60, 0xeb, 0x26, 0x57, 0xb8,
	0xec, 0x42, 0xf9, 0xe5, 0x8c, 0xab, 0x7d, 0xe7, 0x1e, 0xb3, 0x59, 0xb8, 0xbf, 0x09, 0x51, 0x06,
	0xe6, 0x66, 0xd2, 0xc4, 0x95, 0x11, 0xd3, 0x92, 0x0f, 0x9f, 0xa5, 0x88, 0xc7, 0x7e, 0xa9, 0x85,
	0x3a, 0x87, 0x03, 0x9c, 0x49, 0xf7, 0x56, 0xb9, 0x39, 0x5d, 0xad, 0xbd, 0x3d, 0xfc, 0x9f, 0x25,
	0x6a, 0xb5, 0xf6, 0xea, 0xa4, 0xc6, 0x5c, 0x6a, 0x7b, 0xa7, 0xed, 0x99, 0xda, 0x6f, 0x08, 0x57,
	0xef, 0xe2, 0x54, 0x1b, 0xa1, 0xfe, 0x70, 0x95, 0x33, 0x57, 0x85, 0xc7, 0xb6, 0x0a, 0x21, 0x47,
	0xdc, 0x25, 0xde, 0xd6, 0x58, 0xad, 0xbd, 0x92, 0x8f, 0xc2, 0xde, 0x26, 0x44, 0x2f, 0x9f, 0x27,
	0xe8, 0xe9, 0x7c, 0xc7, 0xff, 0x77, 0xd1, 0x1d, 0x95, 0xb3, 0xb3, 0x5e, 0x7e, 0x05, 0x00, 0x00,
	0xff, 0xff, 0x43, 0xc6, 0x14, 0xc3, 0xf3, 0x02, 0x00, 0x00,
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
