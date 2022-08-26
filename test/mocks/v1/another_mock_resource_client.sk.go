// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type AnotherMockResourceWatcher interface {
	// watch namespace-scoped Anothermockresources
	Watch(namespace string, opts clients.WatchOpts) (<-chan AnotherMockResourceList, <-chan error, error)
}

type AnotherMockResourceClient interface {
	BaseClient() clients.ResourceClient
	Register() error
	Read(namespace, name string, opts clients.ReadOpts) (*AnotherMockResource, error)
	Write(resource *AnotherMockResource, opts clients.WriteOpts) (*AnotherMockResource, error)
	Delete(namespace, name string, opts clients.DeleteOpts) error
	List(namespace string, opts clients.ListOpts) (AnotherMockResourceList, error)
	AnotherMockResourceWatcher
}

type anotherMockResourceClient struct {
	rc clients.ResourceClient
}

func NewAnotherMockResourceClient(ctx context.Context, rcFactory factory.ResourceClientFactory) (AnotherMockResourceClient, error) {
	return NewAnotherMockResourceClientWithToken(ctx, rcFactory, "")
}

func NewAnotherMockResourceClientWithToken(ctx context.Context, rcFactory factory.ResourceClientFactory, token string) (AnotherMockResourceClient, error) {
	rc, err := rcFactory.NewResourceClient(ctx, factory.NewResourceClientParams{
		ResourceType: &AnotherMockResource{},
		Token:        token,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating base AnotherMockResource resource client")
	}
	return NewAnotherMockResourceClientWithBase(rc), nil
}

func NewAnotherMockResourceClientWithBase(rc clients.ResourceClient) AnotherMockResourceClient {
	return &anotherMockResourceClient{
		rc: rc,
	}
}

func (client *anotherMockResourceClient) BaseClient() clients.ResourceClient {
	return client.rc
}

func (client *anotherMockResourceClient) Register() error {
	return client.rc.Register()
}

func (client *anotherMockResourceClient) Read(namespace, name string, opts clients.ReadOpts) (*AnotherMockResource, error) {
	opts = opts.WithDefaults()

	resource, err := client.rc.Read(namespace, name, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*AnotherMockResource), nil
}

func (client *anotherMockResourceClient) Write(anotherMockResource *AnotherMockResource, opts clients.WriteOpts) (*AnotherMockResource, error) {
	opts = opts.WithDefaults()
	resource, err := client.rc.Write(anotherMockResource, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*AnotherMockResource), nil
}

func (client *anotherMockResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()

	return client.rc.Delete(namespace, name, opts)
}

func (client *anotherMockResourceClient) List(namespace string, opts clients.ListOpts) (AnotherMockResourceList, error) {
	opts = opts.WithDefaults()

	resourceList, err := client.rc.List(namespace, opts)
	if err != nil {
		return nil, err
	}
	return convertToAnotherMockResource(resourceList), nil
}

func (client *anotherMockResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan AnotherMockResourceList, <-chan error, error) {
	opts = opts.WithDefaults()

	resourcesChan, errs, initErr := client.rc.Watch(namespace, opts)
	if initErr != nil {
		return nil, nil, initErr
	}
	anothermockresourcesChan := make(chan AnotherMockResourceList)
	go func() {
		for {
			select {
			case resourceList := <-resourcesChan:
				select {
				case anothermockresourcesChan <- convertToAnotherMockResource(resourceList):
				case <-opts.Ctx.Done():
					close(anothermockresourcesChan)
					return
				}
			case <-opts.Ctx.Done():
				close(anothermockresourcesChan)
				return
			}
		}
	}()
	return anothermockresourcesChan, errs, nil
}

func convertToAnotherMockResource(resources resources.ResourceList) AnotherMockResourceList {
	var anotherMockResourceList AnotherMockResourceList
	for _, resource := range resources {
		anotherMockResourceList = append(anotherMockResourceList, resource.(*AnotherMockResource))
	}
	return anotherMockResourceList
}
