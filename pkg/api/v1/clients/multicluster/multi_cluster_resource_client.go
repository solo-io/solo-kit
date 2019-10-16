package multicluster

import (
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
	"github.com/solo-io/solo-kit/pkg/multicluster/handler"
	"k8s.io/client-go/rest"
)

var (
	NoClientForClusterError = func(kind, cluster string) error {
		return errors.Errorf("%v client does not exist for %v", kind, cluster)
	}
)

type MultiClusterResourceClient interface {
	clients.ResourceClient
	handler.ClusterHandler
}

type multiClusterResourceClient struct {
	resourceType    resources.Resource
	clientGetter    ClientGetter
	watchAggregator wrapper.WatchAggregator
	clients         map[string]clients.ResourceClient
	clientAccess    sync.RWMutex
}

var _ MultiClusterResourceClient = &multiClusterResourceClient{}

func NewMultiClusterResourceClient(
	resourceType resources.Resource,
	clientGetter ClientGetter,
	watchAggregator wrapper.WatchAggregator,
) *multiClusterResourceClient {
	return &multiClusterResourceClient{
		resourceType:    resourceType,
		clientGetter:    clientGetter,
		watchAggregator: watchAggregator,
		clients:         make(map[string]clients.ResourceClient),
		clientAccess:    sync.RWMutex{},
	}
}

func (rc *multiClusterResourceClient) Kind() string {
	return resources.Kind(rc.resourceType)
}

func (rc *multiClusterResourceClient) NewResource() resources.Resource {
	return resources.Clone(rc.resourceType)
}

func (rc *multiClusterResourceClient) Register() error {
	// not implemented
	// per-cluster clients are registered on ClusterAdded
	return nil
}

func (rc *multiClusterResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	client, err := rc.clientFor(opts.Cluster)
	if err != nil {
		return nil, err
	}
	return client.Read(namespace, name, opts)
}

func (rc *multiClusterResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	client, err := rc.clientFor(resource.GetMetadata().Cluster)
	if err != nil {
		return nil, err
	}
	return client.Write(resource, opts)
}

func (rc *multiClusterResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	client, err := rc.clientFor(opts.Cluster)
	if err != nil {
		return err
	}
	return client.Delete(namespace, name, opts)
}

func (rc *multiClusterResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	client, err := rc.clientFor(opts.Cluster)
	if err != nil {
		return nil, err
	}

	return client.List(namespace, opts)
}

func (rc *multiClusterResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	client, err := rc.clientFor(opts.Cluster)
	if err != nil {
		return nil, nil, err
	}

	return client.Watch(namespace, opts)
}

func (rc *multiClusterResourceClient) ClusterAdded(cluster string, restConfig *rest.Config) {
	client, err := rc.clientGetter.GetClient(cluster, restConfig)
	if err != nil {
		return
	}
	if err := client.Register(); err != nil {
		return
	}
	rc.clientAccess.Lock()
	defer rc.clientAccess.Unlock()
	rc.clients[cluster] = client
	if rc.watchAggregator != nil {
		rc.watchAggregator.AddWatch(client)
	}
}

func (rc *multiClusterResourceClient) ClusterRemoved(cluster string, restConfig *rest.Config) {
	rc.clientAccess.Lock()
	defer rc.clientAccess.Unlock()
	if client, ok := rc.clients[cluster]; ok {
		delete(rc.clients, cluster)
		if rc.watchAggregator != nil {
			rc.watchAggregator.RemoveWatch(client)
		}
	}
}

func (rc *multiClusterResourceClient) clientFor(cluster string) (clients.ResourceClient, error) {
	rc.clientAccess.RLock()
	defer rc.clientAccess.RUnlock()
	if client, ok := rc.clients[cluster]; ok {
		return client, nil
	}
	return nil, NoClientForClusterError(rc.Kind(), cluster)
}
