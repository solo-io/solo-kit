package crd

import (
	"fmt"
	"sync"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/go-utils/kubeutils"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Registry struct {
	crds []CombinedCrd
	mu   sync.RWMutex
}

var (
	registry *Registry

	VersionExistsError = func(version string) error {
		return errors.Errorf("tried adding version %s, but it already exists")
	}

	NotFoundError = func(gk schema.GroupKind) error {
		return errors.Errorf("could not find the crd for %v", gk)
	}
)

func init() {
	registry = &Registry{}
}

func GetRegistry() *Registry {
	return registry
}

func (r *Registry) AddCrd(resource Crd) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, crd := range r.crds {
		if crd.GroupKind() == resource.GroupKind() {
			for _, version := range crd.Versions {
				if version.Version == resource.Version.Version {
					return VersionExistsError(resource.Version.Version)
				}
			}
			crd.Versions = append(crd.Versions, resource.Version)
			return nil
		}
	}
	r.crds = append(r.crds, CombinedCrd{
		Versions: []Version{resource.Version},
		CrdMeta:  resource.CrdMeta,
	})
	return nil
}

func (r *Registry) GetCrd(gk schema.GroupKind) (CombinedCrd, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, crd := range r.crds {
		if crd.GroupKind() == gk {
			return crd, nil
		}
	}
	return CombinedCrd{}, NotFoundError(gk)
}

func (r *Registry) RegisterCrd(gvk schema.GroupVersionKind, clientset apiexts.Interface) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	crd, err := r.GetCrd(gvk.GroupKind())
	if err != nil {
		return err
	}
	scope := v1beta1.NamespaceScoped
	if crd.ClusterScoped {
		scope = v1beta1.ClusterScoped
	}
	versions := make([]v1beta1.CustomResourceDefinitionVersion, len(crd.Versions))
	for i, version := range crd.Versions {
		versionToAdd := v1beta1.CustomResourceDefinitionVersion{
			Name: version.Version,
		}
		if gvk.Version == version.Version {
			versionToAdd.Served = true
			versionToAdd.Storage = true
		}
		versions[i] = versionToAdd

	}
	toRegister := &v1beta1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: crd.FullName()},
		Spec: v1beta1.CustomResourceDefinitionSpec{
			Group: crd.Group,
			Scope: scope,
			Names: v1beta1.CustomResourceDefinitionNames{
				Plural:     crd.Plural,
				Kind:       crd.KindName,
				ShortNames: []string{crd.ShortName},
			},
			Versions: versions,
		},
	}
	_, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(toRegister)
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to register crd: %v", err)
	}
	return kubeutils.WaitForCrdActive(clientset, toRegister.Name)
}
