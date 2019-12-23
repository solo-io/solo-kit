// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: test/mocks/api/v1/more_mock_resources.proto

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
	return fileDescriptor_e10da893cfb8d651, []int{0}
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
	return fileDescriptor_e10da893cfb8d651, []int{1}
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
	proto.RegisterFile("test/mocks/api/v1/more_mock_resources.proto", fileDescriptor_e10da893cfb8d651)
}

var fileDescriptor_e10da893cfb8d651 = []byte{
	// 344 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xcc, 0x52, 0xcf, 0x4a, 0x02, 0x41,
	0x18, 0x6f, 0x6c, 0x13, 0x1b, 0x0f, 0xc6, 0x24, 0xb2, 0x48, 0xa8, 0x18, 0x81, 0x11, 0xed, 0x62,
	0x42, 0x44, 0xb7, 0x0c, 0xba, 0x79, 0xb1, 0x5b, 0x17, 0x19, 0xc7, 0x69, 0x1d, 0xdc, 0xf5, 0x93,
	0x99, 0x59, 0xf1, 0xec, 0xd3, 0xf4, 0x08, 0x3d, 0x42, 0x10, 0xbd, 0x42, 0x87, 0xde, 0xc0, 0x43,
	0xf7, 0x98, 0x99, 0x75, 0xa9, 0xa0, 0x7b, 0xb7, 0xdd, 0xdf, 0x9f, 0xfd, 0x7e, 0xbf, 0xef, 0x5b,
	0x7c, 0xa6, 0xb9, 0xd2, 0x61, 0x02, 0x6c, 0xa6, 0x42, 0xba, 0x10, 0xe1, 0xb2, 0x1b, 0x26, 0x20,
	0xf9, 0xc8, 0x20, 0x23, 0xc9, 0x15, 0xa4, 0x92, 0x71, 0x15, 0x2c, 0x24, 0x68, 0x20, 0x15, 0x23,
	0x16, 0xf3, 0x28, 0x50, 0x10, 0x43, 0x20, 0xa0, 0x5e, 0x8d, 0x20, 0x02, 0xcb, 0x85, 0xe6, 0xc9,
	0xc9, 0xea, 0x84, 0xaf, 0xb4, 0x03, 0xf9, 0x4a, 0x67, 0x58, 0xc3, 0x58, 0xce, 0x67, 0x42, 0xe7,
	0x53, 0xb8, 0xa6, 0x13, 0xaa, 0x69, 0xc6, 0x1f, 0xfd, 0xe6, 0x95, 0xa6, 0x3a, 0x55, 0x7f, 0xb9,
	0xb7, 0xef, 0x8e, 0x6f, 0xbf, 0x21, 0x7c, 0x78, 0x33, 0x07, 0x3d, 0xe5, 0x72, 0x00, 0x6c, 0x36,
	0xcc, 0x72, 0x93, 0x2b, 0x5c, 0xda, 0xce, 0xf1, 0x51, 0x0b, 0x75, 0xca, 0x17, 0xb5, 0x80, 0x81,
	0xe4, 0xdb, 0x02, 0xc1, 0x20, 0x63, 0xfb, 0xde, 0xcb, 0x7b, 0x73, 0x67, 0x98, 0xab, 0xc9, 0x25,
	0x2e, 0xba, 0x04, 0x7e, 0xd1, 0xfa, 0xaa, 0x3f, 0x7d, 0xf7, 0x96, 0xeb, 0x97, 0x9e, 0x3f, 0x3d,
	0x64, 0x9d, 0x99, 0x9a, 0x34, 0x71, 0x79, 0x4c, 0x95, 0x60, 0xa3, 0x47, 0xc1, 0xe3, 0x89, 0x5f,
	0x68, 0xa1, 0xce, 0xfe, 0x10, 0x5b, 0xe8, 0xce, 0x20, 0xd7, 0xc7, 0xeb, 0x8d, 0xb7, 0x87, 0x77,
	0x69, 0x22, 0xd7, 0x1b, 0xaf, 0x46, 0xaa, 0xd4, 0xc5, 0x36, 0xfb, 0xce, 0xd7, 0xdd, 0x7e, 0x45,
	0xb8, 0x72, 0x1b, 0xa7, 0x4a, 0x73, 0xf9, 0x9f, 0xbb, 0x9c, 0xb8, 0x2e, 0x2c, 0x36, 0x5d, 0x08,
	0x39, 0x60, 0x2e, 0x72, 0xde, 0x63, 0xbd, 0xf1, 0x0a, 0x3e, 0xea, 0xf7, 0xcc, 0x97, 0x9f, 0x3e,
	0x1a, 0xe8, 0xe1, 0x34, 0x12, 0x7a, 0x9a, 0x8e, 0x03, 0x06, 0x89, 0x3b, 0xa1, 0x80, 0xfc, 0x94,
	0xe1, 0xb7, 0x3f, 0x71, 0xd9, 0x1d, 0x17, 0xed, 0x65, 0x7b, 0x5f, 0x01, 0x00, 0x00, 0xff, 0xff,
	0x64, 0xa6, 0x01, 0xfe, 0xa1, 0x02, 0x00, 0x00,
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
