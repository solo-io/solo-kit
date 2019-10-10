package bar

import (
	"sync"

	"github.com/solo-io/go-utils/errors"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	. "github.com/solo-io/solo-kit/pkg/api/v1/resources/common/kubernetes"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	"k8s.io/client-go/rest"
)

type PodFooBar interface {
	handler.ClusterHandler
	PodInterface
}

type podFooBar struct {
	clients       map[string]PodClient
	clientAccess  sync.RWMutex
	aggregator    wrapper.WatchAggregator
	factoryGetter factory.ResourceClientFactoryGetter
}

func NewPodFooBar(factoryGetter factory.ResourceClientFactoryGetter) PodFooBar {
	return NewPodFooBarWithWatchAggregator(nil, factoryGetter)
}

func NewPodFooBarWithWatchAggregator(aggregator wrapper.WatchAggregator, factoryGetter factory.ResourceClientFactoryGetter) PodFooBar {
	return &podFooBar{
		clients:       make(map[string]PodClient),
		clientAccess:  sync.RWMutex{},
		aggregator:    aggregator,
		factoryGetter: factoryGetter,
	}
}

func (c *podFooBar) interfaceFor(cluster string) (PodInterface, error) {
	c.clientAccess.RLock()
	defer c.clientAccess.RUnlock()
	if client, ok := c.clients[cluster]; ok {
		return client, nil
	}
	return nil, errors.Errorf("%v.%v client not found for cluster %v", "kubernetes", "Pod", cluster)
}

func (c *podFooBar) ClusterAdded(cluster string, restConfig *rest.Config) {
	client, err := NewPodClient(c.factoryGetter.ForCluster(cluster, restConfig))
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

func (c *podFooBar) ClusterRemoved(cluster string, restConfig *rest.Config) {
	c.clientAccess.Lock()
	defer c.clientAccess.Unlock()
	if client, ok := c.clients[cluster]; ok {
		delete(c.clients, cluster)
		if c.aggregator != nil {
			c.aggregator.RemoveWatch(client.BaseClient())
		}
	}
}

func (c *podFooBar) Read(namespace, name string, opts clients.ReadOpts) (*Pod, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.Read(namespace, name, opts)
}

func (c *podFooBar) Write(pod *Pod, opts clients.WriteOpts) (*Pod, error) {
	clusterInterface, err := c.interfaceFor(pod.GetMetadata().Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.Write(pod, opts)
}

func (c *podFooBar) Delete(namespace, name string, opts clients.DeleteOpts) error {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return err
	}
	return clusterInterface.Delete(namespace, name, opts)
}

func (c *podFooBar) List(namespace string, opts clients.ListOpts) (PodList, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return clusterInterface.List(namespace, opts)
}

func (c *podFooBar) Watch(namespace string, opts clients.WatchOpts) (<-chan PodList, <-chan error, error) {
	clusterInterface, err := c.interfaceFor(opts.Cluster)
	if err != nil {
		return nil, nil, err
	}
	return clusterInterface.Watch(namespace, opts)
}
