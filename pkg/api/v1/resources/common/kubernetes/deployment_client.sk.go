// Code generated by solo-kit. DO NOT EDIT.

package kubernetes

import (
	"context"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/pkg/errors"
)

type DeploymentWatcher interface {
	// watch namespace-scoped deployments
	Watch(namespace string, opts clients.WatchOpts) (<-chan DeploymentList, <-chan error, error)
}

type DeploymentClient interface {
	BaseClient() clients.ResourceClient
	Register() error
	Read(namespace, name string, opts clients.ReadOpts) (*Deployment, error)
	Write(resource *Deployment, opts clients.WriteOpts) (*Deployment, error)
	Delete(namespace, name string, opts clients.DeleteOpts) error
	List(namespace string, opts clients.ListOpts) (DeploymentList, error)
	DeploymentWatcher
}

type deploymentClient struct {
	rc clients.ResourceClient
}

func NewDeploymentClient(ctx context.Context, rcFactory factory.ResourceClientFactory) (DeploymentClient, error) {
	return NewDeploymentClientWithToken(ctx, rcFactory, "")
}

func NewDeploymentClientWithToken(ctx context.Context, rcFactory factory.ResourceClientFactory, token string) (DeploymentClient, error) {
	rc, err := rcFactory.NewResourceClient(ctx, factory.NewResourceClientParams{
		ResourceType: &Deployment{},
		Token:        token,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "creating base Deployment resource client")
	}
	return NewDeploymentClientWithBase(rc), nil
}

func NewDeploymentClientWithBase(rc clients.ResourceClient) DeploymentClient {
	return &deploymentClient{
		rc: rc,
	}
}

func (client *deploymentClient) BaseClient() clients.ResourceClient {
	return client.rc
}

func (client *deploymentClient) Register() error {
	return client.rc.Register()
}

func (client *deploymentClient) Read(namespace, name string, opts clients.ReadOpts) (*Deployment, error) {
	opts = opts.WithDefaults()

	resource, err := client.rc.Read(namespace, name, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*Deployment), nil
}

func (client *deploymentClient) Write(deployment *Deployment, opts clients.WriteOpts) (*Deployment, error) {
	opts = opts.WithDefaults()
	resource, err := client.rc.Write(deployment, opts)
	if err != nil {
		return nil, err
	}
	return resource.(*Deployment), nil
}

func (client *deploymentClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	opts = opts.WithDefaults()

	return client.rc.Delete(namespace, name, opts)
}

func (client *deploymentClient) List(namespace string, opts clients.ListOpts) (DeploymentList, error) {
	opts = opts.WithDefaults()

	resourceList, err := client.rc.List(namespace, opts)
	if err != nil {
		return nil, err
	}
	return convertToDeployment(resourceList), nil
}

func (client *deploymentClient) Watch(namespace string, opts clients.WatchOpts) (<-chan DeploymentList, <-chan error, error) {
	opts = opts.WithDefaults()

	resourcesChan, errs, initErr := client.rc.Watch(namespace, opts)
	if initErr != nil {
		return nil, nil, initErr
	}
	deploymentsChan := make(chan DeploymentList)
	go func() {
		for {
			select {
			case resourceList := <-resourcesChan:
				select {
				case deploymentsChan <- convertToDeployment(resourceList):
				case <-opts.Ctx.Done():
					close(deploymentsChan)
					return
				}
			case <-opts.Ctx.Done():
				close(deploymentsChan)
				return
			}
		}
	}()
	return deploymentsChan, errs, nil
}

func convertToDeployment(resources resources.ResourceList) DeploymentList {
	var deploymentList DeploymentList
	for _, resource := range resources {
		deploymentList = append(deploymentList, resource.(*Deployment))
	}
	return deploymentList
}
