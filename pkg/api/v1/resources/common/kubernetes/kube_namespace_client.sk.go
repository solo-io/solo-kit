// Code generated by solo-kit. DO NOT EDIT.

package kubernetes

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type KubeNamespaceWatcher interface {
	// watch namespace-scoped kubenamespaces
	Watch(namespace string, opts clients.WatchOpts) (<-chan KubeNamespaceList, <-chan error, error)
}

type KubeNamespaceInterface interface {
	Read(namespace, name string, opts clients.ReadOpts) (*KubeNamespace, error)
	Write(resource *KubeNamespace, opts clients.WriteOpts) (*KubeNamespace, error)
	Delete(namespace, name string, opts clients.DeleteOpts) error
	List(namespace string, opts clients.ListOpts) (KubeNamespaceList, error)
	KubeNamespaceWatcher
}

type KubeNamespaceClient interface {
	BaseClient() clients.ResourceClient
	Register() error
	KubeNamespaceInterface
}

type kubeNamespaceClient struct {
	rc clients.ResourceClient
}

func NewKubeNamespaceClient(rcFactory factory.ResourceClientFactory) (KubeNamespaceClient, error) {
	return NewKubeNamespaceClientWithToken(rcFactory, "")
}

func NewKubeNamespaceClientWithToken(rcFactory factory.ResourceClientFactory, token string) (KubeNamespaceClient, error) {
	rc, err := rcFactory.NewResourceClient(factory.NewResourceClientParams{
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

func (client *kubeNamespaceClient) Read(namespace, name string, opts clients.ReadOpts) (*KubeNamespace, error) {
	opts = opts.WithDefaults()

	resource, err := client.rc.Read(namespace, name, opts)
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

func (client *kubeNamespaceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()

	return client.rc.Delete(namespace, name, opts)
}

func (client *kubeNamespaceClient) List(namespace string, opts clients.ListOpts) (KubeNamespaceList, error) {
	opts = opts.WithDefaults()

	resourceList, err := client.rc.List(namespace, opts)
	if err != nil {
		return nil, err
	}
	return convertToKubeNamespace(resourceList), nil
}

func (client *kubeNamespaceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan KubeNamespaceList, <-chan error, error) {
	opts = opts.WithDefaults()

	resourcesChan, errs, initErr := client.rc.Watch(namespace, opts)
	if initErr != nil {
		return nil, nil, initErr
	}
	kubenamespacesChan := make(chan KubeNamespaceList)
	go func() {
		for {
			select {
			case resourceList := <-resourcesChan:
				kubenamespacesChan <- convertToKubeNamespace(resourceList)
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
