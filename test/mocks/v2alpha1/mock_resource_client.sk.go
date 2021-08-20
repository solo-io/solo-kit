// Code generated by solo-kit. DO NOT EDIT.

package v2alpha1

import (
	"context"
	"fmt"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type MockResourceWatcher interface {
	// watch namespace-scoped Mocks
	Watch(namespace string, opts clients.WatchOpts) (<-chan MockResourceList, <-chan error, error)
}

type MockResourceClient interface {
	BaseClient() clients.ResourceClient
	Register() error
	Read(namespace, name string, opts clients.ReadOpts) (*MockResource, error)
	Write(resource *MockResource, opts clients.WriteOpts) (*MockResource, error)
	Delete(namespace, name string, opts clients.DeleteOpts) error
	List(namespace string, opts clients.ListOpts) (MockResourceList, error)
	MockResourceWatcher
}

type mockResourceClient struct {
	rc clients.ResourceClient
}

func NewMockResourceClient(ctx context.Context, rcFactory factory.ResourceClientFactory) (MockResourceClient, error) {
	return NewMockResourceClientWithToken(ctx, rcFactory, "")
}

func NewMockResourceClientWithToken(ctx context.Context, rcFactory factory.ResourceClientFactory, token string) (MockResourceClient, error) {
	rc, err := rcFactory.NewResourceClient(ctx, factory.NewResourceClientParams{
		ResourceType: &MockResource{},
		Token:        token,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating base MockResource resource client")
	}
	return NewMockResourceClientWithBase(rc), nil
}

func NewMockResourceClientWithBase(rc clients.ResourceClient) MockResourceClient {
	return &mockResourceClient{
		rc: rc,
	}
}

func (client *mockResourceClient) BaseClient() clients.ResourceClient {
	return client.rc
}

func (client *mockResourceClient) Register() error {
	return client.rc.Register()
}

func (client *mockResourceClient) Read(namespace, name string, opts clients.ReadOpts) (*MockResource, error) {
	opts = opts.WithDefaults()

	resource, err := client.rc.Read(namespace, name, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*MockResource), nil
}

func (client *mockResourceClient) Write(mockResource *MockResource, opts clients.WriteOpts) (*MockResource, error) {
	fmt.Printf("v2alpha1 writing mock resource %+v\n", mockResource)
	opts = opts.WithDefaults()
	resource, err := client.rc.Write(mockResource, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*MockResource), nil
}

func (client *mockResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()

	return client.rc.Delete(namespace, name, opts)
}

func (client *mockResourceClient) List(namespace string, opts clients.ListOpts) (MockResourceList, error) {
	opts = opts.WithDefaults()

	resourceList, err := client.rc.List(namespace, opts)
	if err != nil {
		return nil, err
	}
	return convertToMockResource(resourceList), nil
}

func (client *mockResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan MockResourceList, <-chan error, error) {
	opts = opts.WithDefaults()

	resourcesChan, errs, initErr := client.rc.Watch(namespace, opts)
	if initErr != nil {
		return nil, nil, initErr
	}
	mocksChan := make(chan MockResourceList)
	go func() {
		for {
			select {
			case resourceList := <-resourcesChan:
				select {
				case mocksChan <- convertToMockResource(resourceList):
				case <-opts.Ctx.Done():
					close(mocksChan)
					return
				}
			case <-opts.Ctx.Done():
				close(mocksChan)
				return
			}
		}
	}()
	return mocksChan, errs, nil
}

func convertToMockResource(resources resources.ResourceList) MockResourceList {
	var mockResourceList MockResourceList
	for _, resource := range resources {
		mockResourceList = append(mockResourceList, resource.(*MockResource))
	}
	return mockResourceList
}
