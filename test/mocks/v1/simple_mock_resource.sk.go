// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"log"
	"sort"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// Compile-time assertion
	_ resources.Resource = new(SimpleMockResource)
)

func NewSimpleMockResourceHashableResource() resources.HashableResource {
	return new(SimpleMockResource)
}

func NewSimpleMockResource(namespace, name string) *SimpleMockResource {
	simplemockresource := &SimpleMockResource{}
	simplemockresource.SetMetadata(&core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return simplemockresource
}

func (r *SimpleMockResource) SetMetadata(meta *core.Metadata) {
	r.Metadata = meta
}

func (r *SimpleMockResource) MustHash() uint64 {
	hashVal, err := r.Hash(nil)
	if err != nil {
		log.Panicf("error while hashing: (%s) this should never happen", err)
	}
	return hashVal
}

func (r *SimpleMockResource) GroupVersionKind() schema.GroupVersionKind {
	return SimpleMockResourceGVK
}

type SimpleMockResourceList []*SimpleMockResource

func (list SimpleMockResourceList) Find(namespace, name string) (*SimpleMockResource, error) {
	for _, simpleMockResource := range list {
		if simpleMockResource.GetMetadata().Name == name && simpleMockResource.GetMetadata().Namespace == namespace {
			return simpleMockResource, nil
		}
	}
	return nil, errors.Errorf("list did not find simpleMockResource %v.%v", namespace, name)
}

func (list SimpleMockResourceList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, simpleMockResource := range list {
		ress = append(ress, simpleMockResource)
	}
	return ress
}

func (list SimpleMockResourceList) Names() []string {
	var names []string
	for _, simpleMockResource := range list {
		names = append(names, simpleMockResource.GetMetadata().Name)
	}
	return names
}

func (list SimpleMockResourceList) NamespacesDotNames() []string {
	var names []string
	for _, simpleMockResource := range list {
		names = append(names, simpleMockResource.GetMetadata().Namespace+"."+simpleMockResource.GetMetadata().Name)
	}
	return names
}

func (list SimpleMockResourceList) Sort() SimpleMockResourceList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list SimpleMockResourceList) Clone() SimpleMockResourceList {
	var simpleMockResourceList SimpleMockResourceList
	for _, simpleMockResource := range list {
		simpleMockResourceList = append(simpleMockResourceList, resources.Clone(simpleMockResource).(*SimpleMockResource))
	}
	return simpleMockResourceList
}

func (list SimpleMockResourceList) Each(f func(element *SimpleMockResource)) {
	for _, simpleMockResource := range list {
		f(simpleMockResource)
	}
}

func (list SimpleMockResourceList) EachResource(f func(element resources.Resource)) {
	for _, simpleMockResource := range list {
		f(simpleMockResource)
	}
}

func (list SimpleMockResourceList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *SimpleMockResource) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

// Kubernetes Adapter for SimpleMockResource

func (o *SimpleMockResource) GetObjectKind() schema.ObjectKind {
	t := SimpleMockResourceCrd.TypeMeta()
	return &t
}

func (o *SimpleMockResource) DeepCopyObject() runtime.Object {
	return resources.Clone(o).(*SimpleMockResource)
}

func (o *SimpleMockResource) DeepCopyInto(out *SimpleMockResource) {
	clone := resources.Clone(o).(*SimpleMockResource)
	*out = *clone
}

var (
	SimpleMockResourceCrd = crd.NewCrd(
		"simplemocks",
		SimpleMockResourceGVK.Group,
		SimpleMockResourceGVK.Version,
		SimpleMockResourceGVK.Kind,
		"smk",
		false,
		&SimpleMockResource{})
)

var (
	SimpleMockResourceGVK = schema.GroupVersionKind{
		Version: "v1",
		Group:   "testing.solo.io",
		Kind:    "SimpleMockResource",
	}
)
