// Code generated by solo-kit. DO NOT EDIT.

package kubernetes

import (
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/multicluster"
	"k8s.io/client-go/rest"
)

type KubeNamespaceMultiClusterClient interface {
	multicluster.ClusterHandler
	KubeNamespaceInterface
}

type kubeNamespaceMultiClusterClient struct {
	clients      map[string]KubeNamespaceClient
	clientAccess sync.RWMutex
	aggregator   wrapper.WatchAggregator
	cacheGetter  multicluster.KubeSharedCacheGetter
	opts         multicluster.KubeResourceFactoryOpts
}

func NewKubeNamespaceMultiClusterClient(cacheGetter multicluster.KubeSharedCacheGetter, opts multicluster.KubeResourceFactoryOpts) KubeNamespaceMultiClusterClient {
	return NewKubeNamespaceClientWithWatchAggregator(cacheGetter, nil, opts)
}

func NewKubeNamespaceMultiClusterClientWithWatchAggregator(cacheGetter multicluster.KubeSharedCacheGetter, aggregator wrapper.WatchAggregator, opts multicluster.KubeResourceFactoryOpts) KubeNamespaceMultiClusterClient {
	return &kubeNamespaceMultiClusterClient{
		clients:      make(map[string]KubeNamespaceInterface),
		clientAccess: sync.RWMutex{},
		cacheGetter:  cacheGetter,
		aggregator:   aggregator,
		opts:         opts,
	}
}

func (c *kubeNamespaceMultiClusterClient) clientFor(cluster string) (KubeNamespaceInterface, error) {
	c.clientAccess.RLock()
	defer c.clientAccess.RUnlock()
	if client, ok := c.clients[cluster]; ok {
		return client, nil
	}
	return nil, multicluster.NoClientForClusterError(KubeNamespaceCrd.GroupVersionKind().String(), cluster)
}

func (c *kubeNamespaceMultiClusterClient) ClusterAdded(cluster string, restConfig *rest.Config) {
	krc := &factory.KubeResourceClientFactory{
		Cluster:            cluster,
		Crd:                KubeNamespaceCrd,
		Cfg:                restConfig,
		SharedCache:        c.cacheGetter.GetCache(cluster),
		SkipCrdCreation:    c.opts.SkipCrdCreation,
		NamespaceWhitelist: c.opts.NamespaceWhitelist,
		ResyncPeriod:       c.opts.ResyncPeriod,
	}
	client, err := NewKubeNamespaceClient(krc)
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

func (c *kubeNamespaceMultiClusterClient) ClusterRemoved(cluster string, restConfig *rest.Config) {
	c.clientAccess.Lock()
	defer c.clientAccess.Unlock()
	if client, ok := c.clients[cluster]; ok {
		delete(c.clients, cluster)
		if c.aggregator != nil {
			c.aggregator.RemoveWatch(client.BaseClient())
		}
	}
}

func (c *kubeNamespaceMultiClusterClient) Read(namespace, name string, opts clients.ReadOpts) (*KubeNamespace, error) {
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterClient.Read(namespace, name, opts)
}

func (c *kubeNamespaceMultiClusterClient) Write(kubeNamespace *KubeNamespace, opts clients.WriteOpts) (*KubeNamespace, error) {
	clusterClient, err := c.clientFor(kubeNamespace.GetMetadata().GetCluster())
	if err != nil {
		return nil, err
	}
	return clusterClient.Write(kubeNamespace, opts)
}

func (c *kubeNamespaceMultiClusterClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return err
	}
	return clusterClient.Delete(namespace, name, opts)
}

func (c *kubeNamespaceMultiClusterClient) List(namespace string, opts clients.ListOpts) (KubeNamespaceList, error) {
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterClient.List(namespace, opts)
}

func (c *kubeNamespaceMultiClusterClient) Watch(namespace string, opts clients.WatchOpts) (<-chan KubeNamespaceList, <-chan error, error) {
	clusterClient, err := c.clientFor(opts.Cluster)
	if err != nil {
		return nil, nil, err
	}
	return clusterClient.Watch(namespace, opts)
}
