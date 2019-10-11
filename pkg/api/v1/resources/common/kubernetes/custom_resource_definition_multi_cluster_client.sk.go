// Code generated by solo-kit. DO NOT EDIT.

package kubernetes

import (
	"sync"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	"k8s.io/client-go/rest"
)

var (
	NoCustomResourceDefinitionClientForClusterError = func(cluster string) error {
		return errors.Errorf("kubernetes.CustomResourceDefinition client not found for cluster %v", cluster)
	}
)

type CustomResourceDefinitionMultiClusterClient interface {
	handler.ClusterHandler
	CustomResourceDefinitionInterface
}

type customResourceDefinitionMultiClusterClient struct {
	clients       map[string]CustomResourceDefinitionClient
	clientAccess  sync.RWMutex
	aggregator    wrapper.WatchAggregator
	factoryGetter factory.ResourceClientFactoryGetter
}

var _ CustomResourceDefinitionMultiClusterClient = &customResourceDefinitionMultiClusterClient{}

func NewCustomResourceDefinitionMultiClusterClient(factoryGetter factory.ResourceClientFactoryGetter) *customResourceDefinitionMultiClusterClient {
	return NewCustomResourceDefinitionMultiClusterClientWithWatchAggregator(nil, factoryGetter)
}

func NewCustomResourceDefinitionMultiClusterClientWithWatchAggregator(aggregator wrapper.WatchAggregator, factoryGetter factory.ResourceClientFactoryGetter) *customResourceDefinitionMultiClusterClient {
	return &customResourceDefinitionMultiClusterClient{
		clients:       make(map[string]CustomResourceDefinitionClient),
		clientAccess:  sync.RWMutex{},
		aggregator:    aggregator,
		factoryGetter: factoryGetter,
	}
}

func (c *customResourceDefinitionMultiClusterClient) interfaceFor(cluster string) (CustomResourceDefinitionInterface, error) {
	c.clientAccess.RLock()
	defer c.clientAccess.RUnlock()
	if client, ok := c.clients[cluster]; ok {
		return client, nil
	}
	return nil, NoCustomResourceDefinitionClientForClusterError(cluster)
}

func (c *customResourceDefinitionMultiClusterClient) ClusterAdded(cluster string, restConfig *rest.Config) {
	client, err := NewCustomResourceDefinitionClient(c.factoryGetter.ForCluster(cluster, restConfig))
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

func (c *customResourceDefinitionMultiClusterClient) ClusterRemoved(cluster string, restConfig *rest.Config) {
	c.clientAccess.Lock()
	defer c.clientAccess.Unlock()
	if client, ok := c.clients[cluster]; ok {
		delete(c.clients, cluster)
		if c.aggregator != nil {
			c.aggregator.RemoveWatch(client.BaseClient())
		}
	}
}

func (c *customResourceDefinitionMultiClusterClient) Read(namespace, name string, opts clients.ReadOpts) (*CustomResourceDefinition, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}

	return clusterInterface.Read(namespace, name, opts)
}

func (c *customResourceDefinitionMultiClusterClient) Write(customResourceDefinition *CustomResourceDefinition, opts clients.WriteOpts) (*CustomResourceDefinition, error) {
	clusterInterface, err := c.interfaceFor(customResourceDefinition.GetMetadata().Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.Write(customResourceDefinition, opts)
}

func (c *customResourceDefinitionMultiClusterClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return err
	}

	return clusterInterface.Delete(namespace, name, opts)
}

func (c *customResourceDefinitionMultiClusterClient) List(namespace string, opts clients.ListOpts) (CustomResourceDefinitionList, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}

	return clusterInterface.List(namespace, opts)
}

func (c *customResourceDefinitionMultiClusterClient) Watch(namespace string, opts clients.WatchOpts) (<-chan CustomResourceDefinitionList, <-chan error, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, nil, err
	}

	return clusterInterface.Watch(namespace, opts)
}
