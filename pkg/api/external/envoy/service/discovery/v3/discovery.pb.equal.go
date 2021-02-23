// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/api/external/envoy/service/discovery/v3/discovery.proto

package v3

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
func (m *DiscoveryRequest) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*DiscoveryRequest)
	if !ok {
		that2, ok := that.(DiscoveryRequest)
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

	if strings.Compare(m.GetVersionInfo(), target.GetVersionInfo()) != 0 {
		return false
	}

	if h, ok := interface{}(m.GetNode()).(equality.Equalizer); ok {
		if !h.Equal(target.GetNode()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetNode(), target.GetNode()) {
			return false
		}
	}

	if len(m.GetResourceNames()) != len(target.GetResourceNames()) {
		return false
	}
	for idx, v := range m.GetResourceNames() {

		if strings.Compare(v, target.GetResourceNames()[idx]) != 0 {
			return false
		}

	}

	if strings.Compare(m.GetTypeUrl(), target.GetTypeUrl()) != 0 {
		return false
	}

	if strings.Compare(m.GetResponseNonce(), target.GetResponseNonce()) != 0 {
		return false
	}

	if h, ok := interface{}(m.GetErrorDetail()).(equality.Equalizer); ok {
		if !h.Equal(target.GetErrorDetail()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetErrorDetail(), target.GetErrorDetail()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *DiscoveryResponse) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*DiscoveryResponse)
	if !ok {
		that2, ok := that.(DiscoveryResponse)
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

	if strings.Compare(m.GetVersionInfo(), target.GetVersionInfo()) != 0 {
		return false
	}

	if len(m.GetResources()) != len(target.GetResources()) {
		return false
	}
	for idx, v := range m.GetResources() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetResources()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetResources()[idx]) {
				return false
			}
		}

	}

	if m.GetCanary() != target.GetCanary() {
		return false
	}

	if strings.Compare(m.GetTypeUrl(), target.GetTypeUrl()) != 0 {
		return false
	}

	if strings.Compare(m.GetNonce(), target.GetNonce()) != 0 {
		return false
	}

	if h, ok := interface{}(m.GetControlPlane()).(equality.Equalizer); ok {
		if !h.Equal(target.GetControlPlane()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetControlPlane(), target.GetControlPlane()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *DeltaDiscoveryRequest) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*DeltaDiscoveryRequest)
	if !ok {
		that2, ok := that.(DeltaDiscoveryRequest)
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

	if h, ok := interface{}(m.GetNode()).(equality.Equalizer); ok {
		if !h.Equal(target.GetNode()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetNode(), target.GetNode()) {
			return false
		}
	}

	if strings.Compare(m.GetTypeUrl(), target.GetTypeUrl()) != 0 {
		return false
	}

	if len(m.GetResourceNamesSubscribe()) != len(target.GetResourceNamesSubscribe()) {
		return false
	}
	for idx, v := range m.GetResourceNamesSubscribe() {

		if strings.Compare(v, target.GetResourceNamesSubscribe()[idx]) != 0 {
			return false
		}

	}

	if len(m.GetResourceNamesUnsubscribe()) != len(target.GetResourceNamesUnsubscribe()) {
		return false
	}
	for idx, v := range m.GetResourceNamesUnsubscribe() {

		if strings.Compare(v, target.GetResourceNamesUnsubscribe()[idx]) != 0 {
			return false
		}

	}

	if len(m.GetInitialResourceVersions()) != len(target.GetInitialResourceVersions()) {
		return false
	}
	for k, v := range m.GetInitialResourceVersions() {

		if strings.Compare(v, target.GetInitialResourceVersions()[k]) != 0 {
			return false
		}

	}

	if strings.Compare(m.GetResponseNonce(), target.GetResponseNonce()) != 0 {
		return false
	}

	if h, ok := interface{}(m.GetErrorDetail()).(equality.Equalizer); ok {
		if !h.Equal(target.GetErrorDetail()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetErrorDetail(), target.GetErrorDetail()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *DeltaDiscoveryResponse) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*DeltaDiscoveryResponse)
	if !ok {
		that2, ok := that.(DeltaDiscoveryResponse)
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

	if strings.Compare(m.GetSystemVersionInfo(), target.GetSystemVersionInfo()) != 0 {
		return false
	}

	if len(m.GetResources()) != len(target.GetResources()) {
		return false
	}
	for idx, v := range m.GetResources() {

		if h, ok := interface{}(v).(equality.Equalizer); ok {
			if !h.Equal(target.GetResources()[idx]) {
				return false
			}
		} else {
			if !proto.Equal(v, target.GetResources()[idx]) {
				return false
			}
		}

	}

	if strings.Compare(m.GetTypeUrl(), target.GetTypeUrl()) != 0 {
		return false
	}

	if len(m.GetRemovedResources()) != len(target.GetRemovedResources()) {
		return false
	}
	for idx, v := range m.GetRemovedResources() {

		if strings.Compare(v, target.GetRemovedResources()[idx]) != 0 {
			return false
		}

	}

	if strings.Compare(m.GetNonce(), target.GetNonce()) != 0 {
		return false
	}

	if h, ok := interface{}(m.GetControlPlane()).(equality.Equalizer); ok {
		if !h.Equal(target.GetControlPlane()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetControlPlane(), target.GetControlPlane()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *Resource) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*Resource)
	if !ok {
		that2, ok := that.(Resource)
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

	if len(m.GetAliases()) != len(target.GetAliases()) {
		return false
	}
	for idx, v := range m.GetAliases() {

		if strings.Compare(v, target.GetAliases()[idx]) != 0 {
			return false
		}

	}

	if strings.Compare(m.GetVersion(), target.GetVersion()) != 0 {
		return false
	}

	if h, ok := interface{}(m.GetResource()).(equality.Equalizer); ok {
		if !h.Equal(target.GetResource()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetResource(), target.GetResource()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetTtl()).(equality.Equalizer); ok {
		if !h.Equal(target.GetTtl()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetTtl(), target.GetTtl()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetCacheControl()).(equality.Equalizer); ok {
		if !h.Equal(target.GetCacheControl()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetCacheControl(), target.GetCacheControl()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *Resource_CacheControl) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*Resource_CacheControl)
	if !ok {
		that2, ok := that.(Resource_CacheControl)
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

	if m.GetDoNotCache() != target.GetDoNotCache() {
		return false
	}

	return true
}
