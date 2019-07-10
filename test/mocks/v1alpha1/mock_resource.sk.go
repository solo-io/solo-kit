// Code generated by solo-kit. DO NOT EDIT.

package v1alpha1

import (
	"log"
	"sort"

	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewMockResource(namespace, name string) *MockResource {
	mockresource := &MockResource{}
	mockresource.SetMetadata(core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return mockresource
}

func (r *MockResource) SetMetadata(meta core.Metadata) {
	r.Metadata = meta
}

func (r *MockResource) SetStatus(status core.Status) {
	r.Status = status
}

func (r *MockResource) Hash() uint64 {
	metaCopy := r.GetMetadata()
	metaCopy.ResourceVersion = ""
	return hashutils.HashAll(
		metaCopy,
		r.Data,
		r.TestOneofFields,
	)
}

type MockResourceList []*MockResource

// namespace is optional, if left empty, names can collide if the list contains more than one with the same name
func (list MockResourceList) Find(namespace, name string) (*MockResource, error) {
	for _, mockResource := range list {
		if mockResource.GetMetadata().Name == name {
			if namespace == "" || mockResource.GetMetadata().Namespace == namespace {
				return mockResource, nil
			}
		}
	}
	return nil, errors.Errorf("list did not find mockResource %v.%v", namespace, name)
}

func (list MockResourceList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, mockResource := range list {
		ress = append(ress, mockResource)
	}
	return ress
}

func (list MockResourceList) AsInputResources() resources.InputResourceList {
	var ress resources.InputResourceList
	for _, mockResource := range list {
		ress = append(ress, mockResource)
	}
	return ress
}

func (list MockResourceList) Names() []string {
	var names []string
	for _, mockResource := range list {
		names = append(names, mockResource.GetMetadata().Name)
	}
	return names
}

func (list MockResourceList) NamespacesDotNames() []string {
	var names []string
	for _, mockResource := range list {
		names = append(names, mockResource.GetMetadata().Namespace+"."+mockResource.GetMetadata().Name)
	}
	return names
}

func (list MockResourceList) Sort() MockResourceList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list MockResourceList) Clone() MockResourceList {
	var mockResourceList MockResourceList
	for _, mockResource := range list {
		mockResourceList = append(mockResourceList, resources.Clone(mockResource).(*MockResource))
	}
	return mockResourceList
}

func (list MockResourceList) Each(f func(element *MockResource)) {
	for _, mockResource := range list {
		f(mockResource)
	}
}

func (list MockResourceList) EachResource(f func(element resources.Resource)) {
	for _, mockResource := range list {
		f(mockResource)
	}
}

func (list MockResourceList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *MockResource) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

var _ resources.Resource = &MockResource{}

// Kubernetes Adapter for MockResource

func (o *MockResource) GetObjectKind() schema.ObjectKind {
	t := MockResourceCrd.TypeMeta()
	return &t
}

func (o *MockResource) DeepCopyObject() runtime.Object {
	return resources.Clone(o).(*MockResource)
}

var (
	MockResourceGVK = schema.GroupVersionKind{
		Version: "v1alpha1",
		Group:   "crds.testing.solo.io",
		Kind:    "MockResource",
	}
	MockResourceCrd = crd.NewCrd(
		"mocks",
		MockResourceGVK.Group,
		MockResourceGVK.Version,
		MockResourceGVK.Kind,
		"mk",
		false,
		&MockResource{})
)

func init() {
	if err := crd.AddCrd(MockResourceCrd); err != nil {
		log.Fatalf("could not add crd to global registry")
	}
}