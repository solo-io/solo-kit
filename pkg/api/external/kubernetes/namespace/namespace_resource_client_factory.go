package namespace

import (
	"context"

	"github.com/rotisserie/eris"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type namespaceResourceClientFactory struct {
	cacheGetter clustercache.CacheGetter
}

var _ factory.ClusterClientFactory = &namespaceResourceClientFactory{}

func NewNamespaceResourceClientFactory(cacheGetter clustercache.CacheGetter) *namespaceResourceClientFactory {
	return &namespaceResourceClientFactory{cacheGetter: cacheGetter}
}

func (g *namespaceResourceClientFactory) GetClient(ctx context.Context, cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	kube, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	kubeCache := g.cacheGetter.GetCache(cluster, restConfig)
	typedCache, ok := kubeCache.(cache.KubeCoreCache)
	if !ok {
		return nil, eris.Errorf("expected KubeCoreCache, got %T", kubeCache)
	}
	return wrapper.NewClusterResourceClient(newResourceClient(kube, typedCache), cluster), nil
}
