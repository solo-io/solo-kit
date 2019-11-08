package multicluster

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

var (
	NoClientForClusterError = func(kind, cluster string) error {
		return errors.Errorf("%v client does not exist for %v", kind, cluster)
	}
)

type multiClusterResourceClient struct {
	resourceType resources.Resource
	clientGetter ClusterClientGetter
}

var _ clients.ResourceClient = &multiClusterResourceClient{}

func NewMultiClusterResourceClient(resourceType resources.Resource, clientSet ClusterClientGetter) *multiClusterResourceClient {
	return &multiClusterResourceClient{
		resourceType: resourceType,
		clientGetter: clientSet,
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

func (rc *multiClusterResourceClient) clientFor(cluster string) (clients.ResourceClient, error) {
	if client, ok := rc.clientGetter.ClientForCluster(cluster); ok {
		return client, nil
	}
	return nil, NoClientForClusterError(rc.Kind(), cluster)
}
