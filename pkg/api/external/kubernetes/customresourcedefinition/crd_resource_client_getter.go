package customresourcedefinition

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	apiexts "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	"k8s.io/client-go/rest"
)

type crdResourceClientGetter struct {
	cacheGetter CrdCacheGetter
}

var _ multicluster.ClientGetter = &crdResourceClientGetter{}

func NewCrdResourceClientGetter(cacheGetter CrdCacheGetter) *crdResourceClientGetter {
	return &crdResourceClientGetter{cacheGetter: cacheGetter}
}

func (g *crdResourceClientGetter) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	kube, err := apiexts.NewForConfig(restConfig)
	if err != nil {
		return nil, err
	}
	return newResourceClient(kube, g.cacheGetter.GetCache(cluster, restConfig)), nil
}
