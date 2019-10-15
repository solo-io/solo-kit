package namespace

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type namespaceResourceClientGetter struct {
	coreCacheGetter clustercache.KubeCoreCacheGetter
}

var _ multicluster.ClientGetter = &namespaceResourceClientGetter{}

func NewNamespaceResourceClientGetter(coreCacheGetter clustercache.KubeCoreCacheGetter) *namespaceResourceClientGetter {
	return &namespaceResourceClientGetter{coreCacheGetter: coreCacheGetter}
}

func (p *namespaceResourceClientGetter) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	kube, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return newResourceClient(kube, p.coreCacheGetter.GetCache(cluster, restConfig)), nil
}
