// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/test/mocks/api/v1/simple_mock_resources.proto

package v1

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/solo-io/protoc-gen-ext/pkg/clone"
	"google.golang.org/protobuf/proto"

	github_com_golang_protobuf_ptypes_any "github.com/golang/protobuf/ptypes/any"

	github_com_golang_protobuf_ptypes_duration "github.com/golang/protobuf/ptypes/duration"

	github_com_golang_protobuf_ptypes_empty "github.com/golang/protobuf/ptypes/empty"

	github_com_golang_protobuf_ptypes_struct "github.com/golang/protobuf/ptypes/struct"

	github_com_golang_protobuf_ptypes_timestamp "github.com/golang/protobuf/ptypes/timestamp"

	github_com_golang_protobuf_ptypes_wrappers "github.com/golang/protobuf/ptypes/wrappers"

	github_com_solo_io_solo_kit_pkg_api_v1_resources_core "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
	_ = binary.LittleEndian
	_ = bytes.Compare
	_ = strings.Compare
	_ = clone.Cloner(nil)
	_ = proto.Message(nil)
)

// Clone function
func (m *SimpleMockResource) Clone() proto.Message {
	var target *SimpleMockResource
	if m == nil {
		return target
	}
	target = &SimpleMockResource{}

	if h, ok := interface{}(m.GetMetadata()).(clone.Cloner); ok {
		target.Metadata = h.Clone().(*github_com_solo_io_solo_kit_pkg_api_v1_resources_core.Metadata)
	} else {
		target.Metadata = proto.Clone(m.GetMetadata()).(*github_com_solo_io_solo_kit_pkg_api_v1_resources_core.Metadata)
	}

	target.Data = m.GetData()

	if m.GetMappedData() != nil {
		target.MappedData = make(map[string]string, len(m.GetMappedData()))
		for k, v := range m.GetMappedData() {

			target.MappedData[k] = v

		}
	}

	if m.GetList() != nil {
		target.List = make([]bool, len(m.GetList()))
		for idx, v := range m.GetList() {

			target.List[idx] = v

		}
	}

	target.Int64Data = m.GetInt64Data()

	target.DataWithLongComment = m.GetDataWithLongComment()

	if h, ok := interface{}(m.GetNestedMessage()).(clone.Cloner); ok {
		target.NestedMessage = h.Clone().(*SimpleMockResource_NestedMessage)
	} else {
		target.NestedMessage = proto.Clone(m.GetNestedMessage()).(*SimpleMockResource_NestedMessage)
	}

	if m.GetNestedMessageList() != nil {
		target.NestedMessageList = make([]*SimpleMockResource_NestedMessage, len(m.GetNestedMessageList()))
		for idx, v := range m.GetNestedMessageList() {

			if h, ok := interface{}(v).(clone.Cloner); ok {
				target.NestedMessageList[idx] = h.Clone().(*SimpleMockResource_NestedMessage)
			} else {
				target.NestedMessageList[idx] = proto.Clone(v).(*SimpleMockResource_NestedMessage)
			}

		}
	}

	if h, ok := interface{}(m.GetAny()).(clone.Cloner); ok {
		target.Any = h.Clone().(*github_com_golang_protobuf_ptypes_any.Any)
	} else {
		target.Any = proto.Clone(m.GetAny()).(*github_com_golang_protobuf_ptypes_any.Any)
	}

	if h, ok := interface{}(m.GetStruct()).(clone.Cloner); ok {
		target.Struct = h.Clone().(*github_com_golang_protobuf_ptypes_struct.Struct)
	} else {
		target.Struct = proto.Clone(m.GetStruct()).(*github_com_golang_protobuf_ptypes_struct.Struct)
	}

	if m.GetMappedStruct() != nil {
		target.MappedStruct = make(map[string]*github_com_golang_protobuf_ptypes_struct.Struct, len(m.GetMappedStruct()))
		for k, v := range m.GetMappedStruct() {

			if h, ok := interface{}(v).(clone.Cloner); ok {
				target.MappedStruct[k] = h.Clone().(*github_com_golang_protobuf_ptypes_struct.Struct)
			} else {
				target.MappedStruct[k] = proto.Clone(v).(*github_com_golang_protobuf_ptypes_struct.Struct)
			}

		}
	}

	if h, ok := interface{}(m.GetBoolValue()).(clone.Cloner); ok {
		target.BoolValue = h.Clone().(*github_com_golang_protobuf_ptypes_wrappers.BoolValue)
	} else {
		target.BoolValue = proto.Clone(m.GetBoolValue()).(*github_com_golang_protobuf_ptypes_wrappers.BoolValue)
	}

	if h, ok := interface{}(m.GetInt32Value()).(clone.Cloner); ok {
		target.Int32Value = h.Clone().(*github_com_golang_protobuf_ptypes_wrappers.Int32Value)
	} else {
		target.Int32Value = proto.Clone(m.GetInt32Value()).(*github_com_golang_protobuf_ptypes_wrappers.Int32Value)
	}

	if h, ok := interface{}(m.GetUint32Value()).(clone.Cloner); ok {
		target.Uint32Value = h.Clone().(*github_com_golang_protobuf_ptypes_wrappers.UInt32Value)
	} else {
		target.Uint32Value = proto.Clone(m.GetUint32Value()).(*github_com_golang_protobuf_ptypes_wrappers.UInt32Value)
	}

	if h, ok := interface{}(m.GetFloatValue()).(clone.Cloner); ok {
		target.FloatValue = h.Clone().(*github_com_golang_protobuf_ptypes_wrappers.FloatValue)
	} else {
		target.FloatValue = proto.Clone(m.GetFloatValue()).(*github_com_golang_protobuf_ptypes_wrappers.FloatValue)
	}

	if h, ok := interface{}(m.GetDuration()).(clone.Cloner); ok {
		target.Duration = h.Clone().(*github_com_golang_protobuf_ptypes_duration.Duration)
	} else {
		target.Duration = proto.Clone(m.GetDuration()).(*github_com_golang_protobuf_ptypes_duration.Duration)
	}

	if h, ok := interface{}(m.GetEmpty()).(clone.Cloner); ok {
		target.Empty = h.Clone().(*github_com_golang_protobuf_ptypes_empty.Empty)
	} else {
		target.Empty = proto.Clone(m.GetEmpty()).(*github_com_golang_protobuf_ptypes_empty.Empty)
	}

	if h, ok := interface{}(m.GetStringValue()).(clone.Cloner); ok {
		target.StringValue = h.Clone().(*github_com_golang_protobuf_ptypes_wrappers.StringValue)
	} else {
		target.StringValue = proto.Clone(m.GetStringValue()).(*github_com_golang_protobuf_ptypes_wrappers.StringValue)
	}

	if h, ok := interface{}(m.GetDoubleValue()).(clone.Cloner); ok {
		target.DoubleValue = h.Clone().(*github_com_golang_protobuf_ptypes_wrappers.DoubleValue)
	} else {
		target.DoubleValue = proto.Clone(m.GetDoubleValue()).(*github_com_golang_protobuf_ptypes_wrappers.DoubleValue)
	}

	if h, ok := interface{}(m.GetTimestamp()).(clone.Cloner); ok {
		target.Timestamp = h.Clone().(*github_com_golang_protobuf_ptypes_timestamp.Timestamp)
	} else {
		target.Timestamp = proto.Clone(m.GetTimestamp()).(*github_com_golang_protobuf_ptypes_timestamp.Timestamp)
	}

	target.EnumOptions = m.GetEnumOptions()

	return target
}

// Clone function
func (m *SimpleMockResource_NestedMessage) Clone() proto.Message {
	var target *SimpleMockResource_NestedMessage
	if m == nil {
		return target
	}
	target = &SimpleMockResource_NestedMessage{}

	target.OptionBool = m.GetOptionBool()

	target.OptionString = m.GetOptionString()

	return target
}
