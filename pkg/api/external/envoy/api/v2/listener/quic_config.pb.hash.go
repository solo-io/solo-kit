// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/api/external/envoy/api/v2/listener/quic_config.proto

package listener

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"

	"github.com/mitchellh/hashstructure"
	safe_hasher "github.com/solo-io/protoc-gen-ext/pkg/hasher"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
	_ = binary.LittleEndian
	_ = new(hash.Hash64)
	_ = fnv.New64
	_ = hashstructure.Hash
	_ = new(safe_hasher.SafeHasher)
)

// Hash function
func (m *QuicProtocolOptions) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("solo.io.envoy.api.v2.listener.github.com/solo-io/solo-kit/pkg/api/external/envoy/api/v2/listener.QuicProtocolOptions")); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetMaxConcurrentStreams()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("MaxConcurrentStreams")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetMaxConcurrentStreams(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("MaxConcurrentStreams")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	if h, ok := interface{}(m.GetIdleTimeout()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("IdleTimeout")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetIdleTimeout(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("IdleTimeout")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	if h, ok := interface{}(m.GetCryptoHandshakeTimeout()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("CryptoHandshakeTimeout")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetCryptoHandshakeTimeout(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("CryptoHandshakeTimeout")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	return hasher.Sum64(), nil
}
