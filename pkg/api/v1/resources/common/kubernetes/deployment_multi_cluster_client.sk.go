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
	NoDeploymentClientForClusterError = func(cluster string) error {
		return errors.Errorf("kubernetes.Deployment client not found for cluster %v", cluster)
	}
)

type DeploymentMultiClusterClient interface {
	handler.ClusterHandler
	DeploymentInterface
}

type deploymentMultiClusterClient struct {
	clients       map[string]DeploymentClient
	clientAccess  sync.RWMutex
	aggregator    wrapper.WatchAggregator
	factoryGetter factory.ResourceClientFactoryGetter
}

var _ DeploymentMultiClusterClient = &deploymentMultiClusterClient{}

func NewDeploymentMultiClusterClient(factoryGetter factory.ResourceClientFactoryGetter) *deploymentMultiClusterClient {
	return NewDeploymentMultiClusterClientWithWatchAggregator(nil, factoryGetter)
}

func NewDeploymentMultiClusterClientWithWatchAggregator(aggregator wrapper.WatchAggregator, factoryGetter factory.ResourceClientFactoryGetter) *deploymentMultiClusterClient {
	return &deploymentMultiClusterClient{
		clients:       make(map[string]DeploymentClient),
		clientAccess:  sync.RWMutex{},
		aggregator:    aggregator,
		factoryGetter: factoryGetter,
	}
}

func (c *deploymentMultiClusterClient) interfaceFor(cluster string) (DeploymentInterface, error) {
	c.clientAccess.RLock()
	defer c.clientAccess.RUnlock()
	if client, ok := c.clients[cluster]; ok {
		return client, nil
	}
	return nil, NoDeploymentClientForClusterError(cluster)
}

func (c *deploymentMultiClusterClient) ClusterAdded(cluster string, restConfig *rest.Config) {
	client, err := NewDeploymentClient(c.factoryGetter.ForCluster(cluster, restConfig))
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

func (c *deploymentMultiClusterClient) ClusterRemoved(cluster string, restConfig *rest.Config) {
	c.clientAccess.Lock()
	defer c.clientAccess.Unlock()
	if client, ok := c.clients[cluster]; ok {
		delete(c.clients, cluster)
		if c.aggregator != nil {
			c.aggregator.RemoveWatch(client.BaseClient())
		}
	}
}

func (c *deploymentMultiClusterClient) Read(namespace, name string, opts clients.ReadOpts) (*Deployment, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}

	return clusterInterface.Read(namespace, name, opts)
}

func (c *deploymentMultiClusterClient) Write(deployment *Deployment, opts clients.WriteOpts) (*Deployment, error) {
	clusterInterface, err := c.interfaceFor(deployment.GetMetadata().Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.Write(deployment, opts)
}

func (c *deploymentMultiClusterClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return err
	}

	return clusterInterface.Delete(namespace, name, opts)
}

func (c *deploymentMultiClusterClient) List(namespace string, opts clients.ListOpts) (DeploymentList, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}

	return clusterInterface.List(namespace, opts)
}

func (c *deploymentMultiClusterClient) Watch(namespace string, opts clients.WatchOpts) (<-chan DeploymentList, <-chan error, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, nil, err
	}

	return clusterInterface.Watch(namespace, opts)
}
