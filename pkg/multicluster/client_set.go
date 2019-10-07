package multicluster

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/test/mocks/v2alpha1"
	"k8s.io/client-go/rest"
)

type MockResourceMultiClusterClient interface {
	ClusterHandler
	// TODO generate strongly typed client here
	ClientFor(cluster string) (v2alpha1.MockResourceClient, error)
}

type mockResourceClientSet struct {
	clients    map[string]v2alpha1.MockResourceClient
	aggregator wrapper.WatchAggregator
	// TODO this should be different depending on which kind of client we have
	// Maybe it's more of a client factory
	cacheGetter CacheGetter
}

func NewMockResourceClientSet(cacheGetter CacheGetter) *mockResourceClientSet {
	return NewMockResourceClientWithWatchAggregator(cacheGetter, nil)
}

func NewMockResourceClientWithWatchAggregator(cacheGetter CacheGetter, aggregator wrapper.WatchAggregator) *mockResourceClientSet {
	return &mockResourceClientSet{
		clients:     make(map[string]v2alpha1.MockResourceClient),
		cacheGetter: cacheGetter,
		aggregator:  aggregator,
	}
}

func (c *mockResourceClientSet) ClientFor(cluster string) (v2alpha1.MockResourceClient, error) {
	if client, ok := c.clients[cluster]; ok {
		return client, nil
	}
	return nil, errors.Errorf("DNE")
}

func (c *mockResourceClientSet) ClusterAdded(cluster string, restConfig *rest.Config) {
	// TODO generate, support other types of clients
	cache := c.cacheGetter.GetCache(cluster)

	krc := &factory.KubeResourceClientFactory{
		Cluster:     cluster,
		Crd:         v2alpha1.MockResourceCrd,
		Cfg:         restConfig,
		SharedCache: cache,
		// TODO Pass in through opts to constructor
		SkipCrdCreation:    false,
		NamespaceWhitelist: nil,
		ResyncPeriod:       0,
	}
	client, err := v2alpha1.NewMockResourceClient(krc)
	if err != nil {
		return
	}
	if err := client.Register(); err != nil {
		return
	}
	// TODO handle cases where clients for the cluster already exist ?
	c.clients[cluster] = client
	if c.aggregator != nil {
		if err := c.aggregator.AddWatch(client.BaseClient()); err != nil {
			// TODO
		}
	}
}

func (c *mockResourceClientSet) ClusterRemoved(cluster string, restConfig *rest.Config) {
	if client, ok := c.clients[cluster]; ok {
		delete(c.clients, cluster)

		if c.aggregator != nil {
			c.aggregator.RemoveWatch(client.BaseClient())
		}
	}
}
