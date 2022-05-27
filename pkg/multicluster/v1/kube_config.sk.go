// Code generated by solo-kit. DO NOT EDIT.

//Source: pkg/code-generator/codegen/templates/resource_template.go
package v1

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"log"
	"sort"

	github_com_solo_io_solo_kit_api_multicluster_v1 "github.com/solo-io/solo-kit/api/multicluster/v1"

	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewKubeConfig(namespace, name string) *KubeConfig {
	kubeconfig := &KubeConfig{}
	kubeconfig.KubeConfig.SetMetadata(&core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return kubeconfig
}

// require custom resource to implement Clone() as well as resources.Resource interface

type CloneableKubeConfig interface {
	resources.Resource
	Clone() *github_com_solo_io_solo_kit_api_multicluster_v1.KubeConfig
}

var _ CloneableKubeConfig = &github_com_solo_io_solo_kit_api_multicluster_v1.KubeConfig{}

type KubeConfig struct {
	github_com_solo_io_solo_kit_api_multicluster_v1.KubeConfig
}

func (r *KubeConfig) Clone() resources.Resource {
	return &KubeConfig{KubeConfig: *r.KubeConfig.Clone()}
}

func (r *KubeConfig) Hash(hasher hash.Hash64) (uint64, error) {
	if hasher == nil {
		hasher = fnv.New64()
	}
	clone := r.KubeConfig.Clone()
	resources.UpdateMetadata(clone, func(meta *core.Metadata) {
		meta.ResourceVersion = ""
	})
	err := binary.Write(hasher, binary.LittleEndian, hashutils.HashAll(clone))
	if err != nil {
		return 0, err
	}
	return hasher.Sum64(), nil
}

func (r *KubeConfig) MustHash() uint64 {
	hashVal, err := r.Hash(nil)
	if err != nil {
		log.Panicf("error while hashing: (%s) this should never happen", err)
	}
	return hashVal
}

func (r *KubeConfig) GroupVersionKind() schema.GroupVersionKind {
	return KubeConfigGVK
}

type KubeConfigList []*KubeConfig

func (list KubeConfigList) Find(namespace, name string) (*KubeConfig, error) {
	for _, kubeConfig := range list {
		if kubeConfig.GetMetadata().Name == name && kubeConfig.GetMetadata().Namespace == namespace {
			return kubeConfig, nil
		}
	}
	return nil, errors.Errorf("list did not find kubeConfig %v.%v", namespace, name)
}

func (list KubeConfigList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, kubeConfig := range list {
		ress = append(ress, kubeConfig)
	}
	return ress
}

func (list KubeConfigList) Names() []string {
	var names []string
	for _, kubeConfig := range list {
		names = append(names, kubeConfig.GetMetadata().Name)
	}
	return names
}

func (list KubeConfigList) NamespacesDotNames() []string {
	var names []string
	for _, kubeConfig := range list {
		names = append(names, kubeConfig.GetMetadata().Namespace+"."+kubeConfig.GetMetadata().Name)
	}
	return names
}

func (list KubeConfigList) Sort() KubeConfigList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list KubeConfigList) Clone() KubeConfigList {
	var kubeConfigList KubeConfigList
	for _, kubeConfig := range list {
		kubeConfigList = append(kubeConfigList, resources.Clone(kubeConfig).(*KubeConfig))
	}
	return kubeConfigList
}

func (list KubeConfigList) Each(f func(element *KubeConfig)) {
	for _, kubeConfig := range list {
		f(kubeConfig)
	}
}

func (list KubeConfigList) EachResource(f func(element resources.Resource)) {
	for _, kubeConfig := range list {
		f(kubeConfig)
	}
}

func (list KubeConfigList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *KubeConfig) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

var (
	KubeConfigGVK = schema.GroupVersionKind{
		Version: "v1",
		Group:   "multicluster.solo.io",
		Kind:    "KubeConfig",
	}
)
