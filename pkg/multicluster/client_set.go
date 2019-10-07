package multicluster

import (
	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/test/mocks/v2alpha1"
	"k8s.io/client-go/rest"
)

type MockResourceMultiClusterClient interface {
	ClusterHandler
	v2alpha1.MockResourceClient
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

var _ MockResourceMultiClusterClient = &mockResourceClientSet{}

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

func (c *mockResourceClientSet) BaseClient() clients.ResourceClient {
	// TODO this doesn't make sense here.
	panic("not implemented")
}

func (c *mockResourceClientSet) Register() error {
	// TODO this doesn't make sense here.
	return errors.Errorf("not implemented")
}

func (c *mockResourceClientSet) Read(namespace, name string, opts clients.ReadOpts) (*v2alpha1.MockResource, error) {
	clusterClient, err := c.ClientFor(opts.Cluster)
	if err != nil {
		return nil, errors.Errorf("Cluster %v is not accessible", opts.Cluster)
	}
	return clusterClient.Read(namespace, name, opts)
}

func (c *mockResourceClientSet) Write(resource *v2alpha1.MockResource, opts clients.WriteOpts) (*v2alpha1.MockResource, error) {
	clusterClient, err := c.ClientFor(resource.GetMetadata().GetCluster())
	if err != nil {
		return nil, errors.Errorf("Cluster %v is not accessible", resource.GetMetadata().GetCluster())
	}
	return clusterClient.Write(resource, opts)
}

func (c *mockResourceClientSet) Delete(namespace, name string, opts clients.DeleteOpts) error {
	clusterClient, err := c.ClientFor(opts.Cluster)
	if err != nil {
		return errors.Errorf("Cluster %v is not accessible", opts.Cluster)
	}
	return clusterClient.Delete(namespace, name, opts)
}

func (c *mockResourceClientSet) List(namespace string, opts clients.ListOpts) (v2alpha1.MockResourceList, error) {
	clusterClient, err := c.ClientFor(opts.Cluster)
	if err != nil {
		return nil, errors.Errorf("Cluster %v is not accessible", opts.Cluster)
	}
	return clusterClient.List(namespace, opts)
}

func (c *mockResourceClientSet) Watch(namespace string, opts clients.WatchOpts) (<-chan v2alpha1.MockResourceList, <-chan error, error) {
	clusterClient, err := c.ClientFor(opts.Cluster)
	if err != nil {
		return nil, nil, errors.Errorf("Cluster %v is not accessible", opts.Cluster)
	}
	return clusterClient.Watch(namespace, opts)
}
