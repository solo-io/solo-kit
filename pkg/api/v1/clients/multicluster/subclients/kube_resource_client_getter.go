package subclients

import (
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"k8s.io/client-go/rest"
)

type kubeResourceClientGetter struct {
	cacheGetter        clustercache.KubeSharedCacheGetter
	crd                crd.Crd
	skipCrdCreation    bool
	namespaceWhitelist []string
	resyncPeriod       time.Duration
	params             factory.NewResourceClientParams
}

var _ multicluster.ClientGetter = &kubeResourceClientGetter{}

func NewKubeResourceClientGetter(
	cacheGetter clustercache.KubeSharedCacheGetter,
	crd crd.Crd,
	skipCrdCreation bool,
	namespaceWhitelist []string,
	resyncPeriod time.Duration,
	params factory.NewResourceClientParams) *kubeResourceClientGetter {

	return &kubeResourceClientGetter{
		cacheGetter:        cacheGetter,
		crd:                crd,
		skipCrdCreation:    skipCrdCreation,
		namespaceWhitelist: namespaceWhitelist,
		resyncPeriod:       resyncPeriod,
		params:             params,
	}
}

func (g *kubeResourceClientGetter) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	f := &factory.KubeResourceClientFactory{
		Crd:                g.crd,
		Cfg:                restConfig,
		SharedCache:        g.cacheGetter.GetCache(cluster),
		SkipCrdCreation:    g.skipCrdCreation,
		NamespaceWhitelist: g.namespaceWhitelist,
		ResyncPeriod:       g.resyncPeriod,
		Cluster:            cluster,
	}
	return f.NewResourceClient(g.params)
}
