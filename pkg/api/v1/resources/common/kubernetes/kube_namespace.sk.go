// Code generated by solo-kit. DO NOT EDIT.

package kubernetes

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"log"
	"sort"

	github_com_solo_io_solo_kit_api_external_kubernetes_namespace "github.com/solo-io/solo-kit/api/external/kubernetes/namespace"

	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewKubeNamespace(namespace, name string) *KubeNamespace {
	kubenamespace := &KubeNamespace{}
	kubenamespace.KubeNamespace.SetMetadata(&core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return kubenamespace
}

// require custom resource to implement Clone() as well as resources.Resource interface

type CloneableKubeNamespace interface {
	resources.Resource
	Clone() *github_com_solo_io_solo_kit_api_external_kubernetes_namespace.KubeNamespace
}

var _ CloneableKubeNamespace = &github_com_solo_io_solo_kit_api_external_kubernetes_namespace.KubeNamespace{}

type KubeNamespace struct {
	github_com_solo_io_solo_kit_api_external_kubernetes_namespace.KubeNamespace
}

func (r *KubeNamespace) Clone() resources.Resource {
	return &KubeNamespace{KubeNamespace: *r.KubeNamespace.Clone()}
}

func (r *KubeNamespace) Hash(hasher hash.Hash64) (uint64, error) {
	if hasher == nil {
		hasher = fnv.New64()
	}
	clone := r.KubeNamespace.Clone()
	resources.UpdateMetadata(clone, func(meta *core.Metadata) {
		meta.ResourceVersion = ""
	})
	err := binary.Write(hasher, binary.LittleEndian, hashutils.HashAll(clone))
	if err != nil {
		return 0, err
	}
	return hasher.Sum64(), nil
}

func (r *KubeNamespace) MustHash() uint64 {
	hashVal, err := r.Hash(nil)
	if err != nil {
		log.Panicf("error while hashing: (%s) this should never happen", err)
	}
	return hashVal
}

func (r *KubeNamespace) GroupVersionKind() schema.GroupVersionKind {
	return KubeNamespaceGVK
}

type KubeNamespaceList []*KubeNamespace

func (list KubeNamespaceList) Find(namespace, name string) (*KubeNamespace, error) {
	for _, kubeNamespace := range list {
		if kubeNamespace.GetMetadata().Name == name && kubeNamespace.GetMetadata().Namespace == namespace {
			return kubeNamespace, nil
		}
	}
	return nil, errors.Errorf("list did not find kubeNamespace %v.%v", namespace, name)
}

func (list KubeNamespaceList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, kubeNamespace := range list {
		ress = append(ress, kubeNamespace)
	}
	return ress
}

func (list KubeNamespaceList) Names() []string {
	var names []string
	for _, kubeNamespace := range list {
		names = append(names, kubeNamespace.GetMetadata().Name)
	}
	return names
}

func (list KubeNamespaceList) NamespacesDotNames() []string {
	var names []string
	for _, kubeNamespace := range list {
		names = append(names, kubeNamespace.GetMetadata().Namespace+"."+kubeNamespace.GetMetadata().Name)
	}
	return names
}

func (list KubeNamespaceList) Sort() KubeNamespaceList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list KubeNamespaceList) Clone() KubeNamespaceList {
	var kubeNamespaceList KubeNamespaceList
	for _, kubeNamespace := range list {
		kubeNamespaceList = append(kubeNamespaceList, resources.Clone(kubeNamespace).(*KubeNamespace))
	}
	return kubeNamespaceList
}

func (list KubeNamespaceList) Each(f func(element *KubeNamespace)) {
	for _, kubeNamespace := range list {
		f(kubeNamespace)
	}
}

func (list KubeNamespaceList) EachResource(f func(element resources.Resource)) {
	for _, kubeNamespace := range list {
		f(kubeNamespace)
	}
}

func (list KubeNamespaceList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *KubeNamespace) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

var (
	KubeNamespaceGVK = schema.GroupVersionKind{
		Version: "kubernetes",
		Group:   "kubernetes.solo.io",
		Kind:    "KubeNamespace",
	}
)
