// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/test/mocks/api/v1/mock_resources.proto

package v1

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
func (m *MockResource) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*MockResource)
	if !ok {
		that2, ok := that.(MockResource)
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

	if h, ok := interface{}(m.GetNamespacedStatuses()).(equality.Equalizer); ok {
		if !h.Equal(target.GetNamespacedStatuses()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetNamespacedStatuses(), target.GetNamespacedStatuses()) {
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

	if strings.Compare(m.GetData(), target.GetData()) != 0 {
		return false
	}

	if strings.Compare(m.GetSomeDumbField(), target.GetSomeDumbField()) != 0 {
		return false
	}

	switch m.TestOneofFields.(type) {

	case *MockResource_OneofOne:
		if _, ok := target.TestOneofFields.(*MockResource_OneofOne); !ok {
			return false
		}

		if strings.Compare(m.GetOneofOne(), target.GetOneofOne()) != 0 {
			return false
		}

	case *MockResource_OneofTwo:
		if _, ok := target.TestOneofFields.(*MockResource_OneofTwo); !ok {
			return false
		}

		if m.GetOneofTwo() != target.GetOneofTwo() {
			return false
		}

	default:
		// m is nil but target is not nil
		if m.TestOneofFields != target.TestOneofFields {
			return false
		}
	}

	switch m.NestedOneofOptions.(type) {

	case *MockResource_OneofNestedoneof:
		if _, ok := target.NestedOneofOptions.(*MockResource_OneofNestedoneof); !ok {
			return false
		}

		if h, ok := interface{}(m.GetOneofNestedoneof()).(equality.Equalizer); ok {
			if !h.Equal(target.GetOneofNestedoneof()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetOneofNestedoneof(), target.GetOneofNestedoneof()) {
				return false
			}
		}

	default:
		// m is nil but target is not nil
		if m.NestedOneofOptions != target.NestedOneofOptions {
			return false
		}
	}

	return true
}

// Equal function
func (m *NestedOneOf) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*NestedOneOf)
	if !ok {
		that2, ok := that.(NestedOneOf)
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

	switch m.Option.(type) {

	case *NestedOneOf_OptionA:
		if _, ok := target.Option.(*NestedOneOf_OptionA); !ok {
			return false
		}

		if strings.Compare(m.GetOptionA(), target.GetOptionA()) != 0 {
			return false
		}

	case *NestedOneOf_OptionB:
		if _, ok := target.Option.(*NestedOneOf_OptionB); !ok {
			return false
		}

		if strings.Compare(m.GetOptionB(), target.GetOptionB()) != 0 {
			return false
		}

	default:
		// m is nil but target is not nil
		if m.Option != target.Option {
			return false
		}
	}

	switch m.AnotherOption.(type) {

	case *NestedOneOf_AnotherOptionA:
		if _, ok := target.AnotherOption.(*NestedOneOf_AnotherOptionA); !ok {
			return false
		}

		if strings.Compare(m.GetAnotherOptionA(), target.GetAnotherOptionA()) != 0 {
			return false
		}

	case *NestedOneOf_AnotherOptionB:
		if _, ok := target.AnotherOption.(*NestedOneOf_AnotherOptionB); !ok {
			return false
		}

		if strings.Compare(m.GetAnotherOptionB(), target.GetAnotherOptionB()) != 0 {
			return false
		}

	default:
		// m is nil but target is not nil
		if m.AnotherOption != target.AnotherOption {
			return false
		}
	}

	switch m.NestedOneof.(type) {

	case *NestedOneOf_AnotherNestedOneofOne:
		if _, ok := target.NestedOneof.(*NestedOneOf_AnotherNestedOneofOne); !ok {
			return false
		}

		if h, ok := interface{}(m.GetAnotherNestedOneofOne()).(equality.Equalizer); ok {
			if !h.Equal(target.GetAnotherNestedOneofOne()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetAnotherNestedOneofOne(), target.GetAnotherNestedOneofOne()) {
				return false
			}
		}

	case *NestedOneOf_AnotherNestedOneofTwo:
		if _, ok := target.NestedOneof.(*NestedOneOf_AnotherNestedOneofTwo); !ok {
			return false
		}

		if h, ok := interface{}(m.GetAnotherNestedOneofTwo()).(equality.Equalizer); ok {
			if !h.Equal(target.GetAnotherNestedOneofTwo()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetAnotherNestedOneofTwo(), target.GetAnotherNestedOneofTwo()) {
				return false
			}
		}

	default:
		// m is nil but target is not nil
		if m.NestedOneof != target.NestedOneof {
			return false
		}
	}

	return true
}

// Equal function
func (m *InternalOneOf) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*InternalOneOf)
	if !ok {
		that2, ok := that.(InternalOneOf)
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

	switch m.Option.(type) {

	case *InternalOneOf_OptionA:
		if _, ok := target.Option.(*InternalOneOf_OptionA); !ok {
			return false
		}

		if strings.Compare(m.GetOptionA(), target.GetOptionA()) != 0 {
			return false
		}

	case *InternalOneOf_OptionB:
		if _, ok := target.Option.(*InternalOneOf_OptionB); !ok {
			return false
		}

		if strings.Compare(m.GetOptionB(), target.GetOptionB()) != 0 {
			return false
		}

	default:
		// m is nil but target is not nil
		if m.Option != target.Option {
			return false
		}
	}

	return true
}

// Equal function
func (m *FakeResource) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*FakeResource)
	if !ok {
		that2, ok := that.(FakeResource)
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

	if m.GetCount() != target.GetCount() {
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

	return true
}

// Equal function
func (m *MockXdsResourceConfig) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*MockXdsResourceConfig)
	if !ok {
		that2, ok := that.(MockXdsResourceConfig)
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

	if strings.Compare(m.GetDomain(), target.GetDomain()) != 0 {
		return false
	}

	return true
}
