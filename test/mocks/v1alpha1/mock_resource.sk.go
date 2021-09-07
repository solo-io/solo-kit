// Code generated by solo-kit. DO NOT EDIT.

package v1alpha1

import (
	"log"
	"sort"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/utils/statusutils"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewMockResource(namespace, name string) *MockResource {
	mockresource := &MockResource{}
	mockresource.SetMetadata(&core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return mockresource
}

func (r *MockResource) SetMetadata(meta *core.Metadata) {
	r.Metadata = meta
}

// Deprecated
func (r *MockResource) SetStatus(status *core.Status) {
	r.SetStatusForNamespace(status)
}

// Deprecated
func (r *MockResource) GetStatus() *core.Status {
	if r != nil {
		s, _ := r.GetStatusForNamespace()
		return s
	}
	return nil
}

func (r *MockResource) SetNamespacedStatuses(statuses *core.NamespacedStatuses) {
	r.NamespacedStatuses = statuses
}

// SetStatusForNamespace inserts the specified status into the NamespacedStatuses.Statuses map for
// the current namespace (as specified by POD_NAMESPACE env var).  If the resource does not yet
// have a NamespacedStatuses, one will be created.
// Note: POD_NAMESPACE environment variable must be set for this function to behave as expected.
// If unset, a podNamespaceErr is returned.
func (r *MockResource) SetStatusForNamespace(status *core.Status) error {
	return statusutils.SetStatusForNamespace(r, status)
}

// GetStatusForNamespace returns the status stored in the NamespacedStatuses.Statuses map for the
// controller specified by the POD_NAMESPACE env var, or nil if no status exists for that
// controller.
// Note: POD_NAMESPACE environment variable must be set for this function to behave as expected.
// If unset, a podNamespaceErr is returned.
func (r *MockResource) GetStatusForNamespace() (*core.Status, error) {
	return statusutils.GetStatusForNamespace(r)
}

func (r *MockResource) MustHash() uint64 {
	hashVal, err := r.Hash(nil)
	if err != nil {
		log.Panicf("error while hashing: (%s) this should never happen", err)
	}
	return hashVal
}

func (r *MockResource) GroupVersionKind() schema.GroupVersionKind {
	return MockResourceGVK
}

type MockResourceList []*MockResource

func (list MockResourceList) Find(namespace, name string) (*MockResource, error) {
	for _, mockResource := range list {
		if mockResource.GetMetadata().Name == name && mockResource.GetMetadata().Namespace == namespace {
			return mockResource, nil
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

// Kubernetes Adapter for MockResource

func (o *MockResource) GetObjectKind() schema.ObjectKind {
	t := MockResourceCrd.TypeMeta()
	return &t
}

func (o *MockResource) DeepCopyObject() runtime.Object {
	return resources.Clone(o).(*MockResource)
}

func (o *MockResource) DeepCopyInto(out *MockResource) {
	clone := resources.Clone(o).(*MockResource)
	*out = *clone
}

var (
	MockResourceCrd = crd.NewCrd(
		"mocks",
		MockResourceGVK.Group,
		MockResourceGVK.Version,
		MockResourceGVK.Kind,
		"mk",
		false,
		&MockResource{})
)

var (
	MockResourceGVK = schema.GroupVersionKind{
		Version: "v1alpha1",
		Group:   "crds.testing.solo.io",
		Kind:    "MockResource",
	}
)
