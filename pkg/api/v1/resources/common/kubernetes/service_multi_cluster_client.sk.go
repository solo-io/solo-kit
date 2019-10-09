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

type ServiceMultiClusterClient interface {
	multicluster.ClusterHandler
	ServiceInterface
}

type serviceMultiClusterClient struct {
	clients       map[string]ServiceClient
	clientAccess  sync.RWMutex
	aggregator    wrapper.WatchAggregator
	factoryGetter factory.ResourceClientFactoryGetter
}

func NewServiceMultiClusterClient(getFactory factory.ResourceFactoryForCluster) ServiceMultiClusterClient {
	return NewServiceClientWithWatchAggregator(nil, getFactory)
}

func NewServiceMultiClusterClientWithWatchAggregator(aggregator wrapper.WatchAggregator, factoryGetter factory.ResourceClientFactoryGetter) ServiceMultiClusterClient {
	return &serviceMultiClusterClient{
		clients:       make(map[string]ServiceClient),
		clientAccess:  sync.RWMutex{},
		aggregator:    aggregator,
		factoryGetter: factoryGetter,
	}
}

func (c *serviceMultiClusterClient) interfaceFor(cluster string) (ServiceInterface, error) {
	c.clientAccess.RLock()
	defer c.clientAccess.RUnlock()
	if client, ok := c.clients[cluster]; ok {
		return client, nil
	}
	return nil, errors.Errorf("%v.%v client not found for cluster %v", "kubernetes", "Service", cluster)
}

func (c *serviceMultiClusterClient) ClusterAdded(cluster string, restConfig *rest.Config) {
	client, err := NewServiceClient(c.factoryGetter.ForCluster(cluster, restConfig))
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

func (c *serviceMultiClusterClient) ClusterRemoved(cluster string, restConfig *rest.Config) {
	c.clientAccess.Lock()
	defer c.clientAccess.Unlock()
	if client, ok := c.clients[cluster]; ok {
		delete(c.clients, cluster)
		if c.aggregator != nil {
			c.aggregator.RemoveWatch(client.BaseClient())
		}
	}
}

func (c *serviceMultiClusterClient) Read(namespace, name string, opts clients.ReadOpts) (*Service, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.Read(namespace, name, opts)
}

func (c *serviceMultiClusterClient) Write(service *Service, opts clients.WriteOpts) (*Service, error) {
	clusterInterface, err := c.interfaceFor(service.GetMetadata().GetCluster())
	if err != nil {
		return nil, err
	}
	return clusterInterface.Write(service, opts)
}

func (c *serviceMultiClusterClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return err
	}
	return clusterInterface.Delete(namespace, name, opts)
}

func (c *serviceMultiClusterClient) List(namespace string, opts clients.ListOpts) (ServiceList, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.List(namespace, opts)
}

func (c *serviceMultiClusterClient) Watch(namespace string, opts clients.WatchOpts) (<-chan ServiceList, <-chan error, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, nil, err
	}
	return clusterInterface.Watch(namespace, opts)
}
