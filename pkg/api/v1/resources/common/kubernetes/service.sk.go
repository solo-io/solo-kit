// Code generated by solo-kit. DO NOT EDIT.

package kubernetes

import (
	"hash"
	"sort"

	github_com_solo_io_solo_kit_api_external_kubernetes_service "github.com/solo-io/solo-kit/api/external/kubernetes/service"

	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewService(namespace, name string) *Service {
	service := &Service{}
	service.Service.SetMetadata(core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return service
}

// require custom resource to implement Clone() as well as resources.Resource interface

type CloneableService interface {
	resources.Resource
	Clone() *github_com_solo_io_solo_kit_api_external_kubernetes_service.Service
}

var _ CloneableService = &github_com_solo_io_solo_kit_api_external_kubernetes_service.Service{}

type Service struct {
	github_com_solo_io_solo_kit_api_external_kubernetes_service.Service
}

func (r *Service) Clone() resources.Resource {
	return &Service{Service: *r.Service.Clone()}
}

func (r *Service) Hash(hasher hash.Hash64) (uint64, error) {
	clone := r.Service.Clone()
	resources.UpdateMetadata(clone, func(meta *core.Metadata) {
		meta.ResourceVersion = ""
	})
	return hashutils.HashAll(clone), nil
}

func (r *Service) GroupVersionKind() schema.GroupVersionKind {
	return ServiceGVK
}

type ServiceList []*Service

// namespace is optional, if left empty, names can collide if the list contains more than one with the same name
func (list ServiceList) Find(namespace, name string) (*Service, error) {
	for _, service := range list {
		if service.GetMetadata().Name == name {
			if namespace == "" || service.GetMetadata().Namespace == namespace {
				return service, nil
			}
		}
	}
	return nil, errors.Errorf("list did not find service %v.%v", namespace, name)
}

func (list ServiceList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, service := range list {
		ress = append(ress, service)
	}
	return ress
}

func (list ServiceList) Names() []string {
	var names []string
	for _, service := range list {
		names = append(names, service.GetMetadata().Name)
	}
	return names
}

func (list ServiceList) NamespacesDotNames() []string {
	var names []string
	for _, service := range list {
		names = append(names, service.GetMetadata().Namespace+"."+service.GetMetadata().Name)
	}
	return names
}

func (list ServiceList) Sort() ServiceList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list ServiceList) Clone() ServiceList {
	var serviceList ServiceList
	for _, service := range list {
		serviceList = append(serviceList, resources.Clone(service).(*Service))
	}
	return serviceList
}

func (list ServiceList) Each(f func(element *Service)) {
	for _, service := range list {
		f(service)
	}
}

func (list ServiceList) EachResource(f func(element resources.Resource)) {
	for _, service := range list {
		f(service)
	}
}

func (list ServiceList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *Service) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

var (
	ServiceGVK = schema.GroupVersionKind{
		Version: "kubernetes",
		Group:   "kubernetes.solo.io",
		Kind:    "Service",
	}
)
