package multicluster

import (
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/test/mocks/v2alpha1"
	"k8s.io/client-go/rest"
)

type MockResourceMultiClusterClient interface {
	ClusterHandler
	v2alpha1.MockResourceClient
}

type mockResourceClientSet struct {
	clients      map[string]v2alpha1.MockResourceClient
	clientAccess sync.RWMutex
	aggregator   wrapper.WatchAggregator
	cacheGetter  KubeSharedCacheGetter
	opts         KubeResourceFactoryOpts
}

func NewMockResourceClientSet(cacheGetter KubeSharedCacheGetter, opts KubeResourceFactoryOpts) MockResourceMultiClusterClient {
	return NewMockResourceClientWithWatchAggregator(cacheGetter, nil, opts)
}

func NewMockResourceClientWithWatchAggregator(cacheGetter KubeSharedCacheGetter, aggregator wrapper.WatchAggregator, opts KubeResourceFactoryOpts) MockResourceMultiClusterClient {
	return &mockResourceClientSet{
		clients:      make(map[string]v2alpha1.MockResourceClient),
		clientAccess: sync.RWMutex{},
		cacheGetter:  cacheGetter,
		aggregator:   aggregator,
		opts:         opts,
	}
}

func (c *mockResourceClientSet) clientFor(cluster string) (v2alpha1.MockResourceClient, error) {
	c.clientAccess.RLock()
	defer c.clientAccess.RUnlock()
	if client, ok := c.clients[cluster]; ok {
		return client, nil
	}
	return nil, NoClientForClusterError(v2alpha1.MockResourceCrd.GroupVersionKind().String(), cluster)
}

func (c *mockResourceClientSet) ClusterAdded(cluster string, restConfig *rest.Config) {
	krc := &factory.KubeResourceClientFactory{
		Cluster:            cluster,
		Crd:                v2alpha1.MockResourceCrd,
		Cfg:                restConfig,
		SharedCache:        c.cacheGetter.GetCache(cluster),
		SkipCrdCreation:    c.opts.SkipCrdCreation,
		NamespaceWhitelist: c.opts.NamespaceWhitelist,
		ResyncPeriod:       c.opts.ResyncPeriod,
	}
	client, err := v2alpha1.NewMockResourceClient(krc)
	if err != nil {
		return
	}
	if err := client.Register(); err != nil {
		return
	}
	c.clientAccess.Lock()
	defer c.clientAccess.Unlock()
	c.clients[cluster] = client
	if c.aggregator != nil {
		c.aggregator.AddWatch(client.BaseClient())
	}
}

func (c *mockResourceClientSet) ClusterRemoved(cluster string, restConfig *rest.Config) {
	c.clientAccess.Lock()
	defer c.clientAccess.Unlock()
	if client, ok := c.clients[cluster]; ok {
		delete(c.clients, cluster)
		if c.aggregator != nil {
			c.aggregator.RemoveWatch(client.BaseClient())
		}
	}
}

// TODO should we split this off the client interface?
func (c *mockResourceClientSet) BaseClient() clients.ResourceClient {
	panic("not implemented")
}

// TODO should we split this off the client interface?
func (c *mockResourceClientSet) Register() error {
	panic("not implemented")
}

func (c *mockResourceClientSet) Read(namespace, name string, opts clients.ReadOpts) (*v2alpha1.MockResource, error) {
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterClient.Read(namespace, name, opts)
}

func (c *mockResourceClientSet) Write(resource *v2alpha1.MockResource, opts clients.WriteOpts) (*v2alpha1.MockResource, error) {
	clusterClient, err := c.clientFor(resource.GetMetadata().GetCluster())
	if err != nil {
		return nil, err
	}
	return clusterClient.Write(resource, opts)
}

func (c *mockResourceClientSet) Delete(namespace, name string, opts clients.DeleteOpts) error {
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return err
	}
	return clusterClient.Delete(namespace, name, opts)
}

func (c *mockResourceClientSet) List(namespace string, opts clients.ListOpts) (v2alpha1.MockResourceList, error) {
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterClient.List(namespace, opts)
}

func (c *mockResourceClientSet) Watch(namespace string, opts clients.WatchOpts) (<-chan v2alpha1.MockResourceList, <-chan error, error) {
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return nil, nil, err
	}
	return clusterClient.Watch(namespace, opts)
}
