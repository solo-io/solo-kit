// Code generated by solo-kit. DO NOT EDIT.

package v2alpha1

import (
	"fmt"
	"hash"
	"hash/fnv"
	"log"

	testing_solo_io "github.com/solo-io/solo-kit/test/mocks/v1"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type TestingSnapshot struct {
	Mocks MockResourceList
	Fcars FrequentlyChangingAnnotationsResourceList
	Fakes testing_solo_io.FakeResourceList
}

func (s TestingSnapshot) Clone() TestingSnapshot {
	return TestingSnapshot{
		Mocks: s.Mocks.Clone(),
		Fcars: s.Fcars.Clone(),
		Fakes: s.Fakes.Clone(),
	}
}

func (s TestingSnapshot) Hash(hasher hash.Hash64) (uint64, error) {
	if hasher == nil {
		hasher = fnv.New64()
	}
	if _, err := s.hashMocks(hasher); err != nil {
		return 0, err
	}
	if _, err := s.hashFcars(hasher); err != nil {
		return 0, err
	}
	if _, err := s.hashFakes(hasher); err != nil {
		return 0, err
	}
	return hasher.Sum64(), nil
}

func (s TestingSnapshot) hashMocks(hasher hash.Hash64) (uint64, error) {
	return hashutils.HashAllSafe(hasher, s.Mocks.AsInterfaces()...)
}

func (s TestingSnapshot) hashFcars(hasher hash.Hash64) (uint64, error) {
	clonedList := s.Fcars.Clone()
	for _, v := range clonedList {
		v.Metadata.Annotations = nil
	}
	return hashutils.HashAllSafe(hasher, clonedList.AsInterfaces()...)
}

func (s TestingSnapshot) hashFakes(hasher hash.Hash64) (uint64, error) {
	return hashutils.HashAllSafe(hasher, s.Fakes.AsInterfaces()...)
}

func (s TestingSnapshot) HashFields() []zap.Field {
	var fields []zap.Field
	hasher := fnv.New64()
	MocksHash, err := s.hashMocks(hasher)
	if err != nil {
		log.Println(eris.Wrapf(err, "error hashing, this should never happen"))
	}
	fields = append(fields, zap.Uint64("mocks", MocksHash))
	FcarsHash, err := s.hashFcars(hasher)
	if err != nil {
		log.Println(eris.Wrapf(err, "error hashing, this should never happen"))
	}
	fields = append(fields, zap.Uint64("fcars", FcarsHash))
	FakesHash, err := s.hashFakes(hasher)
	if err != nil {
		log.Println(eris.Wrapf(err, "error hashing, this should never happen"))
	}
	fields = append(fields, zap.Uint64("fakes", FakesHash))
	snapshotHash, err := s.Hash(hasher)
	if err != nil {
		log.Println(eris.Wrapf(err, "error hashing, this should never happen"))
	}
	return append(fields, zap.Uint64("snapshotHash", snapshotHash))
}

func (s *TestingSnapshot) GetResourcesList(resource resources.Resource) (resources.ResourceList, error) {
	switch resource.(type) {
	case *MockResource:
		return s.Mocks.AsResources(), nil
	case *FrequentlyChangingAnnotationsResource:
		return s.Fcars.AsResources(), nil
	case *testing_solo_io.FakeResource:
		return s.Fakes.AsResources(), nil
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
	case *FrequentlyChangingAnnotationsResource:

		for i, res := range s.Fcars {
			if refKey == res.GetMetadata().Ref().Key() {
				s.Fcars = append(s.Fcars[:i], s.Fcars[i+1:]...)
				break
			}
		}
		return nil
	case *testing_solo_io.FakeResource:

		for i, res := range s.Fakes {
			if refKey == res.GetMetadata().Ref().Key() {
				s.Fakes = append(s.Fakes[:i], s.Fakes[i+1:]...)
				break
			}
		}
		return nil
	default:
		return eris.Errorf("did not remove the resource because its type does not exist [%T]", resource)
	}
}

func (s *TestingSnapshot) RemoveAllResourcesInNamespace(namespace string) {
	var Mocks MockResourceList
	for _, res := range s.Mocks {
		if namespace != res.GetMetadata().GetNamespace() {
			Mocks = append(Mocks, res)
		}
	}
	s.Mocks = Mocks
	var Fcars FrequentlyChangingAnnotationsResourceList
	for _, res := range s.Fcars {
		if namespace != res.GetMetadata().GetNamespace() {
			Fcars = append(Fcars, res)
		}
	}
	s.Fcars = Fcars
	var Fakes testing_solo_io.FakeResourceList
	for _, res := range s.Fakes {
		if namespace != res.GetMetadata().GetNamespace() {
			Fakes = append(Fakes, res)
		}
	}
	s.Fakes = Fakes
}

type Predicate func(metadata *core.Metadata) bool

func (s *TestingSnapshot) RemoveMatches(predicate Predicate) {
	var Mocks MockResourceList
	for _, res := range s.Mocks {
		if matches := predicate(res.GetMetadata()); !matches {
			Mocks = append(Mocks, res)
		}
	}
	s.Mocks = Mocks
	var Fcars FrequentlyChangingAnnotationsResourceList
	for _, res := range s.Fcars {
		if matches := predicate(res.GetMetadata()); !matches {
			Fcars = append(Fcars, res)
		}
	}
	s.Fcars = Fcars
	var Fakes testing_solo_io.FakeResourceList
	for _, res := range s.Fakes {
		if matches := predicate(res.GetMetadata()); !matches {
			Fakes = append(Fakes, res)
		}
	}
	s.Fakes = Fakes
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
	case *FrequentlyChangingAnnotationsResource:
		updated := false
		for i, res := range s.Fcars {
			if refKey == res.GetMetadata().Ref().Key() {
				s.Fcars[i] = typed
				updated = true
			}
		}
		if !updated {
			s.Fcars = append(s.Fcars, typed)
		}
		s.Fcars.Sort()
		return nil
	case *testing_solo_io.FakeResource:
		updated := false
		for i, res := range s.Fakes {
			if refKey == res.GetMetadata().Ref().Key() {
				s.Fakes[i] = typed
				updated = true
			}
		}
		if !updated {
			s.Fakes = append(s.Fakes, typed)
		}
		s.Fakes.Sort()
		return nil
	default:
		return eris.Errorf("did not add/replace the resource type because it does not exist %T", resource)
	}
}

type TestingSnapshotStringer struct {
	Version uint64
	Mocks   []string
	Fcars   []string
	Fakes   []string
}

func (ss TestingSnapshotStringer) String() string {
	s := fmt.Sprintf("TestingSnapshot %v\n", ss.Version)

	s += fmt.Sprintf("  Mocks %v\n", len(ss.Mocks))
	for _, name := range ss.Mocks {
		s += fmt.Sprintf("    %v\n", name)
	}

	s += fmt.Sprintf("  Fcars %v\n", len(ss.Fcars))
	for _, name := range ss.Fcars {
		s += fmt.Sprintf("    %v\n", name)
	}

	s += fmt.Sprintf("  Fakes %v\n", len(ss.Fakes))
	for _, name := range ss.Fakes {
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
		Fcars:   s.Fcars.NamespacesDotNames(),
		Fakes:   s.Fakes.NamespacesDotNames(),
	}
}

var TestingGvkToHashableResource = map[schema.GroupVersionKind]func() resources.HashableResource{
	MockResourceGVK:                          NewMockResourceHashableResource,
	FrequentlyChangingAnnotationsResourceGVK: NewFrequentlyChangingAnnotationsResourceHashableResource,
	testing_solo_io.FakeResourceGVK:          testing_solo_io.NewFakeResourceHashableResource,
}
