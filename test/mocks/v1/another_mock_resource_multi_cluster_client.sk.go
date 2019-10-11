// Code generated by solo-kit. DO NOT EDIT.

package v1

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
	NoAnotherMockResourceClientForClusterError = func(cluster string) error {
		return errors.Errorf("v1.AnotherMockResource client not found for cluster %v", cluster)
	}
)

type AnotherMockResourceMultiClusterClient interface {
	handler.ClusterHandler
	AnotherMockResourceInterface
}

type anotherMockResourceMultiClusterClient struct {
	clients       map[string]AnotherMockResourceClient
	clientAccess  sync.RWMutex
	aggregator    wrapper.WatchAggregator
	factoryGetter factory.ResourceClientFactoryGetter
}

var _ AnotherMockResourceMultiClusterClient = &anotherMockResourceMultiClusterClient{}

func NewAnotherMockResourceMultiClusterClient(factoryGetter factory.ResourceClientFactoryGetter) *anotherMockResourceMultiClusterClient {
	return NewAnotherMockResourceMultiClusterClientWithWatchAggregator(nil, factoryGetter)
}

func NewAnotherMockResourceMultiClusterClientWithWatchAggregator(aggregator wrapper.WatchAggregator, factoryGetter factory.ResourceClientFactoryGetter) *anotherMockResourceMultiClusterClient {
	return &anotherMockResourceMultiClusterClient{
		clients:       make(map[string]AnotherMockResourceClient),
		clientAccess:  sync.RWMutex{},
		aggregator:    aggregator,
		factoryGetter: factoryGetter,
	}
}

func (c *anotherMockResourceMultiClusterClient) interfaceFor(cluster string) (AnotherMockResourceInterface, error) {
	c.clientAccess.RLock()
	defer c.clientAccess.RUnlock()
	if client, ok := c.clients[cluster]; ok {
		return client, nil
	}
	return nil, NoAnotherMockResourceClientForClusterError(cluster)
}

func (c *anotherMockResourceMultiClusterClient) ClusterAdded(cluster string, restConfig *rest.Config) {
	client, err := NewAnotherMockResourceClient(c.factoryGetter.ForCluster(cluster, restConfig))
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

func (c *anotherMockResourceMultiClusterClient) ClusterRemoved(cluster string, restConfig *rest.Config) {
	c.clientAccess.Lock()
	defer c.clientAccess.Unlock()
	if client, ok := c.clients[cluster]; ok {
		delete(c.clients, cluster)
		if c.aggregator != nil {
			c.aggregator.RemoveWatch(client.BaseClient())
		}
	}
}

func (c *anotherMockResourceMultiClusterClient) Read(namespace, name string, opts clients.ReadOpts) (*AnotherMockResource, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}

	return clusterInterface.Read(namespace, name, opts)
}

func (c *anotherMockResourceMultiClusterClient) Write(anotherMockResource *AnotherMockResource, opts clients.WriteOpts) (*AnotherMockResource, error) {
	clusterInterface, err := c.interfaceFor(anotherMockResource.GetMetadata().Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.Write(anotherMockResource, opts)
}

func (c *anotherMockResourceMultiClusterClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return err
	}

	return clusterInterface.Delete(namespace, name, opts)
}

func (c *anotherMockResourceMultiClusterClient) List(namespace string, opts clients.ListOpts) (AnotherMockResourceList, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}

	return clusterInterface.List(namespace, opts)
}

func (c *anotherMockResourceMultiClusterClient) Watch(namespace string, opts clients.WatchOpts) (<-chan AnotherMockResourceList, <-chan error, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, nil, err
	}

	return clusterInterface.Watch(namespace, opts)
}
