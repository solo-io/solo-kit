// Code generated by solo-kit. DO NOT EDIT.

package v1

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type KubeConfigWatcher interface {
	// watch namespace-scoped kubeconfigs
	Watch(namespace string, opts clients.WatchOpts) (<-chan KubeConfigList, <-chan error, error)
}

type KubeConfigClient interface {
	BaseClient() clients.ResourceClient
	Register() error
	Read(namespace, name string, opts clients.ReadOpts) (*KubeConfig, error)
	Write(resource *KubeConfig, opts clients.WriteOpts) (*KubeConfig, error)
	Delete(namespace, name string, opts clients.DeleteOpts) error
	List(namespace string, opts clients.ListOpts) (KubeConfigList, error)
	KubeConfigWatcher
}

type kubeConfigClient struct {
	rc clients.ResourceClient
}

func NewKubeConfigClient(ctx context.Context, rcFactory factory.ResourceClientFactory) (KubeConfigClient, error) {
	return NewKubeConfigClientWithToken(ctx, rcFactory, "")
}

func NewKubeConfigClientWithToken(ctx context.Context, rcFactory factory.ResourceClientFactory, token string) (KubeConfigClient, error) {
	rc, err := rcFactory.NewResourceClient(ctx, factory.NewResourceClientParams{
		ResourceType: &KubeConfig{},
		Token:        token,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating base KubeConfig resource client")
	}
	return NewKubeConfigClientWithBase(rc), nil
}

func NewKubeConfigClientWithBase(rc clients.ResourceClient) KubeConfigClient {
	return &kubeConfigClient{
		rc: rc,
	}
}

func (client *kubeConfigClient) BaseClient() clients.ResourceClient {
	return client.rc
}

func (client *kubeConfigClient) Register() error {
	return client.rc.Register()
}

func (client *kubeConfigClient) Read(namespace, name string, opts clients.ReadOpts) (*KubeConfig, error) {
	opts = opts.WithDefaults()

	resource, err := client.rc.Read(namespace, name, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*KubeConfig), nil
}

func (client *kubeConfigClient) Write(kubeConfig *KubeConfig, opts clients.WriteOpts) (*KubeConfig, error) {
	opts = opts.WithDefaults()
	resource, err := client.rc.Write(kubeConfig, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*KubeConfig), nil
}

func (client *kubeConfigClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()

	return client.rc.Delete(namespace, name, opts)
}

func (client *kubeConfigClient) List(namespace string, opts clients.ListOpts) (KubeConfigList, error) {
	opts = opts.WithDefaults()

	resourceList, err := client.rc.List(namespace, opts)
	if err != nil {
		return nil, err
	}
	return convertToKubeConfig(resourceList), nil
}

func (client *kubeConfigClient) Watch(namespace string, opts clients.WatchOpts) (<-chan KubeConfigList, <-chan error, error) {
	opts = opts.WithDefaults()

	resourcesChan, errs, initErr := client.rc.Watch(namespace, opts)
	if initErr != nil {
		return nil, nil, initErr
	}
	kubeconfigsChan := make(chan KubeConfigList)
	go func() {
		for {
			select {
			case resourceList := <-resourcesChan:
				select {
				case kubeconfigsChan <- convertToKubeConfig(resourceList):
				case <-opts.Ctx.Done():
					close(kubeconfigsChan)
					return
				}
			case <-opts.Ctx.Done():
				close(kubeconfigsChan)
				return
			}
		}
	}()
	return kubeconfigsChan, errs, nil
}

func convertToKubeConfig(resources resources.ResourceList) KubeConfigList {
	var kubeConfigList KubeConfigList
	for _, resource := range resources {
		kubeConfigList = append(kubeConfigList, resource.(*KubeConfig))
	}
	return kubeConfigList
}
