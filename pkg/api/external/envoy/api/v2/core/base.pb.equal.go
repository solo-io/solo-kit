// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/api/external/envoy/api/v2/core/base.proto

package core

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"strings"

	"github.com/golang/protobuf/proto"
	equality "github.com/solo-io/protoc-gen-ext/pkg/equality"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
	_ = binary.LittleEndian
	_ = bytes.Compare
	_ = strings.Compare
	_ = equality.Equalizer(nil)
	_ = proto.Message(nil)
)

// Equal function
func (m *Locality) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*Locality)
	if !ok {
		that2, ok := that.(Locality)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if strings.Compare(m.GetRegion(), target.GetRegion()) != 0 {
		return false
	}

	if strings.Compare(m.GetZone(), target.GetZone()) != 0 {
		return false
	}

	if strings.Compare(m.GetSubZone(), target.GetSubZone()) != 0 {
		return false
	}

	return true
}

// Equal function
func (m *BuildVersion) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*BuildVersion)
	if !ok {
		that2, ok := that.(BuildVersion)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if h, ok := interface{}(m.GetVersion()).(equality.Equalizer); ok {
		if !h.Equal(target.GetVersion()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetVersion(), target.GetVersion()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetMetadata()).(equality.Equalizer); ok {
		if !h.Equal(target.GetMetadata()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetMetadata(), target.GetMetadata()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *Extension) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*Extension)
	if !ok {
		that2, ok := that.(Extension)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if strings.Compare(m.GetName(), target.GetName()) != 0 {
		return false
	}

	if strings.Compare(m.GetCategory(), target.GetCategory()) != 0 {
		return false
	}

	if strings.Compare(m.GetTypeDescriptor(), target.GetTypeDescriptor()) != 0 {
		return false
	}

	if h, ok := interface{}(m.GetVersion()).(equality.Equalizer); ok {
		if !h.Equal(target.GetVersion()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetVersion(), target.GetVersion()) {
			return false
		}
	}

	if m.GetDisabled() != target.GetDisabled() {
		return false
	}

	return true
}

// Equal function
func (m *Node) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*Node)
	if !ok {
		that2, ok := that.(Node)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if strings.Compare(m.GetId(), target.GetId()) != 0 {
		return false
	}

	if strings.Compare(m.GetCluster(), target.GetCluster()) != 0 {
		return false
	}

	if h, ok := interface{}(m.GetMetadata()).(equality.Equalizer); ok {
		if !h.Equal(target.GetMetadata()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetMetadata(), target.GetMetadata()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetLocality()).(equality.Equalizer); ok {
		if !h.Equal(target.GetLocality()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetLocality(), target.GetLocality()) {
			return false
		}
	}

	if strings.Compare(m.GetBuildVersion(), target.GetBuildVersion()) != 0 {
		return false
	}

	if strings.Compare(m.GetUserAgentName(), target.GetUserAgentName()) != 0 {
		return false
	}

	if len(m.GetExtensions()) != len(target.GetExtensions()) {
		return false
	}
	for idx, v := range m.GetExtensions() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetExtensions()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetExtensions()[idx]) {
				return false
			}
		}

	}

	if len(m.GetClientFeatures()) != len(target.GetClientFeatures()) {
		return false
	}
	for idx, v := range m.GetClientFeatures() {

		if strings.Compare(v, target.GetClientFeatures()[idx]) != 0 {
			return false
		}

	}

	if len(m.GetListeningAddresses()) != len(target.GetListeningAddresses()) {
		return false
	}
	for idx, v := range m.GetListeningAddresses() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetListeningAddresses()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetListeningAddresses()[idx]) {
				return false
			}
		}

	}

	switch m.UserAgentVersionType.(type) {

	case *Node_UserAgentVersion:

		if strings.Compare(m.GetUserAgentVersion(), target.GetUserAgentVersion()) != 0 {
			return false
		}

	case *Node_UserAgentBuildVersion:

		if h, ok := interface{}(m.GetUserAgentBuildVersion()).(equality.Equalizer); ok {
			if !h.Equal(target.GetUserAgentBuildVersion()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetUserAgentBuildVersion(), target.GetUserAgentBuildVersion()) {
				return false
			}
		}

	}

	return true
}

// Equal function
func (m *Metadata) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*Metadata)
	if !ok {
		that2, ok := that.(Metadata)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if len(m.GetFilterMetadata()) != len(target.GetFilterMetadata()) {
		return false
	}
	for k, v := range m.GetFilterMetadata() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetFilterMetadata()[k]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetFilterMetadata()[k]) {
				return false
			}
		}

	}

	return true
}

// Equal function
func (m *RuntimeUInt32) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*RuntimeUInt32)
	if !ok {
		that2, ok := that.(RuntimeUInt32)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if m.GetDefaultValue() != target.GetDefaultValue() {
		return false
	}

	if strings.Compare(m.GetRuntimeKey(), target.GetRuntimeKey()) != 0 {
		return false
	}

	return true
}

// Equal function
func (m *RuntimeFeatureFlag) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*RuntimeFeatureFlag)
	if !ok {
		that2, ok := that.(RuntimeFeatureFlag)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if h, ok := interface{}(m.GetDefaultValue()).(equality.Equalizer); ok {
		if !h.Equal(target.GetDefaultValue()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetDefaultValue(), target.GetDefaultValue()) {
			return false
		}
	}

	if strings.Compare(m.GetRuntimeKey(), target.GetRuntimeKey()) != 0 {
		return false
	}

	return true
}

// Equal function
func (m *HeaderValue) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*HeaderValue)
	if !ok {
		that2, ok := that.(HeaderValue)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if strings.Compare(m.GetKey(), target.GetKey()) != 0 {
		return false
	}

	if strings.Compare(m.GetValue(), target.GetValue()) != 0 {
		return false
	}

	return true
}

// Equal function
func (m *HeaderValueOption) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*HeaderValueOption)
	if !ok {
		that2, ok := that.(HeaderValueOption)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if h, ok := interface{}(m.GetAppend()).(equality.Equalizer); ok {
		if !h.Equal(target.GetAppend()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetAppend(), target.GetAppend()) {
			return false
		}
	}

	switch m.HeaderOption.(type) {

	case *HeaderValueOption_Header:

		if h, ok := interface{}(m.GetHeader()).(equality.Equalizer); ok {
			if !h.Equal(target.GetHeader()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetHeader(), target.GetHeader()) {
				return false
			}
		}

	case *HeaderValueOption_HeaderSecretRef:

		if h, ok := interface{}(m.GetHeaderSecretRef()).(equality.Equalizer); ok {
			if !h.Equal(target.GetHeaderSecretRef()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetHeaderSecretRef(), target.GetHeaderSecretRef()) {
				return false
			}
		}

	}

	return true
}

// Equal function
func (m *HeaderMap) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*HeaderMap)
	if !ok {
		that2, ok := that.(HeaderMap)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if len(m.GetHeaders()) != len(target.GetHeaders()) {
		return false
	}
	for idx, v := range m.GetHeaders() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetHeaders()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetHeaders()[idx]) {
				return false
			}
		}

	}

	return true
}

// Equal function
func (m *DataSource) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*DataSource)
	if !ok {
		that2, ok := that.(DataSource)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	switch m.Specifier.(type) {

	case *DataSource_Filename:

		if strings.Compare(m.GetFilename(), target.GetFilename()) != 0 {
			return false
		}

	case *DataSource_InlineBytes:

		if bytes.Compare(m.GetInlineBytes(), target.GetInlineBytes()) != 0 {
			return false
		}

	case *DataSource_InlineString:

		if strings.Compare(m.GetInlineString(), target.GetInlineString()) != 0 {
			return false
		}

	}

	return true
}

// Equal function
func (m *RemoteDataSource) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*RemoteDataSource)
	if !ok {
		that2, ok := that.(RemoteDataSource)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if h, ok := interface{}(m.GetHttpUri()).(equality.Equalizer); ok {
		if !h.Equal(target.GetHttpUri()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetHttpUri(), target.GetHttpUri()) {
			return false
		}
	}

	if strings.Compare(m.GetSha256(), target.GetSha256()) != 0 {
		return false
	}

	return true
}

// Equal function
func (m *AsyncDataSource) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*AsyncDataSource)
	if !ok {
		that2, ok := that.(AsyncDataSource)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	switch m.Specifier.(type) {

	case *AsyncDataSource_Local:

		if h, ok := interface{}(m.GetLocal()).(equality.Equalizer); ok {
			if !h.Equal(target.GetLocal()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetLocal(), target.GetLocal()) {
				return false
			}
		}

	case *AsyncDataSource_Remote:

		if h, ok := interface{}(m.GetRemote()).(equality.Equalizer); ok {
			if !h.Equal(target.GetRemote()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetRemote(), target.GetRemote()) {
				return false
			}
		}

	}

	return true
}

// Equal function
func (m *TransportSocket) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*TransportSocket)
	if !ok {
		that2, ok := that.(TransportSocket)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if strings.Compare(m.GetName(), target.GetName()) != 0 {
		return false
	}

	switch m.ConfigType.(type) {

	case *TransportSocket_Config:

		if h, ok := interface{}(m.GetConfig()).(equality.Equalizer); ok {
			if !h.Equal(target.GetConfig()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetConfig(), target.GetConfig()) {
				return false
			}
		}

	case *TransportSocket_TypedConfig:

		if h, ok := interface{}(m.GetTypedConfig()).(equality.Equalizer); ok {
			if !h.Equal(target.GetTypedConfig()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetTypedConfig(), target.GetTypedConfig()) {
				return false
			}
		}

	}

	return true
}

// Equal function
func (m *RuntimeFractionalPercent) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*RuntimeFractionalPercent)
	if !ok {
		that2, ok := that.(RuntimeFractionalPercent)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if h, ok := interface{}(m.GetDefaultValue()).(equality.Equalizer); ok {
		if !h.Equal(target.GetDefaultValue()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetDefaultValue(), target.GetDefaultValue()) {
			return false
		}
	}

	if strings.Compare(m.GetRuntimeKey(), target.GetRuntimeKey()) != 0 {
		return false
	}

	return true
}

// Equal function
func (m *ControlPlane) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*ControlPlane)
	if !ok {
		that2, ok := that.(ControlPlane)
		if ok {
			target = &that2
		} else {
			return false
		}
	}
	if target == nil {
		return m == nil
	} else if m == nil {
		return false
	}

	if strings.Compare(m.GetIdentifier(), target.GetIdentifier()) != 0 {
		return false
	}

	return true
}
