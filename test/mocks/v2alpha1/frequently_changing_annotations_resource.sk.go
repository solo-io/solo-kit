// Code generated by solo-kit. DO NOT EDIT.

package v2alpha1

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
	_ resources.Resource = new(FrequentlyChangingAnnotationsResource)
)

func NewFrequentlyChangingAnnotationsResourceHashableResource() resources.HashableResource {
	return new(FrequentlyChangingAnnotationsResource)
}

func NewFrequentlyChangingAnnotationsResource(namespace, name string) *FrequentlyChangingAnnotationsResource {
	frequentlychangingannotationsresource := &FrequentlyChangingAnnotationsResource{}
	frequentlychangingannotationsresource.SetMetadata(&core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return frequentlychangingannotationsresource
}

func (r *FrequentlyChangingAnnotationsResource) SetMetadata(meta *core.Metadata) {
	r.Metadata = meta
}

func (r *FrequentlyChangingAnnotationsResource) MustHash() uint64 {
	hashVal, err := r.Hash(nil)
	if err != nil {
		log.Panicf("error while hashing: (%s) this should never happen", err)
	}
	return hashVal
}

func (r *FrequentlyChangingAnnotationsResource) GroupVersionKind() schema.GroupVersionKind {
	return FrequentlyChangingAnnotationsResourceGVK
}

type FrequentlyChangingAnnotationsResourceList []*FrequentlyChangingAnnotationsResource

func (list FrequentlyChangingAnnotationsResourceList) Find(namespace, name string) (*FrequentlyChangingAnnotationsResource, error) {
	for _, frequentlyChangingAnnotationsResource := range list {
		if frequentlyChangingAnnotationsResource.GetMetadata().Name == name && frequentlyChangingAnnotationsResource.GetMetadata().Namespace == namespace {
			return frequentlyChangingAnnotationsResource, nil
		}
	}
	return nil, errors.Errorf("list did not find frequentlyChangingAnnotationsResource %v.%v", namespace, name)
}

func (list FrequentlyChangingAnnotationsResourceList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, frequentlyChangingAnnotationsResource := range list {
		ress = append(ress, frequentlyChangingAnnotationsResource)
	}
	return ress
}

func (list FrequentlyChangingAnnotationsResourceList) Names() []string {
	var names []string
	for _, frequentlyChangingAnnotationsResource := range list {
		names = append(names, frequentlyChangingAnnotationsResource.GetMetadata().Name)
	}
	return names
}

func (list FrequentlyChangingAnnotationsResourceList) NamespacesDotNames() []string {
	var names []string
	for _, frequentlyChangingAnnotationsResource := range list {
		names = append(names, frequentlyChangingAnnotationsResource.GetMetadata().Namespace+"."+frequentlyChangingAnnotationsResource.GetMetadata().Name)
	}
	return names
}

func (list FrequentlyChangingAnnotationsResourceList) Sort() FrequentlyChangingAnnotationsResourceList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list FrequentlyChangingAnnotationsResourceList) Clone() FrequentlyChangingAnnotationsResourceList {
	var frequentlyChangingAnnotationsResourceList FrequentlyChangingAnnotationsResourceList
	for _, frequentlyChangingAnnotationsResource := range list {
		frequentlyChangingAnnotationsResourceList = append(frequentlyChangingAnnotationsResourceList, resources.Clone(frequentlyChangingAnnotationsResource).(*FrequentlyChangingAnnotationsResource))
	}
	return frequentlyChangingAnnotationsResourceList
}

func (list FrequentlyChangingAnnotationsResourceList) Each(f func(element *FrequentlyChangingAnnotationsResource)) {
	for _, frequentlyChangingAnnotationsResource := range list {
		f(frequentlyChangingAnnotationsResource)
	}
}

func (list FrequentlyChangingAnnotationsResourceList) EachResource(f func(element resources.Resource)) {
	for _, frequentlyChangingAnnotationsResource := range list {
		f(frequentlyChangingAnnotationsResource)
	}
}

func (list FrequentlyChangingAnnotationsResourceList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *FrequentlyChangingAnnotationsResource) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

// Kubernetes Adapter for FrequentlyChangingAnnotationsResource

func (o *FrequentlyChangingAnnotationsResource) GetObjectKind() schema.ObjectKind {
	t := FrequentlyChangingAnnotationsResourceCrd.TypeMeta()
	return &t
}

func (o *FrequentlyChangingAnnotationsResource) DeepCopyObject() runtime.Object {
	return resources.Clone(o).(*FrequentlyChangingAnnotationsResource)
}

func (o *FrequentlyChangingAnnotationsResource) DeepCopyInto(out *FrequentlyChangingAnnotationsResource) {
	clone := resources.Clone(o).(*FrequentlyChangingAnnotationsResource)
	*out = *clone
}

var (
	FrequentlyChangingAnnotationsResourceCrd = crd.NewCrd(
		"fcars",
		FrequentlyChangingAnnotationsResourceGVK.Group,
		FrequentlyChangingAnnotationsResourceGVK.Version,
		FrequentlyChangingAnnotationsResourceGVK.Kind,
		"fcar",
		false,
		&FrequentlyChangingAnnotationsResource{})
)

var (
	FrequentlyChangingAnnotationsResourceGVK = schema.GroupVersionKind{
		Version: "v2alpha1",
		Group:   "testing.solo.io",
		Kind:    "FrequentlyChangingAnnotationsResource",
	}
)
