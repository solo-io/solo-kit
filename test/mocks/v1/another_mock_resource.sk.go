// Code generated by solo-kit. DO NOT EDIT.

package v1

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

func NewAnotherMockResource(namespace, name string) *AnotherMockResource {
	anothermockresource := &AnotherMockResource{}
	anothermockresource.SetMetadata(core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return anothermockresource
}

func (r *AnotherMockResource) SetMetadata(meta core.Metadata) {
	r.Metadata = meta
}

func (r *AnotherMockResource) SetStatus(status core.Status) {
	r.Status = status
}

func (r *AnotherMockResource) Hash() uint64 {
	metaCopy := r.GetMetadata()
	metaCopy.ResourceVersion = ""
	return hashutils.HashAll(
		metaCopy,
		r.BasicField,
	)
}

type AnotherMockResourceList []*AnotherMockResource

// namespace is optional, if left empty, names can collide if the list contains more than one with the same name
func (list AnotherMockResourceList) Find(namespace, name string) (*AnotherMockResource, error) {
	for _, anotherMockResource := range list {
		if anotherMockResource.GetMetadata().Name == name {
			if namespace == "" || anotherMockResource.GetMetadata().Namespace == namespace {
				return anotherMockResource, nil
			}
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

var _ resources.Resource = &AnotherMockResource{}

// Kubernetes Adapter for AnotherMockResource

func (o *AnotherMockResource) GetObjectKind() schema.ObjectKind {
	t := AnotherMockResourceCrd.TypeMeta()
	return &t
}

func (o *AnotherMockResource) DeepCopyObject() runtime.Object {
	return resources.Clone(o).(*AnotherMockResource)
}

var (
	AnotherMockResourceGVK = schema.GroupVersionKind{
		Version: "v1",
		Group:   "testing.solo.io",
		Kind:    "AnotherMockResource",
	}
	AnotherMockResourceCrd = crd.NewCrd(
		"anothermockresources",
		AnotherMockResourceGVK.Group,
		AnotherMockResourceGVK.Version,
		AnotherMockResourceGVK.Kind,
		"amr",
		false,
		&AnotherMockResource{})
)

func init() {
	if err := crd.GetRegistry().AddCrd(AnotherMockResourceCrd); err != nil {
		log.Fatalf("could not add crd to global registry")
	}
}
