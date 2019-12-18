// Code generated by solo-kit. DO NOT EDIT.

package v1alpha1

import (
	"fmt"
	"hash"
	"hash/fnv"
	"log"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/hashutils"
	"go.uber.org/zap"
)

type TestingSnapshot struct {
	Mocks MockResourceList
}

func (s TestingSnapshot) Clone() TestingSnapshot {
	return TestingSnapshot{
		Mocks: s.Mocks.Clone(),
	}
}

func (s TestingSnapshot) Hash(hasher hash.Hash64) (uint64, error) {
	if hasher == nil {
		hasher = fnv.New64()
	}
	if _, err := s.hashMocks(hasher); err != nil {
		return 0, err
	}
	return hasher.Sum64(), nil
}

func (s TestingSnapshot) hashMocks(hasher hash.Hash64) (uint64, error) {
	return hashutils.HashAllSafe(hasher, s.Mocks.AsInterfaces()...)
}

func (s TestingSnapshot) HashFields() []zap.Field {
	var fields []zap.Field
	hasher := fnv.New64()
	MocksHash, err := s.hashMocks(hasher)
	if err != nil {
		log.Println(errors.Wrapf(err, "error hashing, this should never happen"))
	}
	fields = append(fields, zap.Uint64("mocks", MocksHash))
	snapshotHash, err := s.Hash(hasher)
	if err != nil {
		log.Println(errors.Wrapf(err, "error hashing, this should never happen"))
	}
	return append(fields, zap.Uint64("snapshotHash", snapshotHash))
}

type TestingSnapshotStringer struct {
	Version uint64
	Mocks   []string
}

func (ss TestingSnapshotStringer) String() string {
	s := fmt.Sprintf("TestingSnapshot %v\n", ss.Version)

	s += fmt.Sprintf("  Mocks %v\n", len(ss.Mocks))
	for _, name := range ss.Mocks {
		s += fmt.Sprintf("    %v\n", name)
	}

	return s
}

func (s TestingSnapshot) Stringer() TestingSnapshotStringer {
	snapshotHash, err := s.Hash(nil)
	if err != nil {
		log.Println(errors.Wrapf(err, "error hashing, this should never happen"))
	}
	return TestingSnapshotStringer{
		Version: snapshotHash,
		Mocks:   s.Mocks.NamespacesDotNames(),
	}
}
