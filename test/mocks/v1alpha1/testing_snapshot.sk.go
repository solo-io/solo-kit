// Code generated by solo-kit. DO NOT EDIT.

package v1alpha1

import (
	"encoding/json"
	"fmt"
	"hash"
	"hash/fnv"
	"log"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var _ json.Marshaler = new(TestingSnapshot)

type TestingSnapshot struct {
	Mocks MockResourceList `json:"mocks"`
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
		log.Println(eris.Wrapf(err, "error hashing, this should never happen"))
	}
	fields = append(fields, zap.Uint64("mocks", MocksHash))
	snapshotHash, err := s.Hash(hasher)
	if err != nil {
		log.Println(eris.Wrapf(err, "error hashing, this should never happen"))
	}
	return append(fields, zap.Uint64("snapshotHash", snapshotHash))
}

func (s TestingSnapshot) MarshalJSON() ([]byte, error) {
	return json.Marshal(&s)

}

func (s *TestingSnapshot) GetResourcesList(resource resources.Resource) (resources.ResourceList, error) {
	switch resource.(type) {
	case *MockResource:
		return s.Mocks.AsResources(), nil
	default:
		return resources.ResourceList{}, eris.New("did not contain the input resource type returning empty list")
	}
}

func (s *TestingSnapshot) RemoveFromResourceList(resource resources.Resource) error {
	refKey := resource.GetMetadata().Ref().Key()
	switch resource.(type) {
	case *MockResource:

		for i, res := range s.Mocks {
			if refKey == res.GetMetadata().Ref().Key() {
				s.Mocks = append(s.Mocks[:i], s.Mocks[i+1:]...)
				break
			}
		}
		return nil
	default:
		return eris.Errorf("did not remove the resource because its type does not exist [%T]", resource)
	}
}

func (s *TestingSnapshot) UpsertToResourceList(resource resources.Resource) error {
	refKey := resource.GetMetadata().Ref().Key()
	switch typed := resource.(type) {
	case *MockResource:
		updated := false
		for i, res := range s.Mocks {
			if refKey == res.GetMetadata().Ref().Key() {
				s.Mocks[i] = typed
				updated = true
			}
		}
		if !updated {
			s.Mocks = append(s.Mocks, typed)
		}
		s.Mocks.Sort()
		return nil
	default:
		return eris.Errorf("did not add/replace the resource type because it does not exist %T", resource)
	}
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
		log.Println(eris.Wrapf(err, "error hashing, this should never happen"))
	}
	return TestingSnapshotStringer{
		Version: snapshotHash,
		Mocks:   s.Mocks.NamespacesDotNames(),
	}
}

var TestingGvkToHashableResource = map[schema.GroupVersionKind]func() resources.HashableResource{
	MockResourceGVK: NewMockResourceHashableResource,
}
