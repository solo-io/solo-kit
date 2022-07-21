// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.21.0
// 	protoc        v3.6.1
// source: solo-kit.proto

package core

import (
	reflect "reflect"
	sync "sync"

	proto "github.com/golang/protobuf/proto"
	descriptor "github.com/golang/protobuf/protoc-gen-go/descriptor"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

type Resource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	// becomes the kubernetes short name for the generated crd
	ShortName string `protobuf:"bytes,1,opt,name=short_name,json=shortName,proto3" json:"short_name,omitempty"`
	// becomes the kubernetes plural name for the generated crd
	PluralName string `protobuf:"bytes,2,opt,name=plural_name,json=pluralName,proto3" json:"plural_name,omitempty"`
	// the resource lives at the cluster level, namespace is ignored by the server
	ClusterScoped bool `protobuf:"varint,3,opt,name=cluster_scoped,json=clusterScoped,proto3" json:"cluster_scoped,omitempty"`
	// indicates whether documentation generation has to be skipped for the given resource, defaults to false
	SkipDocsGen bool `protobuf:"varint,4,opt,name=skip_docs_gen,json=skipDocsGen,proto3" json:"skip_docs_gen,omitempty"`
	// indicates whether annotations should be excluded from the resource's generated hash function.
	// if set to true, changes in annotations will not cause a new snapshot to be emitted
	SkipHashingAnnotations bool `protobuf:"varint,5,opt,name=skip_hashing_annotations,json=skipHashingAnnotations,proto3" json:"skip_hashing_annotations,omitempty"`
}

func (x *Resource) Reset() {
	*x = Resource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_solo_kit_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Resource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Resource) ProtoMessage() {}

func (x *Resource) ProtoReflect() protoreflect.Message {
	mi := &file_solo_kit_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Resource.ProtoReflect.Descriptor instead.
func (*Resource) Descriptor() ([]byte, []int) {
	return file_solo_kit_proto_rawDescGZIP(), []int{0}
}

func (x *Resource) GetShortName() string {
	if x != nil {
		return x.ShortName
	}
	return ""
}

func (x *Resource) GetPluralName() string {
	if x != nil {
		return x.PluralName
	}
	return ""
}

func (x *Resource) GetClusterScoped() bool {
	if x != nil {
		return x.ClusterScoped
	}
	return false
}

func (x *Resource) GetSkipDocsGen() bool {
	if x != nil {
		return x.SkipDocsGen
	}
	return false
}

func (x *Resource) GetSkipHashingAnnotations() bool {
	if x != nil {
		return x.SkipHashingAnnotations
	}
	return false
}

var file_solo_kit_proto_extTypes = []protoimpl.ExtensionInfo{
	{
		ExtendedType:  (*descriptor.MessageOptions)(nil),
		ExtensionType: (*Resource)(nil),
		Field:         10000,
		Name:          "core.solo.io.resource",
		Tag:           "bytes,10000,opt,name=resource",
		Filename:      "solo-kit.proto",
	},
}

// Extension fields to descriptor.MessageOptions.
var (
	// options for a message that's intended to become a solo-kit resource
	//
	// optional core.solo.io.Resource resource = 10000;
	E_Resource = &file_solo_kit_proto_extTypes[0]
)

var File_solo_kit_proto protoreflect.FileDescriptor

var file_solo_kit_proto_rawDesc = []byte{
	0x0a, 0x0e, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x12, 0x0c, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x1a, 0x20,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x64, 0x65, 0x73, 0x63, 0x72, 0x69, 0x70, 0x74, 0x6f, 0x72, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x22, 0xcf, 0x01, 0x0a, 0x08, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x1d, 0x0a,
	0x0a, 0x73, 0x68, 0x6f, 0x72, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x09, 0x73, 0x68, 0x6f, 0x72, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1f, 0x0a, 0x0b,
	0x70, 0x6c, 0x75, 0x72, 0x61, 0x6c, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x0a, 0x70, 0x6c, 0x75, 0x72, 0x61, 0x6c, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x25, 0x0a,
	0x0e, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x5f, 0x73, 0x63, 0x6f, 0x70, 0x65, 0x64, 0x18,
	0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0d, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x53, 0x63,
	0x6f, 0x70, 0x65, 0x64, 0x12, 0x22, 0x0a, 0x0d, 0x73, 0x6b, 0x69, 0x70, 0x5f, 0x64, 0x6f, 0x63,
	0x73, 0x5f, 0x67, 0x65, 0x6e, 0x18, 0x04, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0b, 0x73, 0x6b, 0x69,
	0x70, 0x44, 0x6f, 0x63, 0x73, 0x47, 0x65, 0x6e, 0x12, 0x38, 0x0a, 0x18, 0x73, 0x6b, 0x69, 0x70,
	0x5f, 0x68, 0x61, 0x73, 0x68, 0x69, 0x6e, 0x67, 0x5f, 0x61, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74,
	0x69, 0x6f, 0x6e, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x08, 0x52, 0x16, 0x73, 0x6b, 0x69, 0x70,
	0x48, 0x61, 0x73, 0x68, 0x69, 0x6e, 0x67, 0x41, 0x6e, 0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f,
	0x6e, 0x73, 0x3a, 0x54, 0x0a, 0x08, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x1f,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x4f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x18,
	0x90, 0x4e, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73, 0x6f,
	0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x52, 0x08,
	0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x42, 0x37, 0x5a, 0x35, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73,
	0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x70, 0x6b, 0x67, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x31, 0x2f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x2f, 0x63, 0x6f, 0x72,
	0x65, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_solo_kit_proto_rawDescOnce sync.Once
	file_solo_kit_proto_rawDescData = file_solo_kit_proto_rawDesc
)

func file_solo_kit_proto_rawDescGZIP() []byte {
	file_solo_kit_proto_rawDescOnce.Do(func() {
		file_solo_kit_proto_rawDescData = protoimpl.X.CompressGZIP(file_solo_kit_proto_rawDescData)
	})
	return file_solo_kit_proto_rawDescData
}

var file_solo_kit_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_solo_kit_proto_goTypes = []interface{}{
	(*Resource)(nil),                  // 0: core.solo.io.Resource
	(*descriptor.MessageOptions)(nil), // 1: google.protobuf.MessageOptions
}
var file_solo_kit_proto_depIdxs = []int32{
	1, // 0: core.solo.io.resource:extendee -> google.protobuf.MessageOptions
	0, // 1: core.solo.io.resource:type_name -> core.solo.io.Resource
	2, // [2:2] is the sub-list for method output_type
	2, // [2:2] is the sub-list for method input_type
	1, // [1:2] is the sub-list for extension type_name
	0, // [0:1] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_solo_kit_proto_init() }
func file_solo_kit_proto_init() {
	if File_solo_kit_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_solo_kit_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Resource); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_solo_kit_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 1,
			NumServices:   0,
		},
		GoTypes:           file_solo_kit_proto_goTypes,
		DependencyIndexes: file_solo_kit_proto_depIdxs,
		MessageInfos:      file_solo_kit_proto_msgTypes,
		ExtensionInfos:    file_solo_kit_proto_extTypes,
	}.Build()
	File_solo_kit_proto = out.File
	file_solo_kit_proto_rawDesc = nil
	file_solo_kit_proto_goTypes = nil
	file_solo_kit_proto_depIdxs = nil
}
