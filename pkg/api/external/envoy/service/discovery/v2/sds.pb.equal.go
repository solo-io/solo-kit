// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/api/external/envoy/service/discovery/v2/sds.proto

package v2

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
func (m *SdsDummy) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*SdsDummy)
	if !ok {
		that2, ok := that.(SdsDummy)
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

	return true
}
