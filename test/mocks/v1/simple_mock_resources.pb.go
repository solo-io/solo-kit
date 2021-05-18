// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.6.1
// source: github.com/solo-io/solo-kit/test/mocks/api/v1/simple_mock_resources.proto

package v1

import (
	reflect "reflect"
	sync "sync"

	proto "github.com/golang/protobuf/proto"
	any "github.com/golang/protobuf/ptypes/any"
	_struct "github.com/golang/protobuf/ptypes/struct"
	_ "github.com/solo-io/protoc-gen-ext/extproto"
	core "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	_ "google.golang.org/genproto/googleapis/api/annotations"
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

//
//A SimpleMockResource defines a variety of baseline types to ensure
//that we can generate open api schemas properly. It intentionally avoids
//messages that include oneof and recursive schemas (like core.solo.io.Status)
type SimpleMockResource struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Metadata          *core.Metadata                      `protobuf:"bytes,100,opt,name=metadata,proto3" json:"metadata,omitempty"`
	Data              string                              `protobuf:"bytes,1,opt,name=data,proto3" json:"data,omitempty"`
	MappedData        map[string]string                   `protobuf:"bytes,2,rep,name=mapped_data,json=mappedData,proto3" json:"mapped_data,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	List              []bool                              `protobuf:"varint,3,rep,packed,name=list,proto3" json:"list,omitempty"`
	NestedMessage     *SimpleMockResource_NestedMessage   `protobuf:"bytes,4,opt,name=nested_message,json=nestedMessage,proto3" json:"nested_message,omitempty"`
	NestedMessageList []*SimpleMockResource_NestedMessage `protobuf:"bytes,5,rep,name=nested_message_list,json=nestedMessageList,proto3" json:"nested_message_list,omitempty"`
	Any               *any.Any                            `protobuf:"bytes,11,opt,name=any,proto3" json:"any,omitempty"`
	Struct            *_struct.Struct                     `protobuf:"bytes,12,opt,name=struct,proto3" json:"struct,omitempty"`
	MappedStruct      map[string]*_struct.Struct          `protobuf:"bytes,13,rep,name=mapped_struct,json=mappedStruct,proto3" json:"mapped_struct,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
}

func (x *SimpleMockResource) Reset() {
	*x = SimpleMockResource{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SimpleMockResource) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SimpleMockResource) ProtoMessage() {}

func (x *SimpleMockResource) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SimpleMockResource.ProtoReflect.Descriptor instead.
func (*SimpleMockResource) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_rawDescGZIP(), []int{0}
}

func (x *SimpleMockResource) GetMetadata() *core.Metadata {
	if x != nil {
		return x.Metadata
	}
	return nil
}

func (x *SimpleMockResource) GetData() string {
	if x != nil {
		return x.Data
	}
	return ""
}

func (x *SimpleMockResource) GetMappedData() map[string]string {
	if x != nil {
		return x.MappedData
	}
	return nil
}

func (x *SimpleMockResource) GetList() []bool {
	if x != nil {
		return x.List
	}
	return nil
}

func (x *SimpleMockResource) GetNestedMessage() *SimpleMockResource_NestedMessage {
	if x != nil {
		return x.NestedMessage
	}
	return nil
}

func (x *SimpleMockResource) GetNestedMessageList() []*SimpleMockResource_NestedMessage {
	if x != nil {
		return x.NestedMessageList
	}
	return nil
}

func (x *SimpleMockResource) GetAny() *any.Any {
	if x != nil {
		return x.Any
	}
	return nil
}

func (x *SimpleMockResource) GetStruct() *_struct.Struct {
	if x != nil {
		return x.Struct
	}
	return nil
}

func (x *SimpleMockResource) GetMappedStruct() map[string]*_struct.Struct {
	if x != nil {
		return x.MappedStruct
	}
	return nil
}

type SimpleMockResource_NestedMessage struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	OptionBool   bool   `protobuf:"varint,1,opt,name=option_bool,json=optionBool,proto3" json:"option_bool,omitempty"`
	OptionString string `protobuf:"bytes,2,opt,name=option_string,json=optionString,proto3" json:"option_string,omitempty"`
}

func (x *SimpleMockResource_NestedMessage) Reset() {
	*x = SimpleMockResource_NestedMessage{}
	if protoimpl.UnsafeEnabled {
		mi := &file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *SimpleMockResource_NestedMessage) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*SimpleMockResource_NestedMessage) ProtoMessage() {}

func (x *SimpleMockResource_NestedMessage) ProtoReflect() protoreflect.Message {
	mi := &file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use SimpleMockResource_NestedMessage.ProtoReflect.Descriptor instead.
func (*SimpleMockResource_NestedMessage) Descriptor() ([]byte, []int) {
	return file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_rawDescGZIP(), []int{0, 2}
}

func (x *SimpleMockResource_NestedMessage) GetOptionBool() bool {
	if x != nil {
		return x.OptionBool
	}
	return false
}

func (x *SimpleMockResource_NestedMessage) GetOptionString() string {
	if x != nil {
		return x.OptionString
	}
	return ""
}

var File_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto protoreflect.FileDescriptor

var file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_rawDesc = []byte{
	0x0a, 0x49, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c,
	0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x74, 0x65,
	0x73, 0x74, 0x2f, 0x6d, 0x6f, 0x63, 0x6b, 0x73, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f,
	0x73, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x5f, 0x6d, 0x6f, 0x63, 0x6b, 0x5f, 0x72, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0f, 0x74, 0x65, 0x73,
	0x74, 0x69, 0x6e, 0x67, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x1a, 0x12, 0x65, 0x78,
	0x74, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x65, 0x78, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x61, 0x6e, 0x6e,
	0x6f, 0x74, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x19,
	0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f,
	0x61, 0x6e, 0x79, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x1c, 0x67, 0x6f, 0x6f, 0x67, 0x6c,
	0x65, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2f, 0x73, 0x74, 0x72, 0x75, 0x63,
	0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x2f, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c, 0x6f,
	0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x73, 0x74, 0x61, 0x74,
	0x75, 0x73, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x31, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62,
	0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c,
	0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f, 0x6d, 0x65, 0x74,
	0x61, 0x64, 0x61, 0x74, 0x61, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x1a, 0x31, 0x67, 0x69, 0x74,
	0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f,
	0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x76, 0x31, 0x2f,
	0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x6b, 0x69, 0x74, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x22, 0xc4,
	0x06, 0x0a, 0x12, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x4d, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73,
	0x6f, 0x75, 0x72, 0x63, 0x65, 0x12, 0x32, 0x0a, 0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74,
	0x61, 0x18, 0x64, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x16, 0x2e, 0x63, 0x6f, 0x72, 0x65, 0x2e, 0x73,
	0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x4d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x52,
	0x08, 0x6d, 0x65, 0x74, 0x61, 0x64, 0x61, 0x74, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x64, 0x61, 0x74,
	0x61, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x64, 0x61, 0x74, 0x61, 0x12, 0x54, 0x0a,
	0x0b, 0x6d, 0x61, 0x70, 0x70, 0x65, 0x64, 0x5f, 0x64, 0x61, 0x74, 0x61, 0x18, 0x02, 0x20, 0x03,
	0x28, 0x0b, 0x32, 0x33, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x73, 0x6f, 0x6c,
	0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x4d, 0x6f, 0x63, 0x6b, 0x52,
	0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x4d, 0x61, 0x70, 0x70, 0x65, 0x64, 0x44, 0x61,
	0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0a, 0x6d, 0x61, 0x70, 0x70, 0x65, 0x64, 0x44,
	0x61, 0x74, 0x61, 0x12, 0x12, 0x0a, 0x04, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x03, 0x20, 0x03, 0x28,
	0x08, 0x52, 0x04, 0x6c, 0x69, 0x73, 0x74, 0x12, 0x58, 0x0a, 0x0e, 0x6e, 0x65, 0x73, 0x74, 0x65,
	0x64, 0x5f, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x04, 0x20, 0x01, 0x28, 0x0b, 0x32,
	0x31, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69,
	0x6f, 0x2e, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x4d, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x6f,
	0x75, 0x72, 0x63, 0x65, 0x2e, 0x4e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x52, 0x0d, 0x6e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x12, 0x61, 0x0a, 0x13, 0x6e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x5f, 0x6d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x5f, 0x6c, 0x69, 0x73, 0x74, 0x18, 0x05, 0x20, 0x03, 0x28, 0x0b, 0x32, 0x31,
	0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x73, 0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f,
	0x2e, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x4d, 0x6f, 0x63, 0x6b, 0x52, 0x65, 0x73, 0x6f, 0x75,
	0x72, 0x63, 0x65, 0x2e, 0x4e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x52, 0x11, 0x6e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65,
	0x4c, 0x69, 0x73, 0x74, 0x12, 0x26, 0x0a, 0x03, 0x61, 0x6e, 0x79, 0x18, 0x0b, 0x20, 0x01, 0x28,
	0x0b, 0x32, 0x14, 0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f,
	0x62, 0x75, 0x66, 0x2e, 0x41, 0x6e, 0x79, 0x52, 0x03, 0x61, 0x6e, 0x79, 0x12, 0x2f, 0x0a, 0x06,
	0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x18, 0x0c, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17, 0x2e, 0x67,
	0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66, 0x2e, 0x53,
	0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x06, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x12, 0x5a, 0x0a,
	0x0d, 0x6d, 0x61, 0x70, 0x70, 0x65, 0x64, 0x5f, 0x73, 0x74, 0x72, 0x75, 0x63, 0x74, 0x18, 0x0d,
	0x20, 0x03, 0x28, 0x0b, 0x32, 0x35, 0x2e, 0x74, 0x65, 0x73, 0x74, 0x69, 0x6e, 0x67, 0x2e, 0x73,
	0x6f, 0x6c, 0x6f, 0x2e, 0x69, 0x6f, 0x2e, 0x53, 0x69, 0x6d, 0x70, 0x6c, 0x65, 0x4d, 0x6f, 0x63,
	0x6b, 0x52, 0x65, 0x73, 0x6f, 0x75, 0x72, 0x63, 0x65, 0x2e, 0x4d, 0x61, 0x70, 0x70, 0x65, 0x64,
	0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x52, 0x0c, 0x6d, 0x61, 0x70,
	0x70, 0x65, 0x64, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x1a, 0x3d, 0x0a, 0x0f, 0x4d, 0x61, 0x70,
	0x70, 0x65, 0x64, 0x44, 0x61, 0x74, 0x61, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a, 0x03,
	0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12, 0x14,
	0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x05, 0x76,
	0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02, 0x38, 0x01, 0x1a, 0x58, 0x0a, 0x11, 0x4d, 0x61, 0x70, 0x70,
	0x65, 0x64, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x45, 0x6e, 0x74, 0x72, 0x79, 0x12, 0x10, 0x0a,
	0x03, 0x6b, 0x65, 0x79, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x03, 0x6b, 0x65, 0x79, 0x12,
	0x2d, 0x0a, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x0b, 0x32, 0x17,
	0x2e, 0x67, 0x6f, 0x6f, 0x67, 0x6c, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x62, 0x75, 0x66,
	0x2e, 0x53, 0x74, 0x72, 0x75, 0x63, 0x74, 0x52, 0x05, 0x76, 0x61, 0x6c, 0x75, 0x65, 0x3a, 0x02,
	0x38, 0x01, 0x1a, 0x55, 0x0a, 0x0d, 0x4e, 0x65, 0x73, 0x74, 0x65, 0x64, 0x4d, 0x65, 0x73, 0x73,
	0x61, 0x67, 0x65, 0x12, 0x1f, 0x0a, 0x0b, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x62, 0x6f,
	0x6f, 0x6c, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x0a, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e,
	0x42, 0x6f, 0x6f, 0x6c, 0x12, 0x23, 0x0a, 0x0d, 0x6f, 0x70, 0x74, 0x69, 0x6f, 0x6e, 0x5f, 0x73,
	0x74, 0x72, 0x69, 0x6e, 0x67, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0c, 0x6f, 0x70, 0x74,
	0x69, 0x6f, 0x6e, 0x53, 0x74, 0x72, 0x69, 0x6e, 0x67, 0x3a, 0x1a, 0x82, 0xf1, 0x04, 0x05, 0x0a,
	0x03, 0x73, 0x6d, 0x6b, 0x82, 0xf1, 0x04, 0x0d, 0x12, 0x0b, 0x73, 0x69, 0x6d, 0x70, 0x6c, 0x65,
	0x6d, 0x6f, 0x63, 0x6b, 0x73, 0x42, 0x33, 0x5a, 0x29, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e,
	0x63, 0x6f, 0x6d, 0x2f, 0x73, 0x6f, 0x6c, 0x6f, 0x2d, 0x69, 0x6f, 0x2f, 0x73, 0x6f, 0x6c, 0x6f,
	0x2d, 0x6b, 0x69, 0x74, 0x2f, 0x74, 0x65, 0x73, 0x74, 0x2f, 0x6d, 0x6f, 0x63, 0x6b, 0x73, 0x2f,
	0x76, 0x31, 0xb8, 0xf5, 0x04, 0x01, 0xc0, 0xf5, 0x04, 0x01, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_rawDescOnce sync.Once
	file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_rawDescData = file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_rawDesc
)

func file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_rawDescGZIP() []byte {
	file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_rawDescOnce.Do(func() {
		file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_rawDescData = protoimpl.X.CompressGZIP(file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_rawDescData)
	})
	return file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_rawDescData
}

var file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_msgTypes = make([]protoimpl.MessageInfo, 4)
var file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_goTypes = []interface{}{
	(*SimpleMockResource)(nil),               // 0: testing.solo.io.SimpleMockResource
	nil,                                      // 1: testing.solo.io.SimpleMockResource.MappedDataEntry
	nil,                                      // 2: testing.solo.io.SimpleMockResource.MappedStructEntry
	(*SimpleMockResource_NestedMessage)(nil), // 3: testing.solo.io.SimpleMockResource.NestedMessage
	(*core.Metadata)(nil),                    // 4: core.solo.io.Metadata
	(*any.Any)(nil),                          // 5: google.protobuf.Any
	(*_struct.Struct)(nil),                   // 6: google.protobuf.Struct
}
var file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_depIdxs = []int32{
	4, // 0: testing.solo.io.SimpleMockResource.metadata:type_name -> core.solo.io.Metadata
	1, // 1: testing.solo.io.SimpleMockResource.mapped_data:type_name -> testing.solo.io.SimpleMockResource.MappedDataEntry
	3, // 2: testing.solo.io.SimpleMockResource.nested_message:type_name -> testing.solo.io.SimpleMockResource.NestedMessage
	3, // 3: testing.solo.io.SimpleMockResource.nested_message_list:type_name -> testing.solo.io.SimpleMockResource.NestedMessage
	5, // 4: testing.solo.io.SimpleMockResource.any:type_name -> google.protobuf.Any
	6, // 5: testing.solo.io.SimpleMockResource.struct:type_name -> google.protobuf.Struct
	2, // 6: testing.solo.io.SimpleMockResource.mapped_struct:type_name -> testing.solo.io.SimpleMockResource.MappedStructEntry
	6, // 7: testing.solo.io.SimpleMockResource.MappedStructEntry.value:type_name -> google.protobuf.Struct
	8, // [8:8] is the sub-list for method output_type
	8, // [8:8] is the sub-list for method input_type
	8, // [8:8] is the sub-list for extension type_name
	8, // [8:8] is the sub-list for extension extendee
	0, // [0:8] is the sub-list for field type_name
}

func init() { file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_init() }
func file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_init() {
	if File_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SimpleMockResource); i {
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
		file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*SimpleMockResource_NestedMessage); i {
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
			RawDescriptor: file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   4,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_goTypes,
		DependencyIndexes: file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_depIdxs,
		MessageInfos:      file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_msgTypes,
	}.Build()
	File_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto = out.File
	file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_rawDesc = nil
	file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_goTypes = nil
	file_github_com_solo_io_solo_kit_test_mocks_api_v1_simple_mock_resources_proto_depIdxs = nil
}
