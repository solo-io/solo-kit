package kubernetes

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/configmap"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

type configmapResourceClientGetter struct {
	cacheGetter  clustercache.KubeCoreCacheGetter
	resourceType resources.Resource
	converter    configmap.ConfigMapConverter
}

var _ multicluster.ClientGetter = &configmapResourceClientGetter{}

func NewConfigmapResourceClientGetter(cacheGetter clustercache.KubeCoreCacheGetter) *configmapResourceClientGetter {
	return &configmapResourceClientGetter{cacheGetter: cacheGetter}
}

func (g *configmapResourceClientGetter) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	kube, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return configmap.NewResourceClientWithConverter(kube, g.resourceType, g.cacheGetter.GetCache(cluster, restConfig), g.converter)
}
