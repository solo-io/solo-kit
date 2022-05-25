// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.26.0
// 	protoc        v3.6.1
// source: github.com/solo-io/solo-kit/test/mocks/api/v2alpha1/mock_resources.proto

package v2alpha1

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

//
//The best mock resource you ever done seen
type MockResource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	NamespacedStatuses *core.NamespacedStatuses `protobuf:"bytes,16,opt,name=namespaced_statuses,json=namespacedStatuses,proto3" json:"namespaced_statuses,omitempty"`
	Metadata           *core.Metadata           `protobuf:"bytes,7,opt,name=metadata,proto3" json:"metadata,omitempty"`
	// Types that are assignable to WeStuckItInAOneof:
	//	*MockResource_SomeDumbField
	//	*MockResource_Data
	WeStuckItInAOneof isMockResource_WeStuckItInAOneof `protobuf_oneof:"we_stuck_it_in_a_oneof"`
	// Types that are assignable to TestOneofFields:
	//	*MockResource_OneofOne
	//	*MockResource_OneofTwo
	TestOneofFields isMockResource_TestOneofFields `protobuf_oneof:"test_oneof_fields"`
}

func (x *MockResource) Reset() {
	*x = MockResource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *MockResource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*MockResource) ProtoMessage() {}

func (x *MockResource) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use MockResource.ProtoReflect.Descriptor instead.
func (*MockResource) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_rawDescGZIP(), []int{0}
}

func (x *MockResource) GetNamespacedStatuses() *core.NamespacedStatuses {
	if x != nil {
		return x.NamespacedStatuses
	}
	return nil
}

func (x *MockResource) GetMetadata() *core.Metadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (m *MockResource) GetWeStuckItInAOneof() isMockResource_WeStuckItInAOneof {
	if m != nil {
		return m.WeStuckItInAOneof
	}
	return nil
}

func (x *MockResource) GetSomeDumbField() string {
	if x, ok := x.GetWeStuckItInAOneof().(*MockResource_SomeDumbField); ok {
		return x.SomeDumbField
	}
	return ""
}

func (x *MockResource) GetData() string {
	if x, ok := x.GetWeStuckItInAOneof().(*MockResource_Data); ok {
		return x.Data
	}
	return ""
}

func (m *MockResource) GetTestOneofFields() isMockResource_TestOneofFields {
	if m != nil {
		return m.TestOneofFields
	}
	return nil
}

func (x *MockResource) GetOneofOne() string {
	if x, ok := x.GetTestOneofFields().(*MockResource_OneofOne); ok {
		return x.OneofOne
	}
	return ""
}

func (x *MockResource) GetOneofTwo() bool {
	if x, ok := x.GetTestOneofFields().(*MockResource_OneofTwo); ok {
		return x.OneofTwo
	}
	return false
}

type isMockResource_WeStuckItInAOneof interface {
	isMockResource_WeStuckItInAOneof()
}

type MockResource_SomeDumbField struct {
	SomeDumbField string `protobuf:"bytes,100,opt,name=some_dumb_field,json=someDumbField,proto3,oneof"`
}

type MockResource_Data struct {
	Data string `protobuf:"bytes,1,opt,name=data,json=data.json,proto3,oneof"`
}

func (*MockResource_SomeDumbField) isMockResource_WeStuckItInAOneof() {}

func (*MockResource_Data) isMockResource_WeStuckItInAOneof() {}

type isMockResource_TestOneofFields interface {
	isMockResource_TestOneofFields()
}

type MockResource_OneofOne struct {
	OneofOne string `protobuf:"bytes,3,opt,name=oneof_one,json=oneofOne,proto3,oneof"`
}

type MockResource_OneofTwo struct {
	OneofTwo bool `protobuf:"varint,2,opt,name=oneof_two,json=oneofTwo,proto3,oneof"`
}

func (*MockResource_OneofOne) isMockResource_TestOneofFields() {}

func (*MockResource_OneofTwo) isMockResource_TestOneofFields() {}

type FrequentlyChangingAnnotationsResource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metadata *core.Metadata `protobuf:"bytes,7,opt,name=metadata,proto3" json:"metadata,omitempty"`
	Blah     string         `protobuf:"bytes,1,opt,name=blah,proto3" json:"blah,omitempty"`
}

func (x *FrequentlyChangingAnnotationsResource) Reset() {
	*x = FrequentlyChangingAnnotationsResource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *FrequentlyChangingAnnotationsResource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*FrequentlyChangingAnnotationsResource) ProtoMessage() {}

func (x *FrequentlyChangingAnnotationsResource) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use FrequentlyChangingAnnotationsResource.ProtoReflect.Descriptor instead.
func (*FrequentlyChangingAnnotationsResource) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_rawDescGZIP(), []int{1}
}

func (x *FrequentlyChangingAnnotationsResource) GetMetadata() *core.Metadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *FrequentlyChangingAnnotationsResource) GetBlah() string {
	if x != nil {
		return x.Blah
	}
	return ""
}

var File_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto protoreflect.FileDescriptor

var file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_rawDesc = []byte{
	0x0a, 0x48, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c,
	0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x74, 0x65,
	0x73, 0x74, 0x2f, 0x6d, 0x6f, 0x63, 0x6b, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x32, 0x61,
	0x6c, 0x70, 0x68, 0x61, 0x31, 0x2f, 0x6d, 0x6f, 0x63, 0x6b, 0x5f, 0x72, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x74, 0x65, 0x73, 0x74,
	0x69, 0x6e, 0x67, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x1a, 0x31, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f,
	0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f,
	0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2f,
	0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d,
	0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f,
	0x76, 0x31, 0x2f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a,
	0x31, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f,
	0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x61, 0x70, 0x69,
	0x2f, 0x76, 0x31, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x1a, 0x12, 0x65, 0x78, 0x74, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x78, 0x74,
	0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xee, 0x02, 0x0a, 0x0c, 0x4d, 0x6f, 0x63, 0x6b, 0x52,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x57, 0x0a, 0x13, 0x6e, 0x61, 0x6d, 0x65, 0x73,
	0x70, 0x61, 0x63, 0x65, 0x64, 0x5f, 0x73, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73, 0x18, 0x10,
	0x20, 0x01, 0x28, 0x0b, 0x32, 0x20, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73, 0x6f, 0x6c, 0x6f,
	0x2e, 0x69, 0x6f, 0x2e, 0x4e, 0x61, 0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x64, 0x53, 0x74,
	0x61, 0x74, 0x75, 0x73, 0x65, 0x73, 0x42, 0x04, 0xb8, 0xf5, 0x04, 0x01, 0x52, 0x12, 0x6e, 0x61,
	0x6d, 0x65, 0x73, 0x70, 0x61, 0x63, 0x65, 0x64, 0x53, 0x74, 0x61, 0x74, 0x75, 0x73, 0x65, 0x73,
	0x12, 0x32, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x07, 0x20, 0x01,
	0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69,
	0x6f, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x08, 0x6d, 0x65, 0x74, 0x61,
	0x64, 0x61, 0x74, 0x61, 0x12, 0x2e, 0x0a, 0x0f, 0x73, 0x6f, 0x6d, 0x65, 0x5f, 0x64, 0x75, 0x6d,
	0x62, 0x5f, 0x66, 0x69, 0x65, 0x6c, 0x64, 0x18, 0x64, 0x20, 0x01, 0x28, 0x09, 0x42, 0x04, 0xb8,
	0xf5, 0x04, 0x01, 0x48, 0x00, 0x52, 0x0d, 0x73, 0x6f, 0x6d, 0x65, 0x44, 0x75, 0x6d, 0x62, 0x46,
	0x69, 0x65, 0x6c, 0x64, 0x12, 0x19, 0x0a, 0x04, 0x64, 0x61, 0x74, 0x61, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x48, 0x00, 0x52, 0x09, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x6a, 0x73, 0x6f, 0x6e, 0x12,
	0x1d, 0x0a, 0x09, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x5f, 0x6f, 0x6e, 0x65, 0x18, 0x03, 0x20, 0x01,
	0x28, 0x09, 0x48, 0x01, 0x52, 0x08, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x4f, 0x6e, 0x65, 0x12, 0x1d,
	0x0a, 0x09, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x5f, 0x74, 0x77, 0x6f, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x08, 0x48, 0x01, 0x52, 0x08, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x54, 0x77, 0x6f, 0x3a, 0x13, 0x82,
	0xf1, 0x04, 0x04, 0x0a, 0x02, 0x6d, 0x6b, 0x82, 0xf1, 0x04, 0x07, 0x12, 0x05, 0x6d, 0x6f, 0x63,
	0x6b, 0x73, 0x42, 0x18, 0x0a, 0x16, 0x77, 0x65, 0x5f, 0x73, 0x74, 0x75, 0x63, 0x6b, 0x5f, 0x69,
	0x74, 0x5f, 0x69, 0x6e, 0x5f, 0x61, 0x5f, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x42, 0x13, 0x0a, 0x11,
	0x74, 0x65, 0x73, 0x74, 0x5f, 0x6f, 0x6e, 0x65, 0x6f, 0x66, 0x5f, 0x66, 0x69, 0x65, 0x6c, 0x64,
	0x73, 0x4a, 0x04, 0x08, 0x06, 0x10, 0x07, 0x22, 0x8c, 0x01, 0x0a, 0x25, 0x46, 0x72, 0x65, 0x71,
	0x75, 0x65, 0x6e, 0x74, 0x6c, 0x79, 0x43, 0x68, 0x61, 0x6e, 0x67, 0x69, 0x6e, 0x67, 0x41, 0x6e,
	0x6e, 0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63,
	0x65, 0x12, 0x32, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x18, 0x07, 0x20,
	0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e,
	0x69, 0x6f, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52, 0x08, 0x6d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x62, 0x6c, 0x61, 0x68, 0x18, 0x01, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x04, 0x62, 0x6c, 0x61, 0x68, 0x3a, 0x1b, 0x82, 0xf1, 0x04, 0x06, 0x0a,
	0x04, 0x66, 0x63, 0x61, 0x72, 0x82, 0xf1, 0x04, 0x07, 0x12, 0x05, 0x66, 0x63, 0x61, 0x72, 0x73,
	0x82, 0xf1, 0x04, 0x02, 0x28, 0x01, 0x42, 0x3d, 0x5a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c,
	0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x6d, 0x6f, 0x63, 0x6b, 0x73,
	0x2f, 0x76, 0x32, 0x61, 0x6c, 0x70, 0x68, 0x61, 0x31, 0xb8, 0xf5, 0x04, 0x01, 0xc0, 0xf5, 0x04,
	0x01, 0xd0, 0xf5, 0x04, 0x01, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_rawDescOnce sync.Once
	file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_rawDescData = file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_rawDesc
)

func file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_rawDescGZIP() []byte {
	file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_rawDescOnce.Do(func() {
		file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_rawDescData)
	})
	return file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_rawDescData
}

var file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_goTypes = []interface{}{
	(*MockResource)(nil),                          // 0: testing.solo.io.MockResource
	(*FrequentlyChangingAnnotationsResource)(nil), // 1: testing.solo.io.FrequentlyChangingAnnotationsResource
	(*core.NamespacedStatuses)(nil),               // 2: core.solo.io.NamespacedStatuses
	(*core.Metadata)(nil),                         // 3: core.solo.io.Metadata
}
var file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_depIdxs = []int32{
	2, // 0: testing.solo.io.MockResource.namespaced_statuses:type_name -> core.solo.io.NamespacedStatuses
	3, // 1: testing.solo.io.MockResource.metadata:type_name -> core.solo.io.Metadata
	3, // 2: testing.solo.io.FrequentlyChangingAnnotationsResource.metadata:type_name -> core.solo.io.Metadata
	3, // [3:3] is the sub-list for method output_type
	3, // [3:3] is the sub-list for method input_type
	3, // [3:3] is the sub-list for extension type_name
	3, // [3:3] is the sub-list for extension extendee
	0, // [0:3] is the sub-list for field type_name
}

func init() { file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_init() }
func file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_init() {
	if File_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*MockResource); i {
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
		file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*FrequentlyChangingAnnotationsResource); i {
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
	file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_msgTypes[0].OneofWrappers = []interface{}{
		(*MockResource_SomeDumbField)(nil),
		(*MockResource_Data)(nil),
		(*MockResource_OneofOne)(nil),
		(*MockResource_OneofTwo)(nil),
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_goTypes,
		DependencyIndexes: file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_depIdxs,
		MessageInfos:      file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_msgTypes,
	}.Build()
	File_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto = out.File
	file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_rawDesc = nil
	file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_goTypes = nil
	file_github_com_solo_io_solo_kit_test_mocks_api_v2alpha1_mock_resources_proto_depIdxs = nil
}
