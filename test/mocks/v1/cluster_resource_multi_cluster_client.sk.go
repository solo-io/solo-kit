// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/multicluster"
	"k8s.io/client-go/rest"
)

type ClusterResourceMultiClusterClient interface {
	multicluster.ClusterHandler
	ClusterResourceInterface
}

type clusterResourceMultiClusterClient struct {
	clients      map[string]ClusterResourceClient
	clientAccess sync.RWMutex
	aggregator   wrapper.WatchAggregator
	cacheGetter  multicluster.KubeSharedCacheGetter
	opts         multicluster.KubeResourceFactoryOpts
}

func NewClusterResourceMultiClusterClient(cacheGetter multicluster.KubeSharedCacheGetter, opts multicluster.KubeResourceFactoryOpts) ClusterResourceMultiClusterClient {
	return NewClusterResourceClientWithWatchAggregator(cacheGetter, nil, opts)
}

func NewClusterResourceMultiClusterClientWithWatchAggregator(cacheGetter multicluster.KubeSharedCacheGetter, aggregator wrapper.WatchAggregator, opts multicluster.KubeResourceFactoryOpts) ClusterResourceMultiClusterClient {
	return &clusterResourceMultiClusterClient{
		clients:      make(map[string]ClusterResourceInterface),
		clientAccess: sync.RWMutex{},
		cacheGetter:  cacheGetter,
		aggregator:   aggregator,
		opts:         opts,
	}
}

func (c *clusterResourceMultiClusterClient) clientFor(cluster string) (ClusterResourceInterface, error) {
	c.clientAccess.RLock()
	defer c.clientAccess.RUnlock()
	if client, ok := c.clients[cluster]; ok {
		return client, nil
	}
	return nil, multicluster.NoClientForClusterError(ClusterResourceCrd.GroupVersionKind().String(), cluster)
}

func (c *clusterResourceMultiClusterClient) ClusterAdded(cluster string, restConfig *rest.Config) {
	krc := &factory.KubeResourceClientFactory{
		Cluster:            cluster,
		Crd:                ClusterResourceCrd,
		Cfg:                restConfig,
		SharedCache:        c.cacheGetter.GetCache(cluster),
		SkipCrdCreation:    c.opts.SkipCrdCreation,
		NamespaceWhitelist: c.opts.NamespaceWhitelist,
		ResyncPeriod:       c.opts.ResyncPeriod,
	}
	client, err := NewClusterResourceClient(krc)
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

func (c *clusterResourceMultiClusterClient) ClusterRemoved(cluster string, restConfig *rest.Config) {
	c.clientAccess.Lock()
	defer c.clientAccess.Unlock()
	if client, ok := c.clients[cluster]; ok {
		delete(c.clients, cluster)
		if c.aggregator != nil {
			c.aggregator.RemoveWatch(client.BaseClient())
		}
	}
}

func (c *clusterResourceMultiClusterClient) Read(name string, opts clients.ReadOpts) (*ClusterResource, error) {
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterClient.Read(namespace, name, opts)
}

func (c *clusterResourceMultiClusterClient) Write(clusterResource *ClusterResource, opts clients.WriteOpts) (*ClusterResource, error) {
	clusterClient, err := c.clientFor(clusterResource.GetMetadata().GetCluster())
	if err != nil {
		return nil, err
	}
	return clusterClient.Write(clusterResource, opts)
}

func (c *clusterResourceMultiClusterClient) Delete(name string, opts clients.DeleteOpts) error {
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return err
	}
	return clusterClient.Delete(namespace, name, opts)
}

func (c *clusterResourceMultiClusterClient) List(opts clients.ListOpts) (ClusterResourceList, error) {
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterClient.List(namespace, opts)
}

func (c *clusterResourceMultiClusterClient) Watch(opts clients.WatchOpts) (<-chan ClusterResourceList, <-chan error, error) {
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return nil, nil, err
	}
	return clusterClient.Watch(namespace, opts)
}
