// Code generated by solo-kit. DO NOT EDIT.

package kubernetes

import (
	"encoding/binary"
	"hash"
	"hash/fnv"
	"log"
	"sort"

	github_com_solo_io_solo_kit_api_external_kubernetes_deployment "github.com/solo-io/solo-kit/api/external/kubernetes/deployment"

	"github.com/solo-io/go-utils/hashutils"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func NewDeployment(namespace, name string) *Deployment {
	deployment := &Deployment{}
	deployment.Deployment.SetMetadata(&core.Metadata{
		Name:      name,
		Namespace: namespace,
	})
	return deployment
}

// require custom resource to implement Clone() as well as resources.Resource interface

type CloneableDeployment interface {
	resources.Resource
	Clone() *github_com_solo_io_solo_kit_api_external_kubernetes_deployment.Deployment
}

var _ CloneableDeployment = &github_com_solo_io_solo_kit_api_external_kubernetes_deployment.Deployment{}

type Deployment struct {
	github_com_solo_io_solo_kit_api_external_kubernetes_deployment.Deployment
}

func (r *Deployment) Clone() resources.Resource {
	return &Deployment{Deployment: *r.Deployment.Clone()}
}

func (r *Deployment) Hash(hasher hash.Hash64) (uint64, error) {
	if hasher == nil {
		hasher = fnv.New64()
	}
	clone := r.Deployment.Clone()
	resources.UpdateMetadata(clone, func(meta *core.Metadata) {
		meta.ResourceVersion = ""
	})
	err := binary.Write(hasher, binary.LittleEndian, hashutils.HashAll(clone))
	if err != nil {
		return 0, err
	}
	return hasher.Sum64(), nil
}

func (r *Deployment) MustHash() uint64 {
	hashVal, err := r.Hash(nil)
	if err != nil {
		log.Panicf("error while hashing: (%s) this should never happen", err)
	}
	return hashVal
}

func (r *Deployment) GroupVersionKind() schema.GroupVersionKind {
	return DeploymentGVK
}

type DeploymentList []*Deployment

func (list DeploymentList) Find(namespace, name string) (*Deployment, error) {
	for _, deployment := range list {
		if deployment.GetMetadata().Name == name && deployment.GetMetadata().Namespace == namespace {
			return deployment, nil
		}
	}
	return nil, errors.Errorf("list did not find deployment %v.%v", namespace, name)
}

func (list DeploymentList) AsResources() resources.ResourceList {
	var ress resources.ResourceList
	for _, deployment := range list {
		ress = append(ress, deployment)
	}
	return ress
}

func (list DeploymentList) Names() []string {
	var names []string
	for _, deployment := range list {
		names = append(names, deployment.GetMetadata().Name)
	}
	return names
}

func (list DeploymentList) NamespacesDotNames() []string {
	var names []string
	for _, deployment := range list {
		names = append(names, deployment.GetMetadata().Namespace+"."+deployment.GetMetadata().Name)
	}
	return names
}

func (list DeploymentList) Sort() DeploymentList {
	sort.SliceStable(list, func(i, j int) bool {
		return list[i].GetMetadata().Less(list[j].GetMetadata())
	})
	return list
}

func (list DeploymentList) Clone() DeploymentList {
	var deploymentList DeploymentList
	for _, deployment := range list {
		deploymentList = append(deploymentList, resources.Clone(deployment).(*Deployment))
	}
	return deploymentList
}

func (list DeploymentList) Each(f func(element *Deployment)) {
	for _, deployment := range list {
		f(deployment)
	}
}

func (list DeploymentList) EachResource(f func(element resources.Resource)) {
	for _, deployment := range list {
		f(deployment)
	}
}

func (list DeploymentList) AsInterfaces() []interface{} {
	var asInterfaces []interface{}
	list.Each(func(element *Deployment) {
		asInterfaces = append(asInterfaces, element)
	})
	return asInterfaces
}

var (
	DeploymentGVK = schema.GroupVersionKind{
		Version: "kubernetes",
		Group:   "kubernetes.solo.io",
		Kind:    "Deployment",
	}
)
