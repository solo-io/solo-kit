package bar

import (
	"sync"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/multicluster"
	. "github.com/solo-io/solo-kit/pkg/multicluster/v1"
	"github.com/solo-io/solo-kit/test/mocks/v2alpha1"
	"k8s.io/client-go/rest"
)

type FooBar interface {
	multicluster.ClusterHandler
	KubeConfigInterface
}

type mgr struct {
	manager multicluster.KubeSharedCacheManager
}

func (m *mgr) getFactory() factory.ResourceFactoryForCluster {
	return func(cluster string, restConfig *rest.Config) factory.ResourceClientFactory {
		return &factory.KubeResourceClientFactory{
			Cluster:     cluster,
			Crd:         v2alpha1.MockResourceCrd,
			Cfg:         restConfig,
			SharedCache: m.manager.GetCache(cluster),
		}
	}
}

type barBaz struct {
	clients      map[string]KubeConfigClient
	clientAccess sync.RWMutex
	aggregator   wrapper.WatchAggregator
	factoryFor   factory.ResourceFactoryForCluster
}

func NewFooBar(getFactory factory.ResourceFactoryForCluster) FooBar {
	return NewFooBarWithWatchAggregator(nil, getFactory)
}

func NewFooBarWithWatchAggregator(aggregator wrapper.WatchAggregator, getFactory factory.ResourceFactoryForCluster) FooBar {
	return &barBaz{
		clients:      make(map[string]KubeConfigClient),
		clientAccess: sync.RWMutex{},
		aggregator:   aggregator,
		factoryFor:   getFactory,
	}
}

func (c *barBaz) interfaceFor(cluster string) (KubeConfigInterface, error) {
	c.clientAccess.RLock()
	defer c.clientAccess.RUnlock()
	if client, ok := c.clients[cluster]; ok {
		return client, nil
	}
	return nil, errors.Errorf("%v.%v client not found for cluster %v", "v1", "KubeConfig", cluster)
}

func (c *barBaz) ClusterAdded(cluster string, restConfig *rest.Config) {
	client, err := NewKubeConfigClient(c.factoryFor(cluster, restConfig))
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

func (c *barBaz) ClusterRemoved(cluster string, restConfig *rest.Config) {
	c.clientAccess.Lock()
	defer c.clientAccess.Unlock()
	if client, ok := c.clients[cluster]; ok {
		delete(c.clients, cluster)
		if c.aggregator != nil {
			c.aggregator.RemoveWatch(client.BaseClient())
		}
	}
}

func (c *barBaz) Read(namespace, name string, opts clients.ReadOpts) (*KubeConfig, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.Read(namespace, name, opts)
}

func (c *barBaz) Write(kubeConfig *KubeConfig, opts clients.WriteOpts) (*KubeConfig, error) {
	clusterInterface, err := c.interfaceFor(kubeConfig.GetMetadata().GetCluster())
	if err != nil {
		return nil, err
	}
	return clusterInterface.Write(kubeConfig, opts)
}

func (c *barBaz) Delete(namespace, name string, opts clients.DeleteOpts) error {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return err
	}
	return clusterInterface.Delete(namespace, name, opts)
}

func (c *barBaz) List(namespace string, opts clients.ListOpts) (KubeConfigList, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.List(namespace, opts)
}

func (c *barBaz) Watch(namespace string, opts clients.WatchOpts) (<-chan KubeConfigList, <-chan error, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, nil, err
	}
	return clusterInterface.Watch(namespace, opts)
}
