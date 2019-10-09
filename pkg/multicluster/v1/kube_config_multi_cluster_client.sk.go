package v1

import (
	"sync"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/multicluster"
	"k8s.io/client-go/rest"
)

type KubeConfigMultiClusterClient interface {
	multicluster.ClusterHandler
	KubeConfigInterface
}

type kubeConfigMultiClusterClient struct {
	clients       map[string]KubeConfigClient
	clientAccess  sync.RWMutex
	aggregator    wrapper.WatchAggregator
	factoryGetter factory.ResourceClientFactoryGetter
}

func NewKubeConfigMultiClusterClient(getFactory factory.ResourceFactoryForCluster) KubeConfigMultiClusterClient {
	return NewKubeConfigClientWithWatchAggregator(nil, getFactory)
}

func NewKubeConfigMultiClusterClientWithWatchAggregator(aggregator wrapper.WatchAggregator, factoryGetter factory.ResourceClientFactoryGetter) KubeConfigMultiClusterClient {
	return &kubeConfigMultiClusterClient{
		clients:       make(map[string]KubeConfigClient),
		clientAccess:  sync.RWMutex{},
		aggregator:    aggregator,
		factoryGetter: factoryGetter,
	}
}

func (c *kubeConfigMultiClusterClient) interfaceFor(cluster string) (KubeConfigInterface, error) {
	c.clientAccess.RLock()
	defer c.clientAccess.RUnlock()
	if client, ok := c.clients[cluster]; ok {
		return client, nil
	}
	return nil, errors.Errorf("%v.%v client not found for cluster %v", "v1", "KubeConfig", cluster)
}

func (c *kubeConfigMultiClusterClient) ClusterAdded(cluster string, restConfig *rest.Config) {
	client, err := NewKubeConfigClient(c.factoryGetter.ForCluster(cluster, restConfig))
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

func (c *kubeConfigMultiClusterClient) ClusterRemoved(cluster string, restConfig *rest.Config) {
	c.clientAccess.Lock()
	defer c.clientAccess.Unlock()
	if client, ok := c.clients[cluster]; ok {
		delete(c.clients, cluster)
		if c.aggregator != nil {
			c.aggregator.RemoveWatch(client.BaseClient())
		}
	}
}

func (c *kubeConfigMultiClusterClient) Read(namespace, name string, opts clients.ReadOpts) (*KubeConfig, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.Read(namespace, name, opts)
}

func (c *kubeConfigMultiClusterClient) Write(kubeConfig *KubeConfig, opts clients.WriteOpts) (*KubeConfig, error) {
	clusterInterface, err := c.interfaceFor(kubeConfig.GetMetadata().GetCluster())
	if err != nil {
		return nil, err
	}
	return clusterInterface.Write(kubeConfig, opts)
}

func (c *kubeConfigMultiClusterClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return err
	}
	return clusterInterface.Delete(namespace, name, opts)
}

func (c *kubeConfigMultiClusterClient) List(namespace string, opts clients.ListOpts) (KubeConfigList, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.List(namespace, opts)
}

func (c *kubeConfigMultiClusterClient) Watch(namespace string, opts clients.WatchOpts) (<-chan KubeConfigList, <-chan error, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, nil, err
	}
	return clusterInterface.Watch(namespace, opts)
}
