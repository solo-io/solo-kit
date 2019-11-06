package job

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

type jobResourceClientFactory struct {
	cacheGetter clustercache.CacheGetter
}

var _ factory.ClusterClientFactory = &jobResourceClientFactory{}

func NewJobResourceClientFactory(cacheGetter clustercache.CacheGetter) *jobResourceClientFactory {
	return &jobResourceClientFactory{cacheGetter: cacheGetter}
}

func (g *jobResourceClientFactory) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	kube, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	kubeCache := g.cacheGetter.GetCache(cluster, restConfig)
	typedCache, ok := kubeCache.(cache.KubeJobCache)
	if !ok {
		return nil, errors.Errorf("expected KubeJobCache, got %T", kubeCache)
	}
	return wrapper.NewClusterResourceClient(newResourceClient(kube, typedCache), cluster), nil
}
