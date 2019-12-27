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
	GoMod *GoModImport `protobuf:"bytes,2,opt,name=go_mod,json=goMod,proto3,oneof"`
}

func (*Import_GoMod) isImport_ImportType() {}

func (m *Import) GetImportType() isImport_ImportType {
	if m != nil {
		return m.ImportType
	}
	return nil
}

func (m *Import) GetGoMod() *GoModImport {
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

type GoModImport struct {
	Patterns             []string `protobuf:"bytes,1,rep,name=patterns,proto3" json:"patterns,omitempty"`
	Package              string   `protobuf:"bytes,2,opt,name=package,proto3" json:"package,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *GoModImport) Reset()         { *m = GoModImport{} }
func (m *GoModImport) String() string { return proto.CompactTextString(m) }
func (*GoModImport) ProtoMessage()    {}
func (*GoModImport) Descriptor() ([]byte, []int) {
	return fileDescriptor_7fec50c21b53b759, []int{3}
}

func (m *GoModImport) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GoModImport.Unmarshal(m, b)
}
func (m *GoModImport) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GoModImport.Marshal(b, m, deterministic)
}
func (m *GoModImport) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GoModImport.Merge(m, src)
}
func (m *GoModImport) XXX_Size() int {
	return xxx_messageInfo_GoModImport.Size(m)
}
func (m *GoModImport) XXX_DiscardUnknown() {
	xxx_messageInfo_GoModImport.DiscardUnknown(m)
}

var xxx_messageInfo_GoModImport proto.InternalMessageInfo

func (m *GoModImport) GetPatterns() []string {
	if m != nil {
		return m.Patterns
	}
	return nil
}

func (m *GoModImport) GetPackage() string {
	if m != nil {
		return m.Package
	}
	return ""
}

type GitImport struct {
	Owner string `protobuf:"bytes,1,opt,name=owner,proto3" json:"owner,omitempty"`
	Repo  string `protobuf:"bytes,2,opt,name=repo,proto3" json:"repo,omitempty"`
	// will default to master, therefore can be left empty
	//
	// Types that are valid to be assigned to Revision:
	//	*GitImport_Sha
	//	*GitImport_Tag
	Revision             isGitImport_Revision `protobuf_oneof:"Revision"`
	XXX_NoUnkeyedLiteral struct{}             `json:"-"`
	XXX_unrecognized     []byte               `json:"-"`
	XXX_sizecache        int32                `json:"-"`
}

func (m *GitImport) Reset()         { *m = GitImport{} }
func (m *GitImport) String() string { return proto.CompactTextString(m) }
func (*GitImport) ProtoMessage()    {}
func (*GitImport) Descriptor() ([]byte, []int) {
	return fileDescriptor_7fec50c21b53b759, []int{4}
}

func (m *GitImport) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_GitImport.Unmarshal(m, b)
}
func (m *GitImport) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_GitImport.Marshal(b, m, deterministic)
}
func (m *GitImport) XXX_Merge(src proto.Message) {
	xxx_messageInfo_GitImport.Merge(m, src)
}
func (m *GitImport) XXX_Size() int {
	return xxx_messageInfo_GitImport.Size(m)
}
func (m *GitImport) XXX_DiscardUnknown() {
	xxx_messageInfo_GitImport.DiscardUnknown(m)
}

var xxx_messageInfo_GitImport proto.InternalMessageInfo

func (m *GitImport) GetOwner() string {
	if m != nil {
		return m.Owner
	}
	return ""
}

func (m *GitImport) GetRepo() string {
	if m != nil {
		return m.Repo
	}
	return ""
}

type isGitImport_Revision interface {
	isGitImport_Revision()
}

type GitImport_Sha struct {
	Sha string `protobuf:"bytes,3,opt,name=sha,proto3,oneof"`
}

type GitImport_Tag struct {
	Tag string `protobuf:"bytes,5,opt,name=tag,proto3,oneof"`
}

func (*GitImport_Sha) isGitImport_Revision() {}

func (*GitImport_Tag) isGitImport_Revision() {}

func (m *GitImport) GetRevision() isGitImport_Revision {
	if m != nil {
		return m.Revision
	}
	return nil
}

func (m *GitImport) GetSha() string {
	if x, ok := m.GetRevision().(*GitImport_Sha); ok {
		return x.Sha
	}
	return ""
}

func (m *GitImport) GetTag() string {
	if x, ok := m.GetRevision().(*GitImport_Tag); ok {
		return x.Tag
	}
	return ""
}

// XXX_OneofWrappers is for the internal use of the proto package.
func (*GitImport) XXX_OneofWrappers() []interface{} {
	return []interface{}{
		(*GitImport_Sha)(nil),
		(*GitImport_Tag)(nil),
	}
}

func init() {
	proto.RegisterType((*Config)(nil), "protodep.Config")
	proto.RegisterType((*Import)(nil), "protodep.Import")
	proto.RegisterType((*Local)(nil), "protodep.Local")
	proto.RegisterType((*GoModImport)(nil), "protodep.GoModImport")
	proto.RegisterType((*GitImport)(nil), "protodep.GitImport")
}

func init() { proto.RegisterFile("protodep.proto", fileDescriptor_7fec50c21b53b759) }

var fileDescriptor_7fec50c21b53b759 = []byte{
	// 350 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x64, 0x90, 0xcd, 0x6a, 0xdb, 0x40,
	0x14, 0x85, 0x2d, 0xc9, 0x92, 0xa5, 0x6b, 0x68, 0xcd, 0x94, 0x52, 0x61, 0x28, 0x35, 0x32, 0x2d,
	0xc2, 0x50, 0x09, 0xdc, 0x37, 0x50, 0x17, 0x76, 0xa1, 0xde, 0x88, 0xac, 0x9c, 0x45, 0x18, 0x5b,
	0x93, 0xf1, 0x60, 0x59, 0x77, 0x90, 0x26, 0x0e, 0x59, 0xe5, 0x1d, 0xf2, 0x34, 0x21, 0x2b, 0xbf,
	0x8e, 0xdf, 0x22, 0x68, 0x14, 0xd9, 0xf9, 0x59, 0xcd, 0x3d, 0xe7, 0x7c, 0x73, 0xb8, 0x5c, 0xf8,
	0x24, 0x4b, 0x54, 0x98, 0x31, 0x19, 0xe9, 0x81, 0xb8, 0xad, 0x1e, 0x7e, 0xdb, 0xd3, 0x5c, 0x64,
	0x54, 0xb1, 0xb8, 0x1d, 0x1a, 0x24, 0xb8, 0x04, 0xe7, 0x2f, 0x16, 0xd7, 0x82, 0x93, 0x9f, 0x60,
	0xe7, 0xb8, 0xa6, 0xb9, 0x6f, 0x8c, 0x8c, 0xb0, 0x3f, 0xfd, 0x1c, 0x9d, 0xca, 0xfe, 0xd7, 0x76,
	0xda, 0xa4, 0x64, 0x02, 0x3d, 0xb1, 0x93, 0x58, 0xaa, 0xca, 0x37, 0x47, 0x56, 0xd8, 0x9f, 0x0e,
	0xce, 0xe0, 0x3f, 0x1d, 0xa4, 0x2d, 0x10, 0x2c, 0xc0, 0x69, 0x2c, 0x12, 0x81, 0xc3, 0xf1, 0x6a,
	0x87, 0x99, 0x6f, 0xea, 0xf6, 0xaf, 0xe7, 0x4f, 0x33, 0x5c, 0x60, 0xd6, 0x60, 0xf3, 0x4e, 0x6a,
	0xf3, 0x5a, 0x26, 0x5f, 0x00, 0x1a, 0xeb, 0xe2, 0x4e, 0x32, 0x62, 0x3f, 0x1e, 0x0f, 0x96, 0x11,
	0x8c, 0xc1, 0xd6, 0xab, 0x90, 0x21, 0xb8, 0x92, 0x2a, 0xc5, 0xca, 0xa2, 0xf2, 0x8d, 0x91, 0x15,
	0x7a, 0xe9, 0x49, 0x07, 0x4b, 0xe8, 0xbf, 0x6a, 0x24, 0xbf, 0xde, 0xa3, 0x09, 0x3c, 0x1d, 0x0f,
	0x96, 0xfd, 0x60, 0x98, 0xae, 0x71, 0xfe, 0x46, 0xc6, 0xd0, 0x93, 0x74, 0xbd, 0xa5, 0x9c, 0xe9,
	0x0d, 0xbd, 0xc4, 0xab, 0xb1, 0x6e, 0x69, 0x0e, 0x8c, 0xb4, 0x4d, 0x82, 0x7b, 0xf0, 0x66, 0x42,
	0xbd, 0x34, 0xff, 0x00, 0x1b, 0x6f, 0x0b, 0x56, 0xea, 0x7b, 0xbd, 0xe1, 0x1b, 0x9f, 0x7c, 0x87,
	0x6e, 0xc9, 0x24, 0x7e, 0xec, 0xd3, 0x36, 0x21, 0x60, 0x55, 0x1b, 0xea, 0x5b, 0x75, 0x3a, 0xef,
	0xa4, 0xb5, 0xa8, 0x3d, 0x45, 0xb9, 0x6f, 0xb7, 0x9e, 0xa2, 0x3c, 0x01, 0x70, 0x53, 0xb6, 0x17,
	0x95, 0xc0, 0x22, 0x99, 0x2c, 0x43, 0x2e, 0xd4, 0xe6, 0x66, 0x15, 0xad, 0x71, 0x17, 0x57, 0x98,
	0xe3, 0x6f, 0x81, 0xcd, 0xbb, 0x15, 0x2a, 0x96, 0x5b, 0x1e, 0xb7, 0x77, 0x5d, 0x39, 0x7a, 0xfa,
	0xf3, 0x1c, 0x00, 0x00, 0xff, 0xff, 0xa7, 0x8f, 0x0e, 0x0a, 0x15, 0x02, 0x00, 0x00,
}
