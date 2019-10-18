package customresourcedefinition

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
)

type crdResourceClientGetter struct {
	cacheGetter clustercache.CacheGetter
}

var _ multicluster.ClientGetter = &crdResourceClientGetter{}

func NewCrdResourceClientGetter(cacheGetter clustercache.CacheGetter) *crdResourceClientGetter {
	return &crdResourceClientGetter{cacheGetter: cacheGetter}
}

func (g *crdResourceClientGetter) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	kube, err := apiexts.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	kubeCache := g.cacheGetter.GetCache(cluster, restConfig)
	typedCache, ok := kubeCache.(KubeCustomResourceDefinitionCache)
	if !ok {
		return nil, errors.Errorf("expected KubeCustomResourceDefinitionCache, got %T", kubeCache)
	}
	return newResourceClient(kube, typedCache), nil
}
