package customresourcedefinition

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
)

type crdResourceClientFactory struct {
	cacheGetter clustercache.CacheGetter
}

var _ factory.ClusterClientFactory = &crdResourceClientFactory{}

func NewCrdResourceClientFactory(cacheGetter clustercache.CacheGetter) *crdResourceClientFactory {
	return &crdResourceClientFactory{cacheGetter: cacheGetter}
}

func (g *crdResourceClientFactory) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	kube, err := apiexts.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	kubeCache := g.cacheGetter.GetCache(cluster, restConfig)
	typedCache, ok := kubeCache.(KubeCustomResourceDefinitionCache)
	if !ok {
		return nil, errors.Errorf("expected KubeCustomResourceDefinitionCache, got %T", kubeCache)
	}
	return wrapper.NewClusterResourceClient(newResourceClient(kube, typedCache), cluster), nil
}
