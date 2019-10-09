// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type ClusterResourceWatcher interface {
	// watch cluster-scoped Clusterresources
	Watch(opts clients.WatchOpts) (<-chan ClusterResourceList, <-chan error, error)
}

type ClusterResourceInterface interface {
	Read(name string, opts clients.ReadOpts) (*ClusterResource, error)
	Write(resource *ClusterResource, opts clients.WriteOpts) (*ClusterResource, error)
	Delete(name string, opts clients.DeleteOpts) error
	List(opts clients.ListOpts) (ClusterResourceList, error)
	ClusterResourceWatcher
}

type ClusterResourceClient interface {
	BaseClient() clients.ResourceClient
	Register() error
	ClusterResourceInterface
}

type clusterResourceClient struct {
	rc clients.ResourceClient
}

func NewClusterResourceClient(rcFactory factory.ResourceClientFactory) (ClusterResourceClient, error) {
	return NewClusterResourceClientWithToken(rcFactory, "")
}

func NewClusterResourceClientWithToken(rcFactory factory.ResourceClientFactory, token string) (ClusterResourceClient, error) {
	rc, err := rcFactory.NewResourceClient(factory.NewResourceClientParams{
		ResourceType: &ClusterResource{},
		Token:        token,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating base ClusterResource resource client")
	}
	return NewClusterResourceClientWithBase(rc), nil
}

func NewClusterResourceClientWithBase(rc clients.ResourceClient) ClusterResourceClient {
	return &clusterResourceClient{
		rc: rc,
	}
}

func (client *clusterResourceClient) BaseClient() clients.ResourceClient {
	return client.rc
}

func (client *clusterResourceClient) Register() error {
	return client.rc.Register()
}

func (client *clusterResourceClient) Read(name string, opts clients.ReadOpts) (*ClusterResource, error) {
	opts = opts.WithDefaults()

	resource, err := client.rc.Read("", name, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*ClusterResource), nil
}

func (client *clusterResourceClient) Write(clusterResource *ClusterResource, opts clients.WriteOpts) (*ClusterResource, error) {
	opts = opts.WithDefaults()
	resource, err := client.rc.Write(clusterResource, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*ClusterResource), nil
}

func (client *clusterResourceClient) Delete(name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()

	return client.rc.Delete("", name, opts)
}

func (client *clusterResourceClient) List(opts clients.ListOpts) (ClusterResourceList, error) {
	opts = opts.WithDefaults()

	resourceList, err := client.rc.List("", opts)
	if err != nil {
		return nil, err
	}
	return convertToClusterResource(resourceList), nil
}

func (client *clusterResourceClient) Watch(opts clients.WatchOpts) (<-chan ClusterResourceList, <-chan error, error) {
	opts = opts.WithDefaults()

	resourcesChan, errs, initErr := client.rc.Watch("", opts)
	if initErr != nil {
		return nil, nil, initErr
	}
	clusterresourcesChan := make(chan ClusterResourceList)
	go func() {
		for {
			select {
			case resourceList := <-resourcesChan:
				clusterresourcesChan <- convertToClusterResource(resourceList)
			case <-opts.Ctx.Done():
				close(clusterresourcesChan)
				return
			}
		}
	}()
	return clusterresourcesChan, errs, nil
}

func convertToClusterResource(resources resources.ResourceList) ClusterResourceList {
	var clusterResourceList ClusterResourceList
	for _, resource := range resources {
		clusterResourceList = append(clusterResourceList, resource.(*ClusterResource))
	}
	return clusterResourceList
}
