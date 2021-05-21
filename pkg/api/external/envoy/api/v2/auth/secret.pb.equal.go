// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/api/external/envoy/api/v2/auth/secret.proto

package auth

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
func (m *GenericSecret) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*GenericSecret)
	if !ok {
		that2, ok := that.(GenericSecret)
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

	if h, ok := interface{}(m.GetSecret()).(equality.Equalizer); ok {
		if !h.Equal(target.GetSecret()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetSecret(), target.GetSecret()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *SdsSecretConfig) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*SdsSecretConfig)
	if !ok {
		that2, ok := that.(SdsSecretConfig)
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

	if h, ok := interface{}(m.GetSdsConfig()).(equality.Equalizer); ok {
		if !h.Equal(target.GetSdsConfig()) {
			return false
		}
	} else {
		if !proto.Equal(m.GetSdsConfig(), target.GetSdsConfig()) {
			return false
		}
	}

	return true
}

// Equal function
func (m *Secret) Equal(that interface{}) bool {
	if that == nil {
		return m == nil
	}

	target, ok := that.(*Secret)
	if !ok {
		that2, ok := that.(Secret)
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

	switch m.Type.(type) {

	case *Secret_TlsCertificate:

		if h, ok := interface{}(m.GetTlsCertificate()).(equality.Equalizer); ok {
			if !h.Equal(target.GetTlsCertificate()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetTlsCertificate(), target.GetTlsCertificate()) {
				return false
			}
		}

	case *Secret_SessionTicketKeys:

		if h, ok := interface{}(m.GetSessionTicketKeys()).(equality.Equalizer); ok {
			if !h.Equal(target.GetSessionTicketKeys()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetSessionTicketKeys(), target.GetSessionTicketKeys()) {
				return false
			}
		}

	case *Secret_ValidationContext:

		if h, ok := interface{}(m.GetValidationContext()).(equality.Equalizer); ok {
			if !h.Equal(target.GetValidationContext()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetValidationContext(), target.GetValidationContext()) {
				return false
			}
		}

	case *Secret_GenericSecret:

		if h, ok := interface{}(m.GetGenericSecret()).(equality.Equalizer); ok {
			if !h.Equal(target.GetGenericSecret()) {
				return false
			}
		} else {
			if !proto.Equal(m.GetGenericSecret(), target.GetGenericSecret()) {
				return false
			}
		}

	}

	return true
}
