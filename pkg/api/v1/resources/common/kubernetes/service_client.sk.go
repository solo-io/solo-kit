// Code generated by solo-kit. DO NOT EDIT.

package kubernetes

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type ServiceWatcher interface {
	// watch namespace-scoped services
	Watch(namespace string, opts clients.WatchOpts) (<-chan ServiceList, <-chan error, error)
}

type ServiceClient interface {
	BaseClient() clients.ResourceClient
	Register() error
	Read(namespace, name string, opts clients.ReadOpts) (*Service, error)
	Write(resource *Service, opts clients.WriteOpts) (*Service, error)
	Delete(namespace, name string, opts clients.DeleteOpts) error
	List(namespace string, opts clients.ListOpts) (ServiceList, error)
	ServiceWatcher
}

type serviceClient struct {
	rc clients.ResourceClient
}

func NewServiceClient(ctx context.Context, rcFactory factory.ResourceClientFactory) (ServiceClient, error) {
	return NewServiceClientWithToken(ctx, rcFactory, "")
}

func NewServiceClientWithToken(ctx context.Context, rcFactory factory.ResourceClientFactory, token string) (ServiceClient, error) {
	rc, err := rcFactory.NewResourceClient(ctx, factory.NewResourceClientParams{
		ResourceType: &Service{},
		Token:        token,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating base Service resource client")
	}
	return NewServiceClientWithBase(rc), nil
}

func NewServiceClientWithBase(rc clients.ResourceClient) ServiceClient {
	return &serviceClient{
		rc: rc,
	}
}

func (client *serviceClient) BaseClient() clients.ResourceClient {
	return client.rc
}

func (client *serviceClient) Register() error {
	return client.rc.Register()
}

func (client *serviceClient) Read(namespace, name string, opts clients.ReadOpts) (*Service, error) {
	opts = opts.WithDefaults()

	resource, err := client.rc.Read(namespace, name, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*Service), nil
}

func (client *serviceClient) Write(service *Service, opts clients.WriteOpts) (*Service, error) {
	opts = opts.WithDefaults()
	resource, err := client.rc.Write(service, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*Service), nil
}

func (client *serviceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()

	return client.rc.Delete(namespace, name, opts)
}

func (client *serviceClient) List(namespace string, opts clients.ListOpts) (ServiceList, error) {
	opts = opts.WithDefaults()

	resourceList, err := client.rc.List(namespace, opts)
	if err != nil {
		return nil, err
	}
	return convertToService(resourceList), nil
}

func (client *serviceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan ServiceList, <-chan error, error) {
	opts = opts.WithDefaults()

	resourcesChan, errs, initErr := client.rc.Watch(namespace, opts)
	if initErr != nil {
		return nil, nil, initErr
	}
	servicesChan := make(chan ServiceList)
	go func() {
		for {
			select {
			case resourceList := <-resourcesChan:
				servicesChan <- convertToService(resourceList)
			case <-opts.Ctx.Done():
				close(servicesChan)
				return
			}
		}
	}()
	return servicesChan, errs, nil
}

func convertToService(resources resources.ResourceList) ServiceList {
	var serviceList ServiceList
	for _, resource := range resources {
		serviceList = append(serviceList, resource.(*Service))
	}
	return serviceList
}
