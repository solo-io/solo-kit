package kubernetes

import (
	"context"

	"github.com/solo-io/solo-kit/api/external/kubernetes/configmap"
	cmClient "github.com/solo-io/solo-kit/pkg/api/v1/clients/configmap"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	skkube "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	kubev1 "k8s.io/api/core/v1"

	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"k8s.io/client-go/kubernetes"
)

func NewConfigMapClient(kube kubernetes.Interface, cache cache.KubeCoreCache) skkube.ConfigMapClient {
	resourceClient, _ := cmClient.NewResourceClientWithConverter(
		kube,
		&skkube.ConfigMap{},
		cache,
		&kubeConverter{},
	)
	return skkube.NewConfigMapClientWithBase(resourceClient)
}
func FromKubeConfigMap(cm *kubev1.ConfigMap) *skkube.ConfigMap {

	configMapCopy := cm.DeepCopy()
	kubeConfigMap := configmap.ConfigMap(*configMapCopy)
	resource := &skkube.ConfigMap{
		ConfigMap: kubeConfigMap,
	}

	return resource
}

func ToKubeConfigMap(resource resources.Resource) (*kubev1.ConfigMap, error) {
	cmResource, ok := resource.(*skkube.ConfigMap)
	if !ok {
		return nil, errors.Errorf("internal error: invalid resource %v passed to config-map-only client", resources.Kind(resource))
	}

	cm := kubev1.ConfigMap(cmResource.ConfigMap)

	return &cm, nil
}

type kubeConverter struct{}

func (cc *kubeConverter) FromKubeConfigMap(ctx context.Context, rc *cmClient.ResourceClient, configMap *kubev1.ConfigMap) (resources.Resource, error) {
	return FromKubeConfigMap(configMap), nil
}

func (cc *kubeConverter) ToKubeConfigMap(ctx context.Context, rc *cmClient.ResourceClient, resource resources.Resource) (*kubev1.ConfigMap, error) {
	return ToKubeConfigMap(resource)
}

var _ cmClient.ConfigMapConverter = &kubeConverter{}
