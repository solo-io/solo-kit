// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/test/mocks/api/v1/complicated_mock_resources.proto

package v1

import (
	fmt "fmt"
	math "math"

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
//A ComplicatedMockResource is used to validate our schemagen tool, which converts
//protos to OpenApi schemas with structural constraints.
type ComplicatedMockResource struct {
	Status   *core.Status   `protobuf:"bytes,1,opt,name=status,proto3" json:"status,omitempty"`
	Metadata *core.Metadata `protobuf:"bytes,2,opt,name=metadata,proto3" json:"metadata,omitempty"`
	// comment
	NestedOneOf          *NestedOneOf    `protobuf:"bytes,3,opt,name=nested_one_of,json=nestedOneOf,proto3" json:"nested_one_of,omitempty"`
	Name                 string          `protobuf:"bytes,4,opt,name=name,proto3" json:"name,omitempty"`
	SslConfig            *LocalSslConfig `protobuf:"bytes,5,opt,name=ssl_config,json=sslConfig,proto3" json:"ssl_config,omitempty"`
	XXX_NoUnkeyedLiteral struct{}        `json:"-"`
	XXX_unrecognized     []byte          `json:"-"`
	XXX_sizecache        int32           `json:"-"`
}

func (m *ComplicatedMockResource) Reset()         { *m = ComplicatedMockResource{} }
func (m *ComplicatedMockResource) String() string { return proto.CompactTextString(m) }
func (*ComplicatedMockResource) ProtoMessage()    {}
func (*ComplicatedMockResource) Descriptor() ([]byte, []int) {
	return fileDescriptor_c6e0c163912ae244, []int{0}
}
func (m *ComplicatedMockResource) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_ComplicatedMockResource.Unmarshal(m, b)
}
func (m *ComplicatedMockResource) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_ComplicatedMockResource.Marshal(b, m, deterministic)
}
func (m *ComplicatedMockResource) XXX_Merge(src proto.Message) {
	xxx_messageInfo_ComplicatedMockResource.Merge(m, src)
}
func (m *ComplicatedMockResource) XXX_Size() int {
	return xxx_messageInfo_ComplicatedMockResource.Size(m)
}
func (m *ComplicatedMockResource) XXX_DiscardUnknown() {
	xxx_messageInfo_ComplicatedMockResource.DiscardUnknown(m)
}

var xxx_messageInfo_ComplicatedMockResource proto.InternalMessageInfo

func (m *ComplicatedMockResource) GetStatus() *core.Status {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *ComplicatedMockResource) GetMetadata() *core.Metadata {
	if m != nil {
		return m.Metadata
	}
	return nil
}

func (m *ComplicatedMockResource) GetNestedOneOf() *NestedOneOf {
	if m != nil {
		return m.NestedOneOf
	}
	return nil
}

func (m *ComplicatedMockResource) GetName() string {
	if m != nil {
		return m.Name
	}
	return ""
}

func (m *ComplicatedMockResource) GetSslConfig() *LocalSslConfig {
	if m != nil {
		return m.SslConfig
	}
	return nil
}

// message comment
type NestedOneOf struct {
	// oneof comment
	//
	// Types that are valid to be assigned to Option:
	//	*NestedOneOf_OptionA
	//	*NestedOneOf_OptionB
	Option               isNestedOneOf_Option `protobuf_oneof:"option"`
	Field                string               `protobuf:"bytes,3,opt,name=field,proto3" json:"field,omitempty"`
	MultipleA            []string             `protobuf:"bytes,4,rep,name=multiple_a,json=multipleA,proto3" json:"multiple_a,omitempty"`
	MultipleB            []string             `protobuf:"bytes,5,rep,name=multiple_b,json=multipleB,proto3" json:"multiple_b,omitempty"`
	MultipleC            []string             `protobuf:"bytes,6,rep,name=multiple_c,json=multipleC,proto3" json:"multiple_c,omitempty"`
	Choice               bool                 `protobuf:"varint,7,opt,name=choice,proto3" json:"choice,omitempty"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *NestedOneOf) Reset()         { *m = NestedOneOf{} }
func (m *NestedOneOf) String() string { return proto.CompactTextString(m) }
func (*NestedOneOf) ProtoMessage()    {}
func (*NestedOneOf) Descriptor() ([]byte, []int) {
	return fileDescriptor_c6e0c163912ae244, []int{1}
}
func (m *NestedOneOf) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_NestedOneOf.Unmarshal(m, b)
}
func (m *NestedOneOf) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_NestedOneOf.Marshal(b, m, deterministic)
}
func (m *NestedOneOf) XXX_Merge(src proto.Message) {
	xxx_messageInfo_NestedOneOf.Merge(m, src)
}
func (m *NestedOneOf) XXX_Size() int {
	return xxx_messageInfo_NestedOneOf.Size(m)
}
func (m *NestedOneOf) XXX_DiscardUnknown() {
	xxx_messageInfo_NestedOneOf.DiscardUnknown(m)
}

var xxx_messageInfo_NestedOneOf proto.InternalMessageInfo

type isNestedOneOf_Option interface {
	isNestedOneOf_Option()
}

type NestedOneOf_OptionA struct {
	OptionA string `protobuf:"bytes,1,opt,name=option_a,json=optionA,proto3,oneof" json:"option_a,omitempty"`
}
type NestedOneOf_OptionB struct {
	OptionB string `protobuf:"bytes,2,opt,name=option_b,json=optionB,proto3,oneof" json:"option_b,omitempty"`
}

func (*NestedOneOf_OptionA) isNestedOneOf_Option() {}
func (*NestedOneOf_OptionB) isNestedOneOf_Option() {}

func (m *NestedOneOf) GetOption() isNestedOneOf_Option {
	if m != nil {
		return m.Option
	}
	return nil
}

func (m *NestedOneOf) GetOptionA() string {
	if x, ok := m.GetOption().(*NestedOneOf_OptionA); ok {
		return x.OptionA
	}
	return ""
}

func (m *NestedOneOf) GetOptionB() string {
	if x, ok := m.GetOption().(*NestedOneOf_OptionB); ok {
		return x.OptionB
	}
	return ""
}

func (m *NestedOneOf) GetField() string {
	if m != nil {
		return m.Field
	}
	return ""
}

func (m *NestedOneOf) GetMultipleA() []string {
	if m != nil {
		return m.MultipleA
	}
	return nil
}

func (m *NestedOneOf) GetMultipleB() []string {
	if m != nil {
		return m.MultipleB
	}
	return nil
}

func (m *NestedOneOf) GetMultipleC() []string {
	if m != nil {
		return m.MultipleC
	}
	return nil
}

func (m *NestedOneOf) GetChoice() bool {
	if m != nil {
		return m.Choice
	}
	return false
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*NestedOneOf) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*NestedOneOf_OptionA)(nil),
		(*NestedOneOf_OptionB)(nil),
	}
}

type LocalSslConfig struct {
	// with comments
	//
	// Types that are valid to be assigned to SslConfigOptions:
	//	*LocalSslConfig_A
	//	*LocalSslConfig_B
	SslConfigOptions isLocalSslConfig_SslConfigOptions `protobuf_oneof:"ssl_config_options"`
	// optional. the SNI domains that should be considered for TLS connections
	SniDomains []string `protobuf:"bytes,3,rep,name=sni_domains,json=sniDomains,proto3" json:"sni_domains,omitempty"`
	// Verify that the Subject Alternative Name in the peer certificate is one of the specified values.
	// note that a root_ca must be provided if this option is used.
	VerifySubjectAltName []string `protobuf:"bytes,5,rep,name=verify_subject_alt_name,json=verifySubjectAltName,proto3" json:"verify_subject_alt_name,omitempty"`
	// Set Application Level Protocol Negotiation
	// If empty, defaults to ["h2", "http/1.1"].
	AlpnProtocols []string `protobuf:"bytes,7,rep,name=alpn_protocols,json=alpnProtocols,proto3" json:"alpn_protocols,omitempty"`
	// If the SSL config has the ca.crt (root CA) provided, Gloo uses it to perform mTLS by default.
	// Set oneWayTls to true to disable mTLS in favor of server-only TLS (one-way TLS), even if Gloo has the root CA.
	// Defaults to false.
	OneWayTls            bool     `protobuf:"varint,8,opt,name=one_way_tls,json=oneWayTls,proto3" json:"one_way_tls,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *LocalSslConfig) Reset()         { *m = LocalSslConfig{} }
func (m *LocalSslConfig) String() string { return proto.CompactTextString(m) }
func (*LocalSslConfig) ProtoMessage()    {}
func (*LocalSslConfig) Descriptor() ([]byte, []int) {
	return fileDescriptor_c6e0c163912ae244, []int{2}
}
func (m *LocalSslConfig) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_LocalSslConfig.Unmarshal(m, b)
}
func (m *LocalSslConfig) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_LocalSslConfig.Marshal(b, m, deterministic)
}
func (m *LocalSslConfig) XXX_Merge(src proto.Message) {
	xxx_messageInfo_LocalSslConfig.Merge(m, src)
}
func (m *LocalSslConfig) XXX_Size() int {
	return xxx_messageInfo_LocalSslConfig.Size(m)
}
func (m *LocalSslConfig) XXX_DiscardUnknown() {
	xxx_messageInfo_LocalSslConfig.DiscardUnknown(m)
}

var xxx_messageInfo_LocalSslConfig proto.InternalMessageInfo

type isLocalSslConfig_SslConfigOptions interface {
	isLocalSslConfig_SslConfigOptions()
}

type LocalSslConfig_A struct {
	A string `protobuf:"bytes,1,opt,name=a,proto3,oneof" json:"a,omitempty"`
}
type LocalSslConfig_B struct {
	B string `protobuf:"bytes,2,opt,name=b,proto3,oneof" json:"b,omitempty"`
}

func (*LocalSslConfig_A) isLocalSslConfig_SslConfigOptions() {}
func (*LocalSslConfig_B) isLocalSslConfig_SslConfigOptions() {}

func (m *LocalSslConfig) GetSslConfigOptions() isLocalSslConfig_SslConfigOptions {
	if m != nil {
		return m.SslConfigOptions
	}
	return nil
}

func (m *LocalSslConfig) GetA() string {
	if x, ok := m.GetSslConfigOptions().(*LocalSslConfig_A); ok {
		return x.A
	}
	return ""
}

func (m *LocalSslConfig) GetB() string {
	if x, ok := m.GetSslConfigOptions().(*LocalSslConfig_B); ok {
		return x.B
	}
	return ""
}

func (m *LocalSslConfig) GetSniDomains() []string {
	if m != nil {
		return m.SniDomains
	}
	return nil
}

func (m *LocalSslConfig) GetVerifySubjectAltName() []string {
	if m != nil {
		return m.VerifySubjectAltName
	}
	return nil
}

func (m *LocalSslConfig) GetAlpnProtocols() []string {
	if m != nil {
		return m.AlpnProtocols
	}
	return nil
}

func (m *LocalSslConfig) GetOneWayTls() bool {
	if m != nil {
		return m.OneWayTls
	}
	return false
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*LocalSslConfig) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*LocalSslConfig_A)(nil),
		(*LocalSslConfig_B)(nil),
	}
}

func init() {
	proto.RegisterType((*ComplicatedMockResource)(nil), "testing.solo.io.ComplicatedMockResource")
	proto.RegisterType((*NestedOneOf)(nil), "testing.solo.io.NestedOneOf")
	proto.RegisterType((*LocalSslConfig)(nil), "testing.solo.io.LocalSslConfig")
}

func init() {
	proto.RegisterFile("github.com/solo-io/solo-kit/test/mocks/api/v1/complicated_mock_resources.proto", fileDescriptor_c6e0c163912ae244)
}

var fileDescriptor_c6e0c163912ae244 = []byte{
	// 579 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x54, 0x4f, 0x4f, 0x1b, 0x3f,
	0x10, 0xfd, 0x2d, 0xd9, 0x84, 0xec, 0x44, 0xf0, 0x93, 0xac, 0x08, 0x2c, 0xfa, 0x87, 0x08, 0xa9,
	0x6a, 0x7a, 0xe8, 0xae, 0x00, 0xf5, 0xd2, 0x43, 0x55, 0x42, 0x0f, 0x3d, 0x14, 0xa8, 0x96, 0x4a,
	0x95, 0x7a, 0xb1, 0x1c, 0xc7, 0x01, 0x17, 0xaf, 0xbd, 0x5a, 0x3b, 0x94, 0x5c, 0xf9, 0x54, 0x3d,
	0xf2, 0x41, 0x7a, 0xed, 0x07, 0xe0, 0xc0, 0xbd, 0x5a, 0x7b, 0x97, 0x64, 0x5b, 0xa9, 0xa2, 0xa7,
	0xf5, 0xcc, 0x9b, 0x37, 0xe3, 0x37, 0x7e, 0x5a, 0x38, 0x3e, 0x13, 0xf6, 0x7c, 0x36, 0x8e, 0x99,
	0xce, 0x12, 0xa3, 0xa5, 0x7e, 0x29, 0xb4, 0xff, 0x5e, 0x08, 0x9b, 0x58, 0x6e, 0x6c, 0x92, 0x69,
	0x76, 0x61, 0x12, 0x9a, 0x8b, 0xe4, 0x72, 0x37, 0x61, 0x3a, 0xcb, 0xa5, 0x60, 0xd4, 0xf2, 0x09,
	0x29, 0x01, 0x52, 0x70, 0xa3, 0x67, 0x05, 0xe3, 0x26, 0xce, 0x0b, 0x6d, 0x35, 0xfa, 0xbf, 0xe4,
	0x08, 0x75, 0x16, 0x97, 0x4d, 0x62, 0xa1, 0xb7, 0x10, 0xbf, 0xb2, 0x0e, 0x4a, 0xf8, 0x95, 0xf5,
	0x45, 0x5b, 0xbb, 0x7f, 0x1b, 0x5a, 0x4d, 0xca, 0xb8, 0xa5, 0x13, 0x6a, 0x69, 0x45, 0x49, 0x1e,
	0x40, 0x31, 0x96, 0xda, 0x99, 0xf9, 0x87, 0x19, 0x75, 0xec, 0x29, 0x3b, 0x37, 0x2b, 0xb0, 0x79,
	0xb8, 0x10, 0x78, 0xa4, 0xd9, 0x45, 0x5a, 0xc9, 0x43, 0x7b, 0xd0, 0xf1, 0xed, 0x71, 0x30, 0x08,
	0x86, 0xbd, 0xbd, 0x7e, 0xcc, 0x74, 0xc1, 0x6b, 0x95, 0xf1, 0xa9, 0xc3, 0x46, 0xe1, 0xf7, 0xbb,
	0x30, 0x48, 0xab, 0x4a, 0xb4, 0x07, 0xdd, 0x5a, 0x05, 0x5e, 0x71, 0xac, 0x8d, 0x26, 0xeb, 0xa8,
	0x42, 0xd3, 0xfb, 0x3a, 0xf4, 0x16, 0xd6, 0x14, 0x37, 0xe5, 0x7a, 0xb5, 0xe2, 0x44, 0x4f, 0x71,
	0xcb, 0x11, 0x1f, 0xc7, 0xbf, 0xed, 0x35, 0x3e, 0x76, 0x55, 0x27, 0x8a, 0x9f, 0x4c, 0xd3, 0x9e,
	0x5a, 0x04, 0x08, 0x41, 0xa8, 0x68, 0xc6, 0x71, 0x38, 0x08, 0x86, 0x51, 0xea, 0xce, 0xe8, 0x0d,
	0x80, 0x31, 0x92, 0x30, 0xad, 0xa6, 0xe2, 0x0c, 0xb7, 0x5d, 0xcb, 0xed, 0x3f, 0x5a, 0x7e, 0xd0,
	0x8c, 0xca, 0x53, 0x23, 0x0f, 0x5d, 0x59, 0x1a, 0x99, 0xfa, 0xf8, 0xfa, 0xf9, 0xf5, 0x6d, 0xd8,
	0x86, 0x16, 0xcb, 0x8a, 0xeb, 0xdb, 0x70, 0x0b, 0xe1, 0x25, 0x1b, 0x94, 0x2e, 0xb8, 0x37, 0xc1,
	0xce, 0x8f, 0x00, 0x7a, 0x4b, 0x37, 0x43, 0x8f, 0xa0, 0xab, 0x73, 0x2b, 0xb4, 0x22, 0xd4, 0x2d,
	0x2e, 0x7a, 0xff, 0x5f, 0xba, 0xea, 0x33, 0x07, 0x4b, 0xe0, 0xd8, 0xed, 0x67, 0x09, 0x1c, 0xa1,
	0x3e, 0xb4, 0xa7, 0x82, 0xcb, 0x89, 0x5b, 0x40, 0x94, 0xfa, 0x00, 0x3d, 0x01, 0xc8, 0x66, 0xd2,
	0x8a, 0x5c, 0x72, 0x42, 0x71, 0x38, 0x68, 0x0d, 0xa3, 0x34, 0xaa, 0x33, 0x07, 0x0d, 0x78, 0x8c,
	0xdb, 0x4d, 0x78, 0xd4, 0x80, 0x19, 0xee, 0x34, 0xe1, 0x43, 0xb4, 0x01, 0x1d, 0x76, 0xae, 0x05,
	0xe3, 0x78, 0x75, 0x10, 0x0c, 0xbb, 0x69, 0x15, 0x8d, 0xba, 0xd0, 0xf1, 0xb7, 0xda, 0xf9, 0x19,
	0xc0, 0x7a, 0x73, 0x4b, 0x68, 0x1d, 0x82, 0x85, 0xb4, 0x80, 0x96, 0xf1, 0x42, 0x4d, 0x30, 0x46,
	0xdb, 0xd0, 0x33, 0x4a, 0x90, 0x89, 0xce, 0xa8, 0x50, 0x06, 0xb7, 0xdc, 0x50, 0x30, 0x4a, 0xbc,
	0xf3, 0x19, 0xf4, 0x0a, 0x36, 0x2f, 0x79, 0x21, 0xa6, 0x73, 0x62, 0x66, 0xe3, 0xaf, 0x9c, 0x59,
	0x42, 0xa5, 0x25, 0xee, 0x09, 0xbd, 0x80, 0xbe, 0x87, 0x4f, 0x3d, 0x7a, 0x20, 0xed, 0x71, 0xf9,
	0xa4, 0xcf, 0x60, 0x9d, 0xca, 0x5c, 0x11, 0x67, 0x5d, 0xa6, 0xa5, 0xc1, 0xab, 0xae, 0x7a, 0xad,
	0xcc, 0x7e, 0xac, 0x93, 0xe8, 0x29, 0xf4, 0x4a, 0x23, 0x7d, 0xa3, 0x73, 0x62, 0xa5, 0xc1, 0x5d,
	0x27, 0x2c, 0xd2, 0x8a, 0x7f, 0xa6, 0xf3, 0x4f, 0xd2, 0x8c, 0xfa, 0x80, 0x16, 0xce, 0x20, 0x5e,
	0xa6, 0x19, 0xed, 0x97, 0x3e, 0xbe, 0xb9, 0x0b, 0x83, 0x2f, 0x2f, 0x1e, 0xf8, 0x7f, 0xb8, 0xdc,
	0x1d, 0x77, 0xdc, 0x55, 0xf6, 0x7f, 0x05, 0x00, 0x00, 0xff, 0xff, 0x4c, 0xf0, 0x3f, 0x49, 0x53,
	0x04, 0x00, 0x00,
}
