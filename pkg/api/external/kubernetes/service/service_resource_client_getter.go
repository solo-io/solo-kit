package service

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type serviceResourceClientGetter struct {
	coreCacheGetter clustercache.KubeCoreCacheGetter
}

var _ multicluster.ClientGetter = &serviceResourceClientGetter{}

func NewServiceResourceClientGetter(coreCacheGetter clustercache.KubeCoreCacheGetter) *serviceResourceClientGetter {
	return &serviceResourceClientGetter{coreCacheGetter: coreCacheGetter}
}

func (g *serviceResourceClientGetter) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	kube, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return newResourceClient(kube, g.coreCacheGetter.GetCache(cluster, restConfig)), nil
}
