package ClientFactory

import (
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/multicluster"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/multicluster/clustercache"
	"k8s.io/client-go/rest"
)

type kubeResourceClientFactory struct {
	cacheGetter        clustercache.CacheGetter
	crd                crd.Crd
	skipCrdCreation    bool
	namespaceWhitelist []string
	resyncPeriod       time.Duration
	params             factory.NewResourceClientParams
}

var _ multicluster.ClusterClientFactory = &kubeResourceClientFactory{}

func NewKubeResourceClientFactory(
	cacheGetter clustercache.CacheGetter,
	crd crd.Crd,
	skipCrdCreation bool,
	namespaceWhitelist []string,
	resyncPeriod time.Duration,
	params factory.NewResourceClientParams) *kubeResourceClientFactory {

	return &kubeResourceClientFactory{
		cacheGetter:        cacheGetter,
		crd:                crd,
		skipCrdCreation:    skipCrdCreation,
		namespaceWhitelist: namespaceWhitelist,
		resyncPeriod:       resyncPeriod,
		params:             params,
	}
}

func (g *kubeResourceClientFactory) GetClient(cluster string, restConfig *rest.Config) (clients.ResourceClient, error) {
	kubeCache := g.cacheGetter.GetCache(cluster, restConfig)
	typedCache, ok := kubeCache.(kube.SharedCache)
	if !ok {
		return nil, errors.Errorf("expected KubeSharedCache, got %T", kubeCache)
	}

	f := &factory.KubeResourceClientFactory{
		Crd:                g.crd,
		Cfg:                restConfig,
		SharedCache:        typedCache,
		SkipCrdCreation:    g.skipCrdCreation,
		NamespaceWhitelist: g.namespaceWhitelist,
		ResyncPeriod:       g.resyncPeriod,
		Cluster:            cluster,
	}
	return f.NewResourceClient(g.params)
}
