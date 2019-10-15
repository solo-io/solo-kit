package subclients

import (
	"github.com/solo-io/solo-kit/pkg/api/external/kubernetes/pod"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type podResourceClientGetter struct {
	coreCacheGetter clustercache.KubeCoreCacheGetter
}

var _ multicluster.ClientGetter = &podResourceClientGetter{}

func NewPodResourceClientGetter(coreCacheGetter clustercache.KubeCoreCacheGetter) *podResourceClientGetter {
	return &podResourceClientGetter{coreCacheGetter: coreCacheGetter}
}

func (p *podResourceClientGetter) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	kube, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return pod.NewResourceClient(kube, p.coreCacheGetter.GetCache(cluster, restConfig)), nil
}
