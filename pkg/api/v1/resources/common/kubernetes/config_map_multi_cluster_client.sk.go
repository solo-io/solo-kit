// Code generated by solo-kit. DO NOT EDIT.

package kubernetes

import (
	"sync"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/multicluster"
	"k8s.io/client-go/rest"
)

type ConfigMapMultiClusterClient interface {
	multicluster.ClusterHandler
	ConfigMapInterface
}

type configMapMultiClusterClient struct {
	clients       map[string]ConfigMapClient
	clientAccess  sync.RWMutex
	aggregator    wrapper.WatchAggregator
	factoryGetter factory.ResourceClientFactoryGetter
}

func NewConfigMapMultiClusterClient(getFactory factory.ResourceFactoryForCluster) ConfigMapMultiClusterClient {
	return NewConfigMapClientWithWatchAggregator(nil, getFactory)
}

func NewConfigMapMultiClusterClientWithWatchAggregator(aggregator wrapper.WatchAggregator, factoryGetter factory.ResourceClientFactoryGetter) ConfigMapMultiClusterClient {
	return &configMapMultiClusterClient{
		clients:       make(map[string]ConfigMapClient),
		clientAccess:  sync.RWMutex{},
		aggregator:    aggregator,
		factoryGetter: factoryGetter,
	}
}

func (c *configMapMultiClusterClient) interfaceFor(cluster string) (ConfigMapInterface, error) {
	c.clientAccess.RLock()
	defer c.clientAccess.RUnlock()
	if client, ok := c.clients[cluster]; ok {
		return client, nil
	}
	return nil, errors.Errorf("%v.%v client not found for cluster %v", "kubernetes", "ConfigMap", cluster)
}

func (c *configMapMultiClusterClient) ClusterAdded(cluster string, restConfig *rest.Config) {
	client, err := NewConfigMapClient(c.factoryGetter.ForCluster(cluster, restConfig))
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

func (c *configMapMultiClusterClient) ClusterRemoved(cluster string, restConfig *rest.Config) {
	c.clientAccess.Lock()
	defer c.clientAccess.Unlock()
	if client, ok := c.clients[cluster]; ok {
		delete(c.clients, cluster)
		if c.aggregator != nil {
			c.aggregator.RemoveWatch(client.BaseClient())
		}
	}
}

func (c *configMapMultiClusterClient) Read(namespace, name string, opts clients.ReadOpts) (*ConfigMap, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.Read(namespace, name, opts)
}

func (c *configMapMultiClusterClient) Write(configMap *ConfigMap, opts clients.WriteOpts) (*ConfigMap, error) {
	clusterInterface, err := c.interfaceFor(configMap.GetMetadata().GetCluster())
	if err != nil {
		return nil, err
	}
	return clusterInterface.Write(configMap, opts)
}

func (c *configMapMultiClusterClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return err
	}
	return clusterInterface.Delete(namespace, name, opts)
}

func (c *configMapMultiClusterClient) List(namespace string, opts clients.ListOpts) (ConfigMapList, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.List(namespace, opts)
}

func (c *configMapMultiClusterClient) Watch(namespace string, opts clients.WatchOpts) (<-chan ConfigMapList, <-chan error, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, nil, err
	}
	return clusterInterface.Watch(namespace, opts)
}
