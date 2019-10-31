package service

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type serviceResourceClientFactory struct {
	cacheGetter clustercache.CacheGetter
}

var _ factory.ClusterClientFactory = &serviceResourceClientFactory{}

func NewServiceResourceClientFactory(cacheGetter clustercache.CacheGetter) *serviceResourceClientFactory {
	return &serviceResourceClientFactory{cacheGetter: cacheGetter}
}

func (g *serviceResourceClientFactory) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	kube, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	kubeCache := g.cacheGetter.GetCache(cluster, restConfig)
	typedCache, ok := kubeCache.(cache.KubeCoreCache)
	if !ok {
		return nil, errors.Errorf("expected KubeCoreCache, got %T", kubeCache)
	}
	return wrapper.NewClusterResourceClient(newResourceClient(kube, typedCache), cluster), nil
}
