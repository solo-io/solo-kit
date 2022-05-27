// Code generated by solo-kit. DO NOT EDIT.

//Generated by pkg/code-generator/codegen/templates/resource_client_template.go
package v1

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type MockCustomTypeWatcher interface {
	// watch namespace-scoped mcts
	Watch(namespace string, opts clients.WatchOpts) (<-chan MockCustomTypeList, <-chan error, error)
}

type MockCustomTypeClient interface {
	BaseClient() clients.ResourceClient
	Register() error
	Read(namespace, name string, opts clients.ReadOpts) (*MockCustomType, error)
	Write(resource *MockCustomType, opts clients.WriteOpts) (*MockCustomType, error)
	Delete(namespace, name string, opts clients.DeleteOpts) error
	List(namespace string, opts clients.ListOpts) (MockCustomTypeList, error)
	MockCustomTypeWatcher
}

type mockCustomTypeClient struct {
	rc clients.ResourceClient
}

func NewMockCustomTypeClient(ctx context.Context, rcFactory factory.ResourceClientFactory) (MockCustomTypeClient, error) {
	return NewMockCustomTypeClientWithToken(ctx, rcFactory, "")
}

func NewMockCustomTypeClientWithToken(ctx context.Context, rcFactory factory.ResourceClientFactory, token string) (MockCustomTypeClient, error) {
	rc, err := rcFactory.NewResourceClient(ctx, factory.NewResourceClientParams{
		ResourceType: &MockCustomType{},
		Token:        token,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating base MockCustomType resource client")
	}
	return NewMockCustomTypeClientWithBase(rc), nil
}

func NewMockCustomTypeClientWithBase(rc clients.ResourceClient) MockCustomTypeClient {
	return &mockCustomTypeClient{
		rc: rc,
	}
}

func (client *mockCustomTypeClient) BaseClient() clients.ResourceClient {
	return client.rc
}

func (client *mockCustomTypeClient) Register() error {
	return client.rc.Register()
}

func (client *mockCustomTypeClient) Read(namespace, name string, opts clients.ReadOpts) (*MockCustomType, error) {
	opts = opts.WithDefaults()

	resource, err := client.rc.Read(namespace, name, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*MockCustomType), nil
}

func (client *mockCustomTypeClient) Write(mockCustomType *MockCustomType, opts clients.WriteOpts) (*MockCustomType, error) {
	opts = opts.WithDefaults()
	resource, err := client.rc.Write(mockCustomType, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*MockCustomType), nil
}

func (client *mockCustomTypeClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()

	return client.rc.Delete(namespace, name, opts)
}

func (client *mockCustomTypeClient) List(namespace string, opts clients.ListOpts) (MockCustomTypeList, error) {
	opts = opts.WithDefaults()

	resourceList, err := client.rc.List(namespace, opts)
	if err != nil {
		return nil, err
	}
	return convertToMockCustomType(resourceList), nil
}

func (client *mockCustomTypeClient) Watch(namespace string, opts clients.WatchOpts) (<-chan MockCustomTypeList, <-chan error, error) {
	opts = opts.WithDefaults()

	resourcesChan, errs, initErr := client.rc.Watch(namespace, opts)
	if initErr != nil {
		return nil, nil, initErr
	}
	mctsChan := make(chan MockCustomTypeList)
	go func() {
		for {
			select {
			case resourceList := <-resourcesChan:
				select {
				case mctsChan <- convertToMockCustomType(resourceList):
				case <-opts.Ctx.Done():
					close(mctsChan)
					return
				}
			case <-opts.Ctx.Done():
				close(mctsChan)
				return
			}
		}
	}()
	return mctsChan, errs, nil
}

func convertToMockCustomType(resources resources.ResourceList) MockCustomTypeList {
	var mockCustomTypeList MockCustomTypeList
	for _, resource := range resources {
		mockCustomTypeList = append(mockCustomTypeList, resource.(*MockCustomType))
	}
	return mockCustomTypeList
}
