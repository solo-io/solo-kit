package helpers

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"

	"github.com/rotisserie/eris"
	"github.com/solo-io/go-utils/versionutils/kubeapi"
	"github.com/solo-io/k8s-utils/kubeutils"
	"k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// The CRDRegistry is designed to be used only by tests
// We have moved CRD registration responsibilities outside of this repository.
// However, this is still useful as a testing utility
type crdRegistry struct {
	crds []crd.MultiVersionCrd
	mu   sync.RWMutex
}

var (
	registry *crdRegistry

	VersionExistsError = func(version string) error {
		return eris.Errorf("tried adding version %s, but it already exists", version)
	}

	NotFoundError = func(id string) error {
		return eris.Errorf("could not find the combined crd for %v", id)
	}

	InvalidGVKError = func(gvk schema.GroupVersionKind) error {
		return eris.Errorf("the following gvk %v does not correspond to a crd in the combined crd object", gvk)
	}
)

func init() {
	registry = &crdRegistry{}
}

func getRegistry() *crdRegistry {
	return registry
}

func AddCrd(resource crd.Crd) error {
	return getRegistry().addCrd(resource)
}

func RegisterCrd(ctx context.Context, crd crd.Crd, apiexts apiexts.Interface) error {
	return getRegistry().registerCrd(ctx, crd.GroupVersionKind(), apiexts)
}

func GetCrds() []crd.MultiVersionCrd {
	return getRegistry().crds
}

func GetCrd(gvk schema.GroupVersionKind) (crd.Crd, error) {
	return getRegistry().getCrd(gvk)
}

func GetMultiVersionCrd(gk schema.GroupKind) (crd.MultiVersionCrd, error) {
	return getRegistry().getMultiVersionCrd(gk)
}

func GetKubeCrd(crd crd.MultiVersionCrd, gvk schema.GroupVersionKind) (*v1beta1.CustomResourceDefinition, error) {
	return getRegistry().getKubeCrd(crd, gvk)
}

func (r *crdRegistry) addCrd(resource crd.Crd) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i, crd := range r.crds {
		if crd.GroupKind() == resource.GroupKind() {
			for _, version := range crd.Versions {
				if version.Version == resource.Version.Version {
					return VersionExistsError(resource.Version.Version)
				}
			}
			r.crds[i].Versions = append(crd.Versions, resource.Version)
			return nil
		}
	}
	r.crds = append(r.crds, crd.MultiVersionCrd{
		Versions: []crd.Version{resource.Version},
		CrdMeta:  resource.CrdMeta,
	})
	return nil
}

func (r *crdRegistry) getCrd(gvk schema.GroupVersionKind) (crd.Crd, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	combined, err := r.getMultiVersionCrd(gvk.GroupKind())
	if err != nil {
		return crd.Crd{}, err
	}
	for _, version := range combined.Versions {
		if version.Version == gvk.Version {
			return crd.Crd{
				CrdMeta: combined.CrdMeta,
				Version: version,
			}, nil
		}
	}
	return crd.Crd{}, NotFoundError(gvk.String())
}

func (r *crdRegistry) getMultiVersionCrd(gk schema.GroupKind) (crd.MultiVersionCrd, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, crd := range r.crds {
		if crd.GroupKind() == gk {
			return crd, nil
		}
	}
	return crd.MultiVersionCrd{}, NotFoundError(gk.String())
}

func (r *crdRegistry) registerCrd(ctx context.Context, gvk schema.GroupVersionKind, clientset apiexts.Interface) error {
	r.mu.RLock()
	defer r.mu.RUnlock()
	crd, err := r.getMultiVersionCrd(gvk.GroupKind())
	if err != nil {
		return err
	}
	toRegister, err := r.getKubeCrd(crd, gvk)
	if err != nil {
		return err
	}
	_, err = clientset.ApiextensionsV1beta1().CustomResourceDefinitions().Create(context.TODO(), toRegister, metav1.CreateOptions{})
	if err != nil && !apierrors.IsAlreadyExists(err) {
		return fmt.Errorf("failed to register crd: %v", err)
	}
	return kubeutils.WaitForCrdActive(ctx, clientset, toRegister.Name)
}

func (r *crdRegistry) getKubeCrd(crd crd.MultiVersionCrd, gvk schema.GroupVersionKind) (*v1beta1.CustomResourceDefinition, error) {
	scope := v1beta1.NamespaceScoped
	if crd.ClusterScoped {
		scope = v1beta1.ClusterScoped
	}
	versions := make([]v1beta1.CustomResourceDefinitionVersion, len(crd.Versions))
	validGvk := false
	for i, version := range crd.Versions {
		versionToAdd := v1beta1.CustomResourceDefinitionVersion{
			Name: version.Version,
		}
		if gvk.Version == version.Version {
			versionToAdd.Served = true
			versionToAdd.Storage = true
			validGvk = true
		}
		versions[i] = versionToAdd
	}
	if !validGvk {
		return nil, InvalidGVKError(gvk)
	}

	// Kubernetes expects Version to match the name of the first element specified in the Versions list.
	// Sort so the first version in the list will also be the latest.
	sort.Slice(versions, func(i, j int) bool {
		parsedi, err := kubeapi.ParseVersion(versions[i].Name)
		if err != nil {
			return false
		}
		parsedj, err := kubeapi.ParseVersion(versions[j].Name)
		if err != nil {
			return false
		}
		return parsedi.GreaterThan(parsedj)
	})

	return &v1beta1.CustomResourceDefinition{
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
			Version:  versions[0].Name,
		},
	}, nil
}
