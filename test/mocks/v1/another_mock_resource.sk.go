// Code generated by solo-kit. DO NOT EDIT.

package v1

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

func NewAnotherMockResource(namespace, name string) *AnotherMockResource {
	anothermockresource := &AnotherMockResource{}
	anothermockresource.SetMetadata(&core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return anothermockresource
}

func (r *AnotherMockResource) SetMetadata(meta *core.Metadata) {
	r.Metadata = meta
}

// Deprecated
func (r *AnotherMockResource) SetStatus(status *core.Status) {
	r.SetStatusForNamespace(status)
}

// Deprecated
func (r *AnotherMockResource) GetStatus() *core.Status {
	if r != nil {
		s, _ := r.GetStatusForNamespace()
		return s
	}
	return nil
}

func (r *AnotherMockResource) SetNamespacedStatuses(statuses *core.NamespacedStatuses) {
	r.NamespacedStatuses = statuses
}

// SetStatusForNamespace inserts the specified status into the NamespacedStatuses.Statuses map for
// the current namespace (as specified by POD_NAMESPACE env var).  If the resource does not yet
// have a NamespacedStatuses, one will be created.
// Note: POD_NAMESPACE environment variable must be set for this function to behave as expected.
// If unset, a podNamespaceErr is returned.
func (r *AnotherMockResource) SetStatusForNamespace(status *core.Status) error {
	return statusutils.SetStatusForNamespace(r, status)
}

// GetStatusForNamespace returns the status stored in the NamespacedStatuses.Statuses map for the
// controller specified by the POD_NAMESPACE env var, or nil if no status exists for that
// controller.
// Note: POD_NAMESPACE environment variable must be set for this function to behave as expected.
// If unset, a podNamespaceErr is returned.
func (r *AnotherMockResource) GetStatusForNamespace() (*core.Status, error) {
	return statusutils.GetStatusForNamespace(r)
}

func (r *AnotherMockResource) MustHash() uint64 {
	hashVal, err := r.Hash(nil)
	if err != nil {
		log.Panicf("error while hashing: (%s) this should never happen", err)
	}
	return hashVal
}

func (r *AnotherMockResource) GroupVersionKind() schema.GroupVersionKind {
	return AnotherMockResourceGVK
}

type AnotherMockResourceList []*AnotherMockResource

func (list AnotherMockResourceList) Find(namespace, name string) (*AnotherMockResource, error) {
	for _, anotherMockResource := range list {
		if anotherMockResource.GetMetadata().Name == name && anotherMockResource.GetMetadata().Namespace == namespace {
			return anotherMockResource, nil
		}
	}
	return nil, errors.Errorf("list did not find anotherMockResource %v.%v", namespace, name)
}

func (list AnotherMockResourceList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, anotherMockResource := range list {
		ress = append(ress, anotherMockResource)
	}
	return ress
}

func (list AnotherMockResourceList) AsInputResources() resources.InputResourceList {
	var ress resources.InputResourceList
	for _, anotherMockResource := range list {
		ress = append(ress, anotherMockResource)
	}
	return ress
}

func (list AnotherMockResourceList) Names() []string {
	var names []string
	for _, anotherMockResource := range list {
		names = append(names, anotherMockResource.GetMetadata().Name)
	}
	return names
}

func (list AnotherMockResourceList) NamespacesDotNames() []string {
	var names []string
	for _, anotherMockResource := range list {
		names = append(names, anotherMockResource.GetMetadata().Namespace+"."+anotherMockResource.GetMetadata().Name)
	}
	return names
}

func (list AnotherMockResourceList) Sort() AnotherMockResourceList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list AnotherMockResourceList) Clone() AnotherMockResourceList {
	var anotherMockResourceList AnotherMockResourceList
	for _, anotherMockResource := range list {
		anotherMockResourceList = append(anotherMockResourceList, resources.Clone(anotherMockResource).(*AnotherMockResource))
	}
	return anotherMockResourceList
}

func (list AnotherMockResourceList) Each(f func(element *AnotherMockResource)) {
	for _, anotherMockResource := range list {
		f(anotherMockResource)
	}
}

func (list AnotherMockResourceList) EachResource(f func(element resources.Resource)) {
	for _, anotherMockResource := range list {
		f(anotherMockResource)
	}
}

func (list AnotherMockResourceList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *AnotherMockResource) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

// Kubernetes Adapter for AnotherMockResource

func (o *AnotherMockResource) GetObjectKind() schema.ObjectKind {
	t := AnotherMockResourceCrd.TypeMeta()
	return &t
}

func (o *AnotherMockResource) DeepCopyObject() runtime.Object {
	return resources.Clone(o).(*AnotherMockResource)
}

func (o *AnotherMockResource) DeepCopyInto(out *AnotherMockResource) {
	clone := resources.Clone(o).(*AnotherMockResource)
	*out = *clone
}

var (
	AnotherMockResourceCrd = crd.NewCrd(
		"anothermockresources",
		AnotherMockResourceGVK.Group,
		AnotherMockResourceGVK.Version,
		AnotherMockResourceGVK.Kind,
		"amr",
		false,
		&AnotherMockResource{})
)

var (
	AnotherMockResourceGVK = schema.GroupVersionKind{
		Version: "v1",
		Group:   "testing.solo.io",
		Kind:    "AnotherMockResource",
	}
)
