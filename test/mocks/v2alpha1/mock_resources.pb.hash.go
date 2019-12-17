// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/test/mocks/api/v2alpha1/mock_resources.proto

package v2alpha1

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"

	"github.com/golang/protobuf/ptypes"
	"github.com/mitchellh/hashstructure"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
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

	if h, ok := interface{}(m.GetMetadata()).(interface {
		Hash(hasher hash.Hash64) (uint64, error)
	}); ok {
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if val, err := hashstructure.Hash(m.GetMetadata(), nil); err != nil {
			return 0, err
		} else {
			if err := binary.Write(hasher, binary.LittleEndian, val); err != nil {
				return 0, err
			}
		}
	}

	switch m.WeStuckItInAOneof.(type) {

	case *MockResource_SomeDumbField:

	case *MockResource_Data:

		if _, err = hasher.Write([]byte(m.GetData())); err != nil {
			return 0, err
		}

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
func (m *FrequentlyChangingAnnotationsResource) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error

	if h, ok := interface{}(m.GetMetadata()).(interface {
		Hash(hasher hash.Hash64) (uint64, error)
	}); ok {
		if _, err = h.Hash(hasher); err != nil {
			return 0, err
		}
	} else {
		if val, err := hashstructure.Hash(m.GetMetadata(), nil); err != nil {
			return 0, err
		} else {
			if err := binary.Write(hasher, binary.LittleEndian, val); err != nil {
				return 0, err
			}
		}
	}

	if _, err = hasher.Write([]byte(m.GetBlah())); err != nil {
		return 0, err
	}

	return hasher.Sum64(), nil
}
