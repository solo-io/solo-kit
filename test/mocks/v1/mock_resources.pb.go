// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/test/mocks/api/v1/mock_resources.proto

//
//package Comments
//package Comments a

package v1

import (
	bytes "bytes"
	context "context"
	fmt "fmt"
	v2 "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/solo-io/protoc-gen-ext/ext"
	core "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	_ "google.golang.org/genproto/googleapis/api/annotations"
	grpc "google.golang.org/grpc"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
	math "math"
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
//Mock resources for goofin off
type MockResource struct {
	Status        core.Status   `protobuf:"bytes,6,opt,name=status,proto3" json:"status"`
	Metadata      core.Metadata `protobuf:"bytes,7,opt,name=metadata,proto3" json:"metadata"`
	Data          string        `protobuf:"bytes,1,opt,name=data,json=data.json,proto3" json:"data.json"`
	SomeDumbField string        `protobuf:"bytes,100,opt,name=some_dumb_field,json=someDumbField,proto3" json:"some_dumb_field,omitempty"`
	// Types that are valid to be assigned to TestOneofFields:
	//	*MockResource_OneofOne
	//	*MockResource_OneofTwo
	TestOneofFields      isMockResource_TestOneofFields `protobuf_oneof:"test_oneof_fields"`
	XXX_NoUnkeyedLiteral struct{}                       `json:"-"`
	XXX_unrecognized     []byte                         `json:"-"`
	XXX_sizecache        int32                          `json:"-"`
}

func (m *MockResource) Reset()         { *m = MockResource{} }
func (m *MockResource) String() string { return proto.CompactTextString(m) }
func (*MockResource) ProtoMessage()    {}
func (*MockResource) Descriptor() ([]byte, []int) {
	return fileDescriptor_5de7a91ad5dc71ff, []int{0}
}
func (m *MockResource) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MockResource.Unmarshal(m, b)
}
func (m *MockResource) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MockResource.Marshal(b, m, deterministic)
}
func (m *MockResource) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MockResource.Merge(m, src)
}
func (m *MockResource) XXX_Size() int {
	return xxx_messageInfo_MockResource.Size(m)
}
func (m *MockResource) XXX_DiscardUnknown() {
	xxx_messageInfo_MockResource.DiscardUnknown(m)
}

var xxx_messageInfo_MockResource proto.InternalMessageInfo

type isMockResource_TestOneofFields interface {
	isMockResource_TestOneofFields()
	Equal(interface{}) bool
}

type MockResource_OneofOne struct {
	OneofOne string `protobuf:"bytes,3,opt,name=oneof_one,json=oneofOne,proto3,oneof" json:"oneof_one,omitempty"`
}
type MockResource_OneofTwo struct {
	OneofTwo bool `protobuf:"varint,2,opt,name=oneof_two,json=oneofTwo,proto3,oneof" json:"oneof_two,omitempty"`
}

func (*MockResource_OneofOne) isMockResource_TestOneofFields() {}
func (*MockResource_OneofTwo) isMockResource_TestOneofFields() {}

func (m *MockResource) GetTestOneofFields() isMockResource_TestOneofFields {
	if m != nil {
		return m.TestOneofFields
	}
	return nil
}

func (m *MockResource) GetStatus() core.Status {
	if m != nil {
		return m.Status
	}
	return core.Status{}
}

func (m *MockResource) GetMetadata() core.Metadata {
	if m != nil {
		return m.Metadata
	}
	return core.Metadata{}
}

func (m *MockResource) GetData() string {
	if m != nil {
		return m.Data
	}
	return ""
}

func (m *MockResource) GetSomeDumbField() string {
	if m != nil {
		return m.SomeDumbField
	}
	return ""
}

func (m *MockResource) GetOneofOne() string {
	if x, ok := m.GetTestOneofFields().(*MockResource_OneofOne); ok {
		return x.OneofOne
	}
	return ""
}

func (m *MockResource) GetOneofTwo() bool {
	if x, ok := m.GetTestOneofFields().(*MockResource_OneofTwo); ok {
		return x.OneofTwo
	}
	return false
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*MockResource) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*MockResource_OneofOne)(nil),
		(*MockResource_OneofTwo)(nil),
	}
}

type FakeResource struct {
	Count                uint32        `protobuf:"varint,1,opt,name=count,proto3" json:"count,omitempty"`
	Metadata             core.Metadata `protobuf:"bytes,7,opt,name=metadata,proto3" json:"metadata"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *FakeResource) Reset()         { *m = FakeResource{} }
func (m *FakeResource) String() string { return proto.CompactTextString(m) }
func (*FakeResource) ProtoMessage()    {}
func (*FakeResource) Descriptor() ([]byte, []int) {
	return fileDescriptor_5de7a91ad5dc71ff, []int{1}
}
func (m *FakeResource) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_FakeResource.Unmarshal(m, b)
}
func (m *FakeResource) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_FakeResource.Marshal(b, m, deterministic)
}
func (m *FakeResource) XXX_Merge(src proto.Message) {
	xxx_messageInfo_FakeResource.Merge(m, src)
}
func (m *FakeResource) XXX_Size() int {
	return xxx_messageInfo_FakeResource.Size(m)
}
func (m *FakeResource) XXX_DiscardUnknown() {
	xxx_messageInfo_FakeResource.DiscardUnknown(m)
}

var xxx_messageInfo_FakeResource proto.InternalMessageInfo

func (m *FakeResource) GetCount() uint32 {
	if m != nil {
		return m.Count
	}
	return 0
}

func (m *FakeResource) GetMetadata() core.Metadata {
	if m != nil {
		return m.Metadata
	}
	return core.Metadata{}
}

//
//@solo-kit:xds-service=MockXdsResourceDiscoveryService
//@solo-kit:resource.no_references
type MockXdsResourceConfig struct {
	// @solo-kit:resource.name
	Domain               string   `protobuf:"bytes,1,opt,name=domain,proto3" json:"domain,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *MockXdsResourceConfig) Reset()         { *m = MockXdsResourceConfig{} }
func (m *MockXdsResourceConfig) String() string { return proto.CompactTextString(m) }
func (*MockXdsResourceConfig) ProtoMessage()    {}
func (*MockXdsResourceConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_5de7a91ad5dc71ff, []int{2}
}
func (m *MockXdsResourceConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_MockXdsResourceConfig.Unmarshal(m, b)
}
func (m *MockXdsResourceConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_MockXdsResourceConfig.Marshal(b, m, deterministic)
}
func (m *MockXdsResourceConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_MockXdsResourceConfig.Merge(m, src)
}
func (m *MockXdsResourceConfig) XXX_Size() int {
	return xxx_messageInfo_MockXdsResourceConfig.Size(m)
}
func (m *MockXdsResourceConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_MockXdsResourceConfig.DiscardUnknown(m)
}

var xxx_messageInfo_MockXdsResourceConfig proto.InternalMessageInfo

func (m *MockXdsResourceConfig) GetDomain() string {
	if m != nil {
		return m.Domain
	}
	return ""
}

func init() {
	proto.RegisterType((*MockResource)(nil), "testing.solo.io.MockResource")
	proto.RegisterType((*FakeResource)(nil), "testing.solo.io.FakeResource")
	proto.RegisterType((*MockXdsResourceConfig)(nil), "testing.solo.io.MockXdsResourceConfig")
}

func init() {
	proto.RegisterFile("github.com/solo-io/solo-kit/test/mocks/api/v1/mock_resources.proto", fileDescriptor_5de7a91ad5dc71ff)
}

var fileDescriptor_5de7a91ad5dc71ff = []byte{
	// 593 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xb4, 0x94, 0xc1, 0x6e, 0xd3, 0x4c,
	0x10, 0xc7, 0xbb, 0x69, 0x9a, 0xb6, 0xfb, 0x35, 0xaa, 0x3e, 0xb7, 0x54, 0x91, 0x29, 0x4d, 0x95,
	0x22, 0x14, 0x40, 0xd8, 0x34, 0x08, 0x09, 0xf5, 0x68, 0xaa, 0x8a, 0x4b, 0x85, 0xe4, 0x72, 0x40,
	0x5c, 0xa2, 0x8d, 0x3d, 0x71, 0x17, 0xc7, 0x3b, 0xc1, 0xbb, 0x76, 0xd3, 0x6b, 0x0e, 0xdc, 0x38,
	0xf1, 0x12, 0x3c, 0x04, 0x0f, 0xc0, 0x9d, 0x7b, 0x0f, 0x88, 0x13, 0xb7, 0xde, 0x72, 0x44, 0x5e,
	0xdb, 0xa9, 0x52, 0x4a, 0x55, 0x90, 0x38, 0x79, 0x67, 0xff, 0xf3, 0x9b, 0x99, 0x1d, 0xcf, 0x2e,
	0x75, 0x02, 0xae, 0x8e, 0x93, 0x9e, 0xe5, 0x61, 0x64, 0x4b, 0x1c, 0xe0, 0x23, 0x8e, 0xf9, 0x37,
	0xe4, 0xca, 0x56, 0x20, 0x95, 0x1d, 0xa1, 0x17, 0x4a, 0x9b, 0x0d, 0xb9, 0x9d, 0xee, 0x6a, 0xa3,
	0x1b, 0x83, 0xc4, 0x24, 0xf6, 0x40, 0x5a, 0xc3, 0x18, 0x15, 0x1a, 0xab, 0x99, 0x1f, 0x17, 0x81,
	0x95, 0x81, 0x16, 0x47, 0x73, 0x13, 0x44, 0x8a, 0xa7, 0x39, 0xd3, 0xb1, 0x7d, 0x2e, 0x3d, 0x4c,
	0x21, 0x3e, 0xcd, 0xdd, 0xcd, 0xcd, 0x00, 0x31, 0x18, 0x80, 0x96, 0x99, 0x10, 0xa8, 0x98, 0xe2,
	0x28, 0x8a, 0x60, 0xe6, 0x7a, 0x80, 0x01, 0xea, 0xa5, 0x9d, 0xad, 0x8a, 0xdd, 0x3a, 0x8c, 0x94,
	0x0d, 0x23, 0x55, 0x98, 0xbb, 0xd7, 0x55, 0x5d, 0x96, 0x0a, 0x8a, 0xf9, 0x4c, 0xb1, 0x02, 0xb1,
	0x6f, 0x80, 0x48, 0xc5, 0x54, 0x22, 0xff, 0x20, 0x47, 0x69, 0xe7, 0x48, 0xeb, 0x73, 0x85, 0xae,
	0x1c, 0xa2, 0x17, 0xba, 0x45, 0x83, 0x8c, 0xa7, 0xb4, 0x96, 0xc7, 0x6c, 0xd4, 0xb6, 0x49, 0xfb,
	0xbf, 0xce, 0xba, 0xe5, 0x61, 0x0c, 0x65, 0x9f, 0xac, 0x23, 0xad, 0x39, 0x8b, 0x13, 0x87, 0x7c,
	0x39, 0x6b, 0xce, 0xb9, 0x85, 0xb3, 0xf1, 0x8c, 0x2e, 0x95, 0xd5, 0x37, 0x16, 0x35, 0xb8, 0x31,
	0x0b, 0x1e, 0x16, 0xaa, 0x53, 0xd5, 0xdc, 0xd4, 0xdb, 0xb8, 0x47, 0xab, 0x9a, 0x22, 0xdb, 0xa4,
	0xbd, 0xec, 0xd4, 0x7f, 0x9c, 0x35, 0x97, 0x75, 0x0f, 0xde, 0x4a, 0x14, 0xee, 0xc5, 0xd2, 0x78,
	0x48, 0x57, 0x25, 0x46, 0xd0, 0xf5, 0x93, 0xa8, 0xd7, 0xed, 0x73, 0x18, 0xf8, 0x0d, 0x5f, 0x23,
	0xf3, 0x13, 0x87, 0xb8, 0xf5, 0x4c, 0xdb, 0x4f, 0xa2, 0xde, 0x41, 0xa6, 0x18, 0x77, 0xe8, 0x32,
	0x0a, 0xc0, 0x7e, 0x17, 0x05, 0x34, 0xe6, 0x33, 0xb7, 0x17, 0x73, 0xee, 0x92, 0xde, 0x7a, 0x29,
	0xe0, 0x42, 0x56, 0x27, 0xd8, 0xa8, 0x6c, 0x93, 0xf6, 0xd2, 0x54, 0x7e, 0x75, 0x82, 0x7b, 0x6b,
	0xe3, 0xf3, 0x6a, 0x95, 0x56, 0xa2, 0x70, 0x7c, 0x5e, 0x5d, 0x34, 0x16, 0xf4, 0x38, 0x39, 0x6b,
	0xf4, 0xff, 0x6c, 0x68, 0xba, 0x39, 0xa8, 0x0b, 0x90, 0x2d, 0x49, 0x57, 0x0e, 0x58, 0x08, 0xd3,
	0xee, 0xad, 0xd3, 0x05, 0x0f, 0x13, 0xa1, 0xf4, 0x69, 0xea, 0x6e, 0x6e, 0xfc, 0x7d, 0x73, 0xca,
	0x4a, 0xfa, 0x45, 0x25, 0x7d, 0x16, 0x82, 0x6c, 0xd9, 0xf4, 0x56, 0xf6, 0xcb, 0x5e, 0xfb, 0xb2,
	0xcc, 0xfb, 0x1c, 0x45, 0x9f, 0x07, 0xc6, 0x06, 0xad, 0xf9, 0x18, 0x31, 0x2e, 0xf2, 0x66, 0xba,
	0x85, 0xd5, 0x79, 0x3f, 0x4f, 0x9b, 0x97, 0x88, 0xfd, 0x72, 0xc2, 0x8f, 0x20, 0x4e, 0xb9, 0x07,
	0x86, 0x4f, 0x6f, 0x1f, 0xa9, 0x18, 0x58, 0x74, 0x75, 0xe8, 0x2d, 0x4b, 0x5f, 0x10, 0x8b, 0x0d,
	0xb9, 0x95, 0x76, 0xac, 0x29, 0xee, 0xc2, 0xbb, 0x04, 0xa4, 0x32, 0x9b, 0xbf, 0xd5, 0xe5, 0x10,
	0x85, 0x84, 0xd6, 0x5c, 0x9b, 0x3c, 0x26, 0x46, 0x44, 0xcd, 0x7d, 0x18, 0x28, 0x76, 0x75, 0x92,
	0x9d, 0x4b, 0x41, 0x32, 0xcf, 0x5f, 0x32, 0xdd, 0xbd, 0xde, 0x69, 0x26, 0xdd, 0x07, 0x42, 0xcd,
	0x03, 0x50, 0xde, 0xf1, 0x3f, 0x3a, 0x94, 0x35, 0xfe, 0xfa, 0xfd, 0x63, 0xa5, 0xdd, 0xda, 0x99,
	0x79, 0x34, 0xf6, 0xb2, 0x81, 0x19, 0xf9, 0xb2, 0x7c, 0x74, 0x3c, 0x9d, 0x6d, 0x8f, 0x3c, 0x70,
	0x3a, 0x13, 0x87, 0x7c, 0xfa, 0xb6, 0x45, 0xde, 0xdc, 0xbf, 0xe1, 0x1b, 0x96, 0xee, 0xf6, 0x6a,
	0xfa, 0xa2, 0x3e, 0xf9, 0x19, 0x00, 0x00, 0xff, 0xff, 0x51, 0x59, 0x49, 0x71, 0xf7, 0x04, 0x00,
	0x00,
}

func (this *MockResource) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*MockResource)
	if !ok {
		that2, ok := that.(MockResource)
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
	if !this.Status.Equal(&that1.Status) {
		return false
	}
	if !this.Metadata.Equal(&that1.Metadata) {
		return false
	}
	if this.Data != that1.Data {
		return false
	}
	if this.SomeDumbField != that1.SomeDumbField {
		return false
	}
	if that1.TestOneofFields == nil {
		if this.TestOneofFields != nil {
			return false
		}
	} else if this.TestOneofFields == nil {
		return false
	} else if !this.TestOneofFields.Equal(that1.TestOneofFields) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *MockResource_OneofOne) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*MockResource_OneofOne)
	if !ok {
		that2, ok := that.(MockResource_OneofOne)
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
	if this.OneofOne != that1.OneofOne {
		return false
	}
	return true
}
func (this *MockResource_OneofTwo) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*MockResource_OneofTwo)
	if !ok {
		that2, ok := that.(MockResource_OneofTwo)
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
	if this.OneofTwo != that1.OneofTwo {
		return false
	}
	return true
}
func (this *FakeResource) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*FakeResource)
	if !ok {
		that2, ok := that.(FakeResource)
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
	if this.Count != that1.Count {
		return false
	}
	if !this.Metadata.Equal(&that1.Metadata) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
func (this *MockXdsResourceConfig) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*MockXdsResourceConfig)
	if !ok {
		that2, ok := that.(MockXdsResourceConfig)
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
	if this.Domain != that1.Domain {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// MockXdsResourceDiscoveryServiceClient is the client API for MockXdsResourceDiscoveryService service.
//
// For semantics around ctx use and closing/ending streaming RPCs, please refer to https://godoc.org/google.golang.org/grpc#ClientConn.NewStream.
type MockXdsResourceDiscoveryServiceClient interface {
	StreamMockXdsResourceConfig(ctx context.Context, opts ...grpc.CallOption) (MockXdsResourceDiscoveryService_StreamMockXdsResourceConfigClient, error)
	DeltaMockXdsResourceConfig(ctx context.Context, opts ...grpc.CallOption) (MockXdsResourceDiscoveryService_DeltaMockXdsResourceConfigClient, error)
	FetchMockXdsResourceConfig(ctx context.Context, in *v2.DiscoveryRequest, opts ...grpc.CallOption) (*v2.DiscoveryResponse, error)
}

type mockXdsResourceDiscoveryServiceClient struct {
	cc *grpc.ClientConn
}

func NewMockXdsResourceDiscoveryServiceClient(cc *grpc.ClientConn) MockXdsResourceDiscoveryServiceClient {
	return &mockXdsResourceDiscoveryServiceClient{cc}
}

func (c *mockXdsResourceDiscoveryServiceClient) StreamMockXdsResourceConfig(ctx context.Context, opts ...grpc.CallOption) (MockXdsResourceDiscoveryService_StreamMockXdsResourceConfigClient, error) {
	stream, err := c.cc.NewStream(ctx, &_MockXdsResourceDiscoveryService_serviceDesc.Streams[0], "/testing.solo.io.MockXdsResourceDiscoveryService/StreamMockXdsResourceConfig", opts...)
	if err != nil {
		return nil, err
	}
	x := &mockXdsResourceDiscoveryServiceStreamMockXdsResourceConfigClient{stream}
	return x, nil
}

type MockXdsResourceDiscoveryService_StreamMockXdsResourceConfigClient interface {
	Send(*v2.DiscoveryRequest) error
	Recv() (*v2.DiscoveryResponse, error)
	grpc.ClientStream
}

type mockXdsResourceDiscoveryServiceStreamMockXdsResourceConfigClient struct {
	grpc.ClientStream
}

func (x *mockXdsResourceDiscoveryServiceStreamMockXdsResourceConfigClient) Send(m *v2.DiscoveryRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *mockXdsResourceDiscoveryServiceStreamMockXdsResourceConfigClient) Recv() (*v2.DiscoveryResponse, error) {
	m := new(v2.DiscoveryResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *mockXdsResourceDiscoveryServiceClient) DeltaMockXdsResourceConfig(ctx context.Context, opts ...grpc.CallOption) (MockXdsResourceDiscoveryService_DeltaMockXdsResourceConfigClient, error) {
	stream, err := c.cc.NewStream(ctx, &_MockXdsResourceDiscoveryService_serviceDesc.Streams[1], "/testing.solo.io.MockXdsResourceDiscoveryService/DeltaMockXdsResourceConfig", opts...)
	if err != nil {
		return nil, err
	}
	x := &mockXdsResourceDiscoveryServiceDeltaMockXdsResourceConfigClient{stream}
	return x, nil
}

type MockXdsResourceDiscoveryService_DeltaMockXdsResourceConfigClient interface {
	Send(*v2.DeltaDiscoveryRequest) error
	Recv() (*v2.DeltaDiscoveryResponse, error)
	grpc.ClientStream
}

type mockXdsResourceDiscoveryServiceDeltaMockXdsResourceConfigClient struct {
	grpc.ClientStream
}

func (x *mockXdsResourceDiscoveryServiceDeltaMockXdsResourceConfigClient) Send(m *v2.DeltaDiscoveryRequest) error {
	return x.ClientStream.SendMsg(m)
}

func (x *mockXdsResourceDiscoveryServiceDeltaMockXdsResourceConfigClient) Recv() (*v2.DeltaDiscoveryResponse, error) {
	m := new(v2.DeltaDiscoveryResponse)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *mockXdsResourceDiscoveryServiceClient) FetchMockXdsResourceConfig(ctx context.Context, in *v2.DiscoveryRequest, opts ...grpc.CallOption) (*v2.DiscoveryResponse, error) {
	out := new(v2.DiscoveryResponse)
	err := c.cc.Invoke(ctx, "/testing.solo.io.MockXdsResourceDiscoveryService/FetchMockXdsResourceConfig", in, out, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// MockXdsResourceDiscoveryServiceServer is the server API for MockXdsResourceDiscoveryService service.
type MockXdsResourceDiscoveryServiceServer interface {
	StreamMockXdsResourceConfig(MockXdsResourceDiscoveryService_StreamMockXdsResourceConfigServer) error
	DeltaMockXdsResourceConfig(MockXdsResourceDiscoveryService_DeltaMockXdsResourceConfigServer) error
	FetchMockXdsResourceConfig(context.Context, *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error)
}

// UnimplementedMockXdsResourceDiscoveryServiceServer can be embedded to have forward compatible implementations.
type UnimplementedMockXdsResourceDiscoveryServiceServer struct {
}

func (*UnimplementedMockXdsResourceDiscoveryServiceServer) StreamMockXdsResourceConfig(srv MockXdsResourceDiscoveryService_StreamMockXdsResourceConfigServer) error {
	return status.Errorf(codes.Unimplemented, "method StreamMockXdsResourceConfig not implemented")
}
func (*UnimplementedMockXdsResourceDiscoveryServiceServer) DeltaMockXdsResourceConfig(srv MockXdsResourceDiscoveryService_DeltaMockXdsResourceConfigServer) error {
	return status.Errorf(codes.Unimplemented, "method DeltaMockXdsResourceConfig not implemented")
}
func (*UnimplementedMockXdsResourceDiscoveryServiceServer) FetchMockXdsResourceConfig(ctx context.Context, req *v2.DiscoveryRequest) (*v2.DiscoveryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "method FetchMockXdsResourceConfig not implemented")
}

func RegisterMockXdsResourceDiscoveryServiceServer(s *grpc.Server, srv MockXdsResourceDiscoveryServiceServer) {
	s.RegisterService(&_MockXdsResourceDiscoveryService_serviceDesc, srv)
}

func _MockXdsResourceDiscoveryService_StreamMockXdsResourceConfig_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MockXdsResourceDiscoveryServiceServer).StreamMockXdsResourceConfig(&mockXdsResourceDiscoveryServiceStreamMockXdsResourceConfigServer{stream})
}

type MockXdsResourceDiscoveryService_StreamMockXdsResourceConfigServer interface {
	Send(*v2.DiscoveryResponse) error
	Recv() (*v2.DiscoveryRequest, error)
	grpc.ServerStream
}

type mockXdsResourceDiscoveryServiceStreamMockXdsResourceConfigServer struct {
	grpc.ServerStream
}

func (x *mockXdsResourceDiscoveryServiceStreamMockXdsResourceConfigServer) Send(m *v2.DiscoveryResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *mockXdsResourceDiscoveryServiceStreamMockXdsResourceConfigServer) Recv() (*v2.DiscoveryRequest, error) {
	m := new(v2.DiscoveryRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _MockXdsResourceDiscoveryService_DeltaMockXdsResourceConfig_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(MockXdsResourceDiscoveryServiceServer).DeltaMockXdsResourceConfig(&mockXdsResourceDiscoveryServiceDeltaMockXdsResourceConfigServer{stream})
}

type MockXdsResourceDiscoveryService_DeltaMockXdsResourceConfigServer interface {
	Send(*v2.DeltaDiscoveryResponse) error
	Recv() (*v2.DeltaDiscoveryRequest, error)
	grpc.ServerStream
}

type mockXdsResourceDiscoveryServiceDeltaMockXdsResourceConfigServer struct {
	grpc.ServerStream
}

func (x *mockXdsResourceDiscoveryServiceDeltaMockXdsResourceConfigServer) Send(m *v2.DeltaDiscoveryResponse) error {
	return x.ServerStream.SendMsg(m)
}

func (x *mockXdsResourceDiscoveryServiceDeltaMockXdsResourceConfigServer) Recv() (*v2.DeltaDiscoveryRequest, error) {
	m := new(v2.DeltaDiscoveryRequest)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func _MockXdsResourceDiscoveryService_FetchMockXdsResourceConfig_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(v2.DiscoveryRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(MockXdsResourceDiscoveryServiceServer).FetchMockXdsResourceConfig(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/testing.solo.io.MockXdsResourceDiscoveryService/FetchMockXdsResourceConfig",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(MockXdsResourceDiscoveryServiceServer).FetchMockXdsResourceConfig(ctx, req.(*v2.DiscoveryRequest))
	}
	return interceptor(ctx, in, info, handler)
}

var _MockXdsResourceDiscoveryService_serviceDesc = grpc.ServiceDesc{
	ServiceName: "testing.solo.io.MockXdsResourceDiscoveryService",
	HandlerType: (*MockXdsResourceDiscoveryServiceServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "FetchMockXdsResourceConfig",
			Handler:    _MockXdsResourceDiscoveryService_FetchMockXdsResourceConfig_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamMockXdsResourceConfig",
			Handler:       _MockXdsResourceDiscoveryService_StreamMockXdsResourceConfig_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
		{
			StreamName:    "DeltaMockXdsResourceConfig",
			Handler:       _MockXdsResourceDiscoveryService_DeltaMockXdsResourceConfig_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "github.com/solo-io/solo-kit/test/mocks/api/v1/mock_resources.proto",
}
