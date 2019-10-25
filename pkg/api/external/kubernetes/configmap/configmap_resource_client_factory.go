package kubernetes

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/configmap"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type configmapResourceClientFactory struct {
	cacheGetter  clustercache.CacheGetter
	resourceType resources.Resource
	converter    configmap.ConfigMapConverter
}

var _ factory.ClusterClientFactory = &configmapResourceClientFactory{}

func NewConfigmapResourceClientFactory(cacheGetter clustercache.CacheGetter, resourceType resources.Resource) *configmapResourceClientFactory {
	return &configmapResourceClientFactory{
		cacheGetter:  cacheGetter,
		resourceType: resourceType,
	}
}

func (g *configmapResourceClientFactory) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	kube, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	kubeCache := g.cacheGetter.GetCache(cluster, restConfig)
	typedCache, ok := kubeCache.(cache.KubeCoreCache)
	if !ok {
		return nil, errors.Errorf("expected KubeCoreCache, got %T", kubeCache)
	}
	return configmap.NewResourceClient(kube, g.resourceType, typedCache, true)
}
