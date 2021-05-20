// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/test/mocks/api/v1/mock_resources.proto

package v1

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
func (m *MockResource) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("testing.solo.io.github.com/solo-io/solo-kit/test/mocks/v1.MockResource")); err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetMetadata()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("Metadata")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetMetadata(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("Metadata")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	if _, err = hasher.Write([]byte(m.GetData())); err != nil {
		return 0, err
	}

	switch m.TestOneofFields.(type) {

	case *MockResource_OneofOne:

		if _, err = hasher.Write([]byte(m.GetOneofOne())); err != nil {
			return 0, err
		}

	case *MockResource_OneofTwo:

		err = binary.Write(hasher, binary.LittleEndian, m.GetOneofTwo())
		if err != nil {
			return 0, err
		}

	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *FakeResource) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("testing.solo.io.github.com/solo-io/solo-kit/test/mocks/v1.FakeResource")); err != nil {
		return 0, err
	}

	err = binary.Write(hasher, binary.LittleEndian, m.GetCount())
	if err != nil {
		return 0, err
	}

	if h, ok := interface{}(m.GetMetadata()).(safe_hasher.SafeHasher); ok {
		if _, err = hasher.Write([]byte("Metadata")); err != nil {
			return 0, err
		}
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if fieldValue, err := hashstructure.Hash(m.GetMetadata(), nil); err != nil {
			return 0, err
		} else {
			if _, err = hasher.Write([]byte("Metadata")); err != nil {
				return 0, err
			}
			if err := binary.Write(hasher, binary.LittleEndian, fieldValue); err != nil {
				return 0, err
			}
		}
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *MockXdsResourceConfig) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error
	if _, err = hasher.Write([]byte("testing.solo.io.github.com/solo-io/solo-kit/test/mocks/v1.MockXdsResourceConfig")); err != nil {
		return 0, err
	}

	if _, err = hasher.Write([]byte(m.GetDomain())); err != nil {
		return 0, err
	}

	return hasher.Sum64(), nil
}
