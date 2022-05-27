// Code generated by solo-kit. DO NOT EDIT.

//Generated by pkg/code-generator/codegen/templates/resource_client_template.go
package kubernetes

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type KubeNamespaceWatcher interface {
	// watch cluster-scoped kubenamespaces
	Watch(opts clients.WatchOpts) (<-chan KubeNamespaceList, <-chan error, error)
}

type KubeNamespaceClient interface {
	BaseClient() clients.ResourceClient
	Register() error
	Read(name string, opts clients.ReadOpts) (*KubeNamespace, error)
	Write(resource *KubeNamespace, opts clients.WriteOpts) (*KubeNamespace, error)
	Delete(name string, opts clients.DeleteOpts) error
	List(opts clients.ListOpts) (KubeNamespaceList, error)
	KubeNamespaceWatcher
}

type kubeNamespaceClient struct {
	rc clients.ResourceClient
}

func NewKubeNamespaceClient(ctx context.Context, rcFactory factory.ResourceClientFactory) (KubeNamespaceClient, error) {
	return NewKubeNamespaceClientWithToken(ctx, rcFactory, "")
}

func NewKubeNamespaceClientWithToken(ctx context.Context, rcFactory factory.ResourceClientFactory, token string) (KubeNamespaceClient, error) {
	rc, err := rcFactory.NewResourceClient(ctx, factory.NewResourceClientParams{
		ResourceType: &KubeNamespace{},
		Token:        token,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating base KubeNamespace resource client")
	}
	return NewKubeNamespaceClientWithBase(rc), nil
}

func NewKubeNamespaceClientWithBase(rc clients.ResourceClient) KubeNamespaceClient {
	return &kubeNamespaceClient{
		rc: rc,
	}
}

func (client *kubeNamespaceClient) BaseClient() clients.ResourceClient {
	return client.rc
}

func (client *kubeNamespaceClient) Register() error {
	return client.rc.Register()
}

func (client *kubeNamespaceClient) Read(name string, opts clients.ReadOpts) (*KubeNamespace, error) {
	opts = opts.WithDefaults()

	resource, err := client.rc.Read("", name, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*KubeNamespace), nil
}

func (client *kubeNamespaceClient) Write(kubeNamespace *KubeNamespace, opts clients.WriteOpts) (*KubeNamespace, error) {
	opts = opts.WithDefaults()
	resource, err := client.rc.Write(kubeNamespace, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*KubeNamespace), nil
}

func (client *kubeNamespaceClient) Delete(name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()

	return client.rc.Delete("", name, opts)
}

func (client *kubeNamespaceClient) List(opts clients.ListOpts) (KubeNamespaceList, error) {
	opts = opts.WithDefaults()

	resourceList, err := client.rc.List("", opts)
	if err != nil {
		return nil, err
	}
	return convertToKubeNamespace(resourceList), nil
}

func (client *kubeNamespaceClient) Watch(opts clients.WatchOpts) (<-chan KubeNamespaceList, <-chan error, error) {
	opts = opts.WithDefaults()

	resourcesChan, errs, initErr := client.rc.Watch("", opts)
	if initErr != nil {
		return nil, nil, initErr
	}
	kubenamespacesChan := make(chan KubeNamespaceList)
	go func() {
		for {
			select {
			case resourceList := <-resourcesChan:
				select {
				case kubenamespacesChan <- convertToKubeNamespace(resourceList):
				case <-opts.Ctx.Done():
					close(kubenamespacesChan)
					return
				}
			case <-opts.Ctx.Done():
				close(kubenamespacesChan)
				return
			}
		}
	}()
	return kubenamespacesChan, errs, nil
}

func convertToKubeNamespace(resources resources.ResourceList) KubeNamespaceList {
	var kubeNamespaceList KubeNamespaceList
	for _, resource := range resources {
		kubeNamespaceList = append(kubeNamespaceList, resource.(*KubeNamespace))
	}
	return kubeNamespaceList
}
