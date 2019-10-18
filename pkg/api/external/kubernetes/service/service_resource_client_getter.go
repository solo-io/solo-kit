package service

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type serviceResourceClientGetter struct {
	cacheGetter clustercache.CacheGetter
}

var _ multicluster.ClientGetter = &serviceResourceClientGetter{}

func NewServiceResourceClientGetter(cacheGetter clustercache.CacheGetter) *serviceResourceClientGetter {
	return &serviceResourceClientGetter{cacheGetter: cacheGetter}
}

func (g *serviceResourceClientGetter) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	kube, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	kubeCache := g.cacheGetter.GetCache(cluster, restConfig)
	typedCache, ok := kubeCache.(cache.KubeCoreCache)
	if !ok {
		return nil, errors.Errorf("expected KubeCoreCache, got %T", kubeCache)
	}
	return newResourceClient(kube, typedCache), nil
}
