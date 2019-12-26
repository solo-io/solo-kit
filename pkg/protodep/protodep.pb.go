// Code generated by protoc-gen-go. DO NOT EDIT.
// source: protodep.proto

package protodep

import (
	fmt "fmt"
	math "math"

	_ "github.com/envoyproxy/protoc-gen-validate/validate"
	proto "github.com/golang/protobuf/proto"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion3 // please upgrade the proto package

type Config struct {
	Local                *Local    `protobuf:"bytes,1,opt,name=local,proto3" json:"local,omitempty"`
	Imports              []*Import `protobuf:"bytes,2,rep,name=imports,proto3" json:"imports,omitempty"`
	XXX_NoUnkeyedLiteral struct{}  `json:"-"`
	XXX_unrecognized     []byte    `json:"-"`
	XXX_sizecache        int32     `json:"-"`
}

func (m *Config) Reset()         { *m = Config{} }
func (m *Config) String() string { return proto.CompactTextString(m) }
func (*Config) ProtoMessage()    {}
func (*Config) Descriptor() ([]byte, []int) {
	return fileDescriptor_7fec50c21b53b759, []int{0}
}

func (m *Config) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Config.Unmarshal(m, b)
}
func (m *Config) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Config.Marshal(b, m, deterministic)
}
func (m *Config) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Config.Merge(m, src)
}
func (m *Config) XXX_Size() int {
	return xxx_messageInfo_Config.Size(m)
}
func (m *Config) XXX_DiscardUnknown() {
	xxx_messageInfo_Config.DiscardUnknown(m)
}

var xxx_messageInfo_Config proto.InternalMessageInfo

func (m *Config) GetLocal() *Local {
	if m != nil {
		return m.Local
	}
	return nil
}

func (m *Config) GetImports() []*Import {
	if m != nil {
		return m.Imports
	}
	return nil
}

type Import struct {
	// Types that are valid to be assigned to ImportType:
	//	*Import_GoMod
	ImportType           isImport_ImportType `protobuf_oneof:"ImportType"`
	XXX_NoUnkeyedLiteral struct{}            `json:"-"`
	XXX_unrecognized     []byte              `json:"-"`
	XXX_sizecache        int32               `json:"-"`
}

func (m *Import) Reset()         { *m = Import{} }
func (m *Import) String() string { return proto.CompactTextString(m) }
func (*Import) ProtoMessage()    {}
func (*Import) Descriptor() ([]byte, []int) {
	return fileDescriptor_7fec50c21b53b759, []int{1}
}

func (m *Import) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Import.Unmarshal(m, b)
}
func (m *Import) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Import.Marshal(b, m, deterministic)
}
func (m *Import) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Import.Merge(m, src)
}
func (m *Import) XXX_Size() int {
	return xxx_messageInfo_Import.Size(m)
}
func (m *Import) XXX_DiscardUnknown() {
	xxx_messageInfo_Import.DiscardUnknown(m)
}

var xxx_messageInfo_Import proto.InternalMessageInfo

type isImport_ImportType interface {
	isImport_ImportType()
}

type Import_GoMod struct {
	GoMod *GoMod `protobuf:"bytes,2,opt,name=go_mod,json=goMod,proto3,oneof"`
}

func (*Import_GoMod) isImport_ImportType() {}

func (m *Import) GetImportType() isImport_ImportType {
	if m != nil {
		return m.ImportType
	}
	return nil
}

func (m *Import) GetGoMod() *GoMod {
	if x, ok := m.GetImportType().(*Import_GoMod); ok {
		return x.GoMod
	}
	return nil
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*Import) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*Import_GoMod)(nil),
	}
}

type Local struct {
	Patterns             []string `protobuf:"bytes,1,rep,name=patterns,proto3" json:"patterns,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Local) Reset()         { *m = Local{} }
func (m *Local) String() string { return proto.CompactTextString(m) }
func (*Local) ProtoMessage()    {}
func (*Local) Descriptor() ([]byte, []int) {
	return fileDescriptor_7fec50c21b53b759, []int{2}
}

func (m *Local) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Local.Unmarshal(m, b)
}
func (m *Local) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Local.Marshal(b, m, deterministic)
}
func (m *Local) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Local.Merge(m, src)
}
func (m *Local) XXX_Size() int {
	return xxx_messageInfo_Local.Size(m)
}
func (m *Local) XXX_DiscardUnknown() {
	xxx_messageInfo_Local.DiscardUnknown(m)
}

var xxx_messageInfo_Local proto.InternalMessageInfo

func (m *Local) GetPatterns() []string {
	if m != nil {
		return m.Patterns
	}
	return nil
}

type GoMod struct {
	Patterns             []string `protobuf:"bytes,1,rep,name=patterns,proto3" json:"patterns,omitempty"`
	Package              string   `protobuf:"bytes,2,opt,name=package,proto3" json:"package,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GoMod) Reset()         { *m = GoMod{} }
func (m *GoMod) String() string { return proto.CompactTextString(m) }
func (*GoMod) ProtoMessage()    {}
func (*GoMod) Descriptor() ([]byte, []int) {
	return fileDescriptor_7fec50c21b53b759, []int{3}
}

func (m *GoMod) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GoMod.Unmarshal(m, b)
}
func (m *GoMod) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GoMod.Marshal(b, m, deterministic)
}
func (m *GoMod) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GoMod.Merge(m, src)
}
func (m *GoMod) XXX_Size() int {
	return xxx_messageInfo_GoMod.Size(m)
}
func (m *GoMod) XXX_DiscardUnknown() {
	xxx_messageInfo_GoMod.DiscardUnknown(m)
}

var xxx_messageInfo_GoMod proto.InternalMessageInfo

func (m *GoMod) GetPatterns() []string {
	if m != nil {
		return m.Patterns
	}
	return nil
}

func (m *GoMod) GetPackage() string {
	if m != nil {
		return m.Package
	}
	return ""
}

type Git struct {
	Owner                string   `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
	Repo                 string   `protobuf:"bytes,2,opt,name=repo,proto3" json:"repo,omitempty"`
	Revision             string   `protobuf:"bytes,3,opt,name=revision,proto3" json:"revision,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *Git) Reset()         { *m = Git{} }
func (m *Git) String() string { return proto.CompactTextString(m) }
func (*Git) ProtoMessage()    {}
func (*Git) Descriptor() ([]byte, []int) {
	return fileDescriptor_7fec50c21b53b759, []int{4}
}

func (m *Git) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_Git.Unmarshal(m, b)
}
func (m *Git) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_Git.Marshal(b, m, deterministic)
}
func (m *Git) XXX_Merge(src proto.Message) {
	xxx_messageInfo_Git.Merge(m, src)
}
func (m *Git) XXX_Size() int {
	return xxx_messageInfo_Git.Size(m)
}
func (m *Git) XXX_DiscardUnknown() {
	xxx_messageInfo_Git.DiscardUnknown(m)
}

var xxx_messageInfo_Git proto.InternalMessageInfo

func (m *Git) GetOwner() string {
	if m != nil {
		return m.Owner
	}
	return ""
}

func (m *Git) GetRepo() string {
	if m != nil {
		return m.Repo
	}
	return ""
}

func (m *Git) GetRevision() string {
	if m != nil {
		return m.Revision
	}
	return ""
}

func init() {
	proto.RegisterType((*Config)(nil), "protodep.Config")
	proto.RegisterType((*Import)(nil), "protodep.Import")
	proto.RegisterType((*Local)(nil), "protodep.Local")
	proto.RegisterType((*GoMod)(nil), "protodep.GoMod")
	proto.RegisterType((*Git)(nil), "protodep.Git")
}

func init() { proto.RegisterFile("protodep.proto", fileDescriptor_7fec50c21b53b759) }

var fileDescriptor_7fec50c21b53b759 = []byte{
	// 286 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x8f, 0x4f, 0x4b, 0x03, 0x31,
	0x10, 0xc5, 0x6d, 0xd7, 0xdd, 0xb6, 0x53, 0x51, 0x09, 0x82, 0xa1, 0xa7, 0xb2, 0x22, 0x2c, 0x05,
	0xbb, 0x50, 0xcf, 0x82, 0xd4, 0x43, 0x15, 0xf5, 0x12, 0x3c, 0xe9, 0x41, 0xd2, 0xdd, 0x18, 0xc3,
	0xfe, 0x99, 0x90, 0x8d, 0x15, 0xbf, 0xbd, 0x6c, 0xe2, 0xae, 0xe0, 0xc1, 0x53, 0xde, 0x6f, 0xe6,
	0xcd, 0x9b, 0x0c, 0x1c, 0x6a, 0x83, 0x16, 0x73, 0xa1, 0x97, 0x4e, 0x90, 0x71, 0xc7, 0xb3, 0xd3,
	0x1d, 0x2f, 0x55, 0xce, 0xad, 0x48, 0x3b, 0xe1, 0x2d, 0xf1, 0x0b, 0x44, 0x37, 0x58, 0xbf, 0x29,
	0x49, 0xce, 0x21, 0x2c, 0x31, 0xe3, 0x25, 0x1d, 0xcc, 0x07, 0xc9, 0x74, 0x75, 0xb4, 0xec, 0xc3,
	0x1e, 0xda, 0x32, 0xf3, 0x5d, 0xb2, 0x80, 0x91, 0xaa, 0x34, 0x1a, 0xdb, 0xd0, 0xe1, 0x3c, 0x48,
	0xa6, 0xab, 0xe3, 0x5f, 0xe3, 0x9d, 0x6b, 0xb0, 0xce, 0x10, 0x5f, 0x43, 0xe4, 0x4b, 0x24, 0x81,
	0x48, 0xe2, 0x6b, 0x85, 0x39, 0x1d, 0xfe, 0x4d, 0xdf, 0xe0, 0x23, 0xe6, 0xb7, 0x7b, 0x2c, 0x94,
	0xad, 0x58, 0x1f, 0x00, 0xf8, 0x99, 0xa7, 0x2f, 0x2d, 0xe2, 0x33, 0x08, 0xdd, 0x76, 0x32, 0x83,
	0xb1, 0xe6, 0xd6, 0x0a, 0x53, 0x37, 0x74, 0x30, 0x0f, 0x92, 0x09, 0xeb, 0x39, 0xbe, 0x82, 0xd0,
	0x85, 0xfc, 0x67, 0x22, 0x14, 0x46, 0x9a, 0x67, 0x05, 0x97, 0xc2, 0x7d, 0x61, 0xc2, 0x3a, 0x8c,
	0xef, 0x21, 0xd8, 0x28, 0x4b, 0x4e, 0x20, 0xc4, 0xcf, 0x5a, 0x18, 0x77, 0xff, 0x84, 0x79, 0x20,
	0x04, 0xf6, 0x8d, 0xd0, 0xf8, 0x33, 0xe3, 0x74, 0xbb, 0xc6, 0x88, 0x9d, 0x6a, 0x14, 0xd6, 0x34,
	0x70, 0xf5, 0x9e, 0xd7, 0x8b, 0xe7, 0x44, 0x2a, 0xfb, 0xfe, 0xb1, 0x5d, 0x66, 0x58, 0xa5, 0x0d,
	0x96, 0x78, 0xa1, 0xd0, 0xbf, 0x85, 0xb2, 0xa9, 0x2e, 0x64, 0xda, 0x5d, 0xbe, 0x8d, 0x9c, 0xba,
	0xfc, 0x0e, 0x00, 0x00, 0xff, 0xff, 0x36, 0xbd, 0x4a, 0x99, 0xb7, 0x01, 0x00, 0x00,
}
