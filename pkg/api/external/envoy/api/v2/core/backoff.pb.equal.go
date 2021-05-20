// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/api/external/envoy/api/v2/core/backoff.proto

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
func (m *BackoffStrategy) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*BackoffStrategy)
	if !ok {
		that2, ok := that.(BackoffStrategy)
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

	if h, ok := interface{}(m.GetBaseInterval()).(equality.Equalizer); ok {
		if !h.Equal(target.GetBaseInterval()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetBaseInterval(), target.GetBaseInterval()) {
			return false
		}
	}

	if h, ok := interface{}(m.GetMaxInterval()).(equality.Equalizer); ok {
		if !h.Equal(target.GetMaxInterval()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetMaxInterval(), target.GetMaxInterval()) {
			return false
		}
	}

	return true
}
