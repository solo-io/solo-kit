// Code generated by solo-kit. DO NOT EDIT.

package kubernetes

import (
	"hash"
	"sort"

	github_com_solo_io_solo_kit_api_external_kubernetes_customresourcedefinition "github.com/solo-io/solo-kit/api/external/kubernetes/customresourcedefinition"

	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewCustomResourceDefinition(namespace, name string) *CustomResourceDefinition {
	customresourcedefinition := &CustomResourceDefinition{}
	customresourcedefinition.CustomResourceDefinition.SetMetadata(core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return customresourcedefinition
}

// require custom resource to implement Clone() as well as resources.Resource interface

type CloneableCustomResourceDefinition interface {
	resources.Resource
	Clone() *github_com_solo_io_solo_kit_api_external_kubernetes_customresourcedefinition.CustomResourceDefinition
}

var _ CloneableCustomResourceDefinition = &github_com_solo_io_solo_kit_api_external_kubernetes_customresourcedefinition.CustomResourceDefinition{}

type CustomResourceDefinition struct {
	github_com_solo_io_solo_kit_api_external_kubernetes_customresourcedefinition.CustomResourceDefinition
}

func (r *CustomResourceDefinition) Clone() resources.Resource {
	return &CustomResourceDefinition{CustomResourceDefinition: *r.CustomResourceDefinition.Clone()}
}

func (r *CustomResourceDefinition) Hash(hasher hash.Hash64) (uint64, error) {
	clone := r.CustomResourceDefinition.Clone()
	resources.UpdateMetadata(clone, func(meta *core.Metadata) {
		meta.ResourceVersion = ""
	})
	return hashutils.HashAll(clone), nil
}

func (r *CustomResourceDefinition) GroupVersionKind() schema.GroupVersionKind {
	return CustomResourceDefinitionGVK
}

type CustomResourceDefinitionList []*CustomResourceDefinition

// namespace is optional, if left empty, names can collide if the list contains more than one with the same name
func (list CustomResourceDefinitionList) Find(namespace, name string) (*CustomResourceDefinition, error) {
	for _, customResourceDefinition := range list {
		if customResourceDefinition.GetMetadata().Name == name {
			if namespace == "" || customResourceDefinition.GetMetadata().Namespace == namespace {
				return customResourceDefinition, nil
			}
		}
	}
	return nil, errors.Errorf("list did not find customResourceDefinition %v.%v", namespace, name)
}

func (list CustomResourceDefinitionList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, customResourceDefinition := range list {
		ress = append(ress, customResourceDefinition)
	}
	return ress
}

func (list CustomResourceDefinitionList) Names() []string {
	var names []string
	for _, customResourceDefinition := range list {
		names = append(names, customResourceDefinition.GetMetadata().Name)
	}
	return names
}

func (list CustomResourceDefinitionList) NamespacesDotNames() []string {
	var names []string
	for _, customResourceDefinition := range list {
		names = append(names, customResourceDefinition.GetMetadata().Namespace+"."+customResourceDefinition.GetMetadata().Name)
	}
	return names
}

func (list CustomResourceDefinitionList) Sort() CustomResourceDefinitionList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list CustomResourceDefinitionList) Clone() CustomResourceDefinitionList {
	var customResourceDefinitionList CustomResourceDefinitionList
	for _, customResourceDefinition := range list {
		customResourceDefinitionList = append(customResourceDefinitionList, resources.Clone(customResourceDefinition).(*CustomResourceDefinition))
	}
	return customResourceDefinitionList
}

func (list CustomResourceDefinitionList) Each(f func(element *CustomResourceDefinition)) {
	for _, customResourceDefinition := range list {
		f(customResourceDefinition)
	}
}

func (list CustomResourceDefinitionList) EachResource(f func(element resources.Resource)) {
	for _, customResourceDefinition := range list {
		f(customResourceDefinition)
	}
}

func (list CustomResourceDefinitionList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *CustomResourceDefinition) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

var (
	CustomResourceDefinitionGVK = schema.GroupVersionKind{
		Version: "kubernetes",
		Group:   "kubernetes.solo.io",
		Kind:    "CustomResourceDefinition",
	}
)
