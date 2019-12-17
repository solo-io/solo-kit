// Code generated by protoc-gen-ext. DO NOT EDIT.
// source: github.com/solo-io/solo-kit/api/external/envoy/type/percent.proto

package _type

import (
	"encoding/binary"
	"errors"
	"fmt"
	"hash"
	"hash/fnv"
)

// ensure the imports are used
var (
	_ = errors.New("")
	_ = fmt.Print
)

// Hash function
func (m *Percent) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error

	err = binary.Write(hasher, binary.LittleEndian, m.GetValue())
	if err != nil {
		return 0, err
	}

	return hasher.Sum64(), nil
}

// Hash function
func (m *FractionalPercent) Hash(hasher hash.Hash64) (uint64, error) {
	if m == nil {
		return 0, nil
	}
	if hasher == nil {
		hasher = fnv.New64()
	}
	var err error

	err = binary.Write(hasher, binary.LittleEndian, m.GetNumerator())
	if err != nil {
		return 0, err
	}

	err = binary.Write(hasher, binary.LittleEndian, m.GetDenominator())
	if err != nil {
		return 0, err
	}

	return hasher.Sum64(), nil
}