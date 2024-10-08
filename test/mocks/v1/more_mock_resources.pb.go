// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.33.0
// 	protoc        v3.6.1
// source: github.com/solo-io/solo-kit/test/mocks/api/v1/more_mock_resources.proto

package v1

import (
	reflect "reflect"
	sync "sync"

	_ "github.com/solo-io/protoc-gen-ext/extproto"
	core "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Description of the AnotherMockResource
type AnotherMockResource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metadata           *core.Metadata           `protobuf:"bytes,1,opt,name=metadata,proto3" json:"metadata,omitempty"`
	NamespacedStatuses *core.NamespacedStatuses `protobuf:"bytes,7,opt,name=namespaced_statuses,json=namespacedStatuses,proto3" json:"namespaced_statuses,omitempty"`
	// comments that go above the basic field in our docs
	BasicField string `protobuf:"bytes,2,opt,name=basic_field,json=basicField,proto3" json:"basic_field,omitempty"`
}

func (x *AnotherMockResource) Reset() {
	*x = AnotherMockResource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *AnotherMockResource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*AnotherMockResource) ProtoMessage() {}

func (x *AnotherMockResource) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use AnotherMockResource.ProtoReflect.Descriptor instead.
func (*AnotherMockResource) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_rawDescGZIP(), []int{0}
}

func (x *AnotherMockResource) GetMetadata() *core.Metadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *AnotherMockResource) GetNamespacedStatuses() *core.NamespacedStatuses {
	if x != nil {
		return x.NamespacedStatuses
	}
	return nil
}

func (x *AnotherMockResource) GetBasicField() string {
	if x != nil {
		return x.BasicField
	}
	return ""
}

type ClusterResource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metadata           *core.Metadata           `protobuf:"bytes,1,opt,name=metadata,proto3" json:"metadata,omitempty"`
	NamespacedStatuses *core.NamespacedStatuses `protobuf:"bytes,7,opt,name=namespaced_statuses,json=namespacedStatuses,proto3" json:"namespaced_statuses,omitempty"`
	// comments that go above the basic field in our docs
	BasicField string `protobuf:"bytes,2,opt,name=basic_field,json=basicField,proto3" json:"basic_field,omitempty"`
}

func (x *ClusterResource) Reset() {
	*x = ClusterResource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ClusterResource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ClusterResource) ProtoMessage() {}

func (x *ClusterResource) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ClusterResource.ProtoReflect.Descriptor instead.
func (*ClusterResource) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_rawDescGZIP(), []int{1}
}

func (x *ClusterResource) GetMetadata() *core.Metadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *ClusterResource) GetNamespacedStatuses() *core.NamespacedStatuses {
	if x != nil {
		return x.NamespacedStatuses
	}
	return nil
}

func (x *ClusterResource) GetBasicField() string {
	if x != nil {
		return x.BasicField
	}
	return ""
}

var File_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto protoreflect.FileDescriptor

var file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_rawDesc = []byte{
	0x0a, 0x47, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c,
	0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x74, 0x65,
	0x73, 0x74, 0x2f, 0x6d, 0x6f, 0x63, 0x6b, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f,
	0x6d, 0x6f, 0x72, 0x65, 0x5f, 0x6d, 0x6f, 0x63, 0x6b, 0x5f, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72,
	0x63, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x74, 0x65, 0x73, 0x74, 0x69,
	0x6e, 0x67, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x1a, 0x12, 0x65, 0x78, 0x74, 0x70,
	0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x78, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x31,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d,
	0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x31, 0x2f, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x1a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f,
	0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x61,
	0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x31, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73,
	0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f,
	0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xea, 0x01, 0x0a, 0x13, 0x41, 0x6e, 0x6f, 0x74, 0x68, 0x65,
	0x72, 0x4d, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x32, 0x0a,
	0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x16, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x4d,
	0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x12, 0x57, 0x0a, 0x13, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x64, 0x5f,
	0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20,
	0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x4e, 0x61,
	0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73,
	0x42, 0x04, 0xb8, 0xf5, 0x04, 0x01, 0x52, 0x12, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63,
	0x65, 0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x62, 0x61,
	0x73, 0x69, 0x63, 0x5f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0a, 0x62, 0x61, 0x73, 0x69, 0x63, 0x46, 0x69, 0x65, 0x6c, 0x64, 0x3a, 0x1f, 0x82, 0xf1, 0x04,
	0x1b, 0x0a, 0x03, 0x61, 0x6d, 0x72, 0x12, 0x14, 0x61, 0x6e, 0x6f, 0x74, 0x68, 0x65, 0x72, 0x6d,
	0x6f, 0x63, 0x6b, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x73, 0x4a, 0x04, 0x08, 0x06,
	0x10, 0x07, 0x22, 0xe4, 0x01, 0x0a, 0x0f, 0x43, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x52, 0x65,
	0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x32, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61,
	0x74, 0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e,
	0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61,
	0x52, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x57, 0x0a, 0x13, 0x6e, 0x61,
	0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x64, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65,
	0x73, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73,
	0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65,
	0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73, 0x42, 0x04, 0xb8, 0xf5, 0x04, 0x01, 0x52,
	0x12, 0x6e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x64, 0x53, 0x74, 0x61, 0x74, 0x75,
	0x73, 0x65, 0x73, 0x12, 0x1f, 0x0a, 0x0b, 0x62, 0x61, 0x73, 0x69, 0x63, 0x5f, 0x66, 0x69, 0x65,
	0x6c, 0x64, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x62, 0x61, 0x73, 0x69, 0x63, 0x46,
	0x69, 0x65, 0x6c, 0x64, 0x3a, 0x1d, 0x82, 0xf1, 0x04, 0x19, 0x0a, 0x03, 0x63, 0x6c, 0x72, 0x12,
	0x10, 0x63, 0x6c, 0x75, 0x73, 0x74, 0x65, 0x72, 0x72, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65,
	0x73, 0x18, 0x01, 0x4a, 0x04, 0x08, 0x06, 0x10, 0x07, 0x42, 0x33, 0xb8, 0xf5, 0x04, 0x01, 0xc0,
	0xf5, 0x04, 0x01, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f,
	0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74,
	0x2f, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x6d, 0x6f, 0x63, 0x6b, 0x73, 0x2f, 0x76, 0x31, 0x62, 0x06,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_rawDescOnce sync.Once
	file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_rawDescData = file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_rawDesc
)

func file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_rawDescGZIP() []byte {
	file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_rawDescOnce.Do(func() {
		file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_rawDescData)
	})
	return file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_rawDescData
}

var file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_goTypes = []interface{}{
	(*AnotherMockResource)(nil),     // 0: testing.solo.io.AnotherMockResource
	(*ClusterResource)(nil),         // 1: testing.solo.io.ClusterResource
	(*core.Metadata)(nil),           // 2: core.solo.io.Metadata
	(*core.NamespacedStatuses)(nil), // 3: core.solo.io.NamespacedStatuses
}
var file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_depIdxs = []int32{
	2, // 0: testing.solo.io.AnotherMockResource.metadata:type_name -> core.solo.io.Metadata
	3, // 1: testing.solo.io.AnotherMockResource.namespaced_statuses:type_name -> core.solo.io.NamespacedStatuses
	2, // 2: testing.solo.io.ClusterResource.metadata:type_name -> core.solo.io.Metadata
	3, // 3: testing.solo.io.ClusterResource.namespaced_statuses:type_name -> core.solo.io.NamespacedStatuses
	4, // [4:4] is the sub-list for method output_type
	4, // [4:4] is the sub-list for method input_type
	4, // [4:4] is the sub-list for extension type_name
	4, // [4:4] is the sub-list for extension extendee
	0, // [0:4] is the sub-list for field type_name
}

func init() { file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_init() }
func file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_init() {
	if File_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*AnotherMockResource); i {
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
		file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*ClusterResource); i {
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
			RawDescriptor: file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_goTypes,
		DependencyIndexes: file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_depIdxs,
		MessageInfos:      file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_msgTypes,
	}.Build()
	File_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto = out.File
	file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_rawDesc = nil
	file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_goTypes = nil
	file_github_com_solo_io_solo_kit_test_mocks_api_v1_more_mock_resources_proto_depIdxs = nil
}
