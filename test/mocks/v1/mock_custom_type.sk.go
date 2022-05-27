// Code generated by solo-kit. DO NOT EDIT.

//Source: pkg/code-generator/codegen/templates/resource_template.go
package v1

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"log"
	"sort"

	github_com_solo_io_solo_kit_test_mocks_api_v1_customtype "github.com/solo-io/solo-kit/test/mocks/api/v1/customtype"

	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewMockCustomType(namespace, name string) *MockCustomType {
	mockcustomtype := &MockCustomType{}
	mockcustomtype.MockCustomType.SetMetadata(&core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return mockcustomtype
}

// require custom resource to implement Clone() as well as resources.Resource interface

type CloneableMockCustomType interface {
	resources.Resource
	Clone() *github_com_solo_io_solo_kit_test_mocks_api_v1_customtype.MockCustomType
}

var _ CloneableMockCustomType = &github_com_solo_io_solo_kit_test_mocks_api_v1_customtype.MockCustomType{}

type MockCustomType struct {
	github_com_solo_io_solo_kit_test_mocks_api_v1_customtype.MockCustomType
}

func (r *MockCustomType) Clone() resources.Resource {
	return &MockCustomType{MockCustomType: *r.MockCustomType.Clone()}
}

func (r *MockCustomType) Hash(hasher hash.Hash64) (uint64, error) {
	if hasher == nil {
		hasher = fnv.New64()
	}
	clone := r.MockCustomType.Clone()
	resources.UpdateMetadata(clone, func(meta *core.Metadata) {
		meta.ResourceVersion = ""
	})
	err := binary.Write(hasher, binary.LittleEndian, hashutils.HashAll(clone))
	if err != nil {
		return 0, err
	}
	return hasher.Sum64(), nil
}

func (r *MockCustomType) MustHash() uint64 {
	hashVal, err := r.Hash(nil)
	if err != nil {
		log.Panicf("error while hashing: (%s) this should never happen", err)
	}
	return hashVal
}

func (r *MockCustomType) GroupVersionKind() schema.GroupVersionKind {
	return MockCustomTypeGVK
}

type MockCustomTypeList []*MockCustomType

func (list MockCustomTypeList) Find(namespace, name string) (*MockCustomType, error) {
	for _, mockCustomType := range list {
		if mockCustomType.GetMetadata().Name == name && mockCustomType.GetMetadata().Namespace == namespace {
			return mockCustomType, nil
		}
	}
	return nil, errors.Errorf("list did not find mockCustomType %v.%v", namespace, name)
}

func (list MockCustomTypeList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, mockCustomType := range list {
		ress = append(ress, mockCustomType)
	}
	return ress
}

func (list MockCustomTypeList) Names() []string {
	var names []string
	for _, mockCustomType := range list {
		names = append(names, mockCustomType.GetMetadata().Name)
	}
	return names
}

func (list MockCustomTypeList) NamespacesDotNames() []string {
	var names []string
	for _, mockCustomType := range list {
		names = append(names, mockCustomType.GetMetadata().Namespace+"."+mockCustomType.GetMetadata().Name)
	}
	return names
}

func (list MockCustomTypeList) Sort() MockCustomTypeList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list MockCustomTypeList) Clone() MockCustomTypeList {
	var mockCustomTypeList MockCustomTypeList
	for _, mockCustomType := range list {
		mockCustomTypeList = append(mockCustomTypeList, resources.Clone(mockCustomType).(*MockCustomType))
	}
	return mockCustomTypeList
}

func (list MockCustomTypeList) Each(f func(element *MockCustomType)) {
	for _, mockCustomType := range list {
		f(mockCustomType)
	}
}

func (list MockCustomTypeList) EachResource(f func(element resources.Resource)) {
	for _, mockCustomType := range list {
		f(mockCustomType)
	}
}

func (list MockCustomTypeList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *MockCustomType) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

var (
	MockCustomTypeGVK = schema.GroupVersionKind{
		Version: "v1",
		Group:   "testing.solo.io",
		Kind:    "MockCustomType",
	}
)
