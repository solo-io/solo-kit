package vault

import (
	"time"

	"github.com/solo-io/solo-kit/pkg/api/shared"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

var _ clients.ResourceClient = new(PkiResourceClient)

type PkiResourceClient struct {
	//vault *api.Client

	resourceType resources.VersionedResource
}

func (p PkiResourceClient) Kind() string {
	return resources.Kind(p.resourceType)
}

func (p PkiResourceClient) NewResource() resources.Resource {
	return resources.Clone(p.resourceType)
}

func (p PkiResourceClient) Register() error {
	return nil
}

func (p PkiResourceClient) Read(namespace, name string, opts clients.ReadOpts) (resources.Resource, error) {
	//TODO implement me
	panic("implement me")
}

func (p PkiResourceClient) Write(resource resources.Resource, opts clients.WriteOpts) (resources.Resource, error) {
	//TODO implement me
	panic("implement me")
}

func (p PkiResourceClient) Delete(namespace, name string, opts clients.DeleteOpts) error {
	//TODO implement me
	panic("implement me")
}

func (p PkiResourceClient) List(namespace string, opts clients.ListOpts) (resources.ResourceList, error) {
	//TODO implement me
	panic("implement me")
}

func (p PkiResourceClient) ApplyStatus(statusClient resources.StatusClient, inputResource resources.InputResource, opts clients.ApplyStatusOpts) (resources.Resource, error) {
	return shared.ApplyStatus(p, statusClient, inputResource, opts)
}

func (p PkiResourceClient) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	opts = opts.WithDefaults()
	resourcesChan := make(chan resources.ResourceList)
	errs := make(chan error)

	listOpts := clients.ListOpts{
		Ctx:                opts.Ctx,
		Selector:           opts.Selector,
		ExpressionSelector: opts.ExpressionSelector,
	}

	go func() {
		// watch should open up with an initial read
		initialResourceList, initialResourceListErr := p.List(namespace, listOpts)
		if initialResourceListErr != nil {
			errs <- initialResourceListErr
			return
		}
		resourcesChan <- initialResourceList
		for {
			select {
			case <-time.After(opts.RefreshRate):
				resourceList, resourceListErr := p.List(namespace, listOpts)
				if resourceListErr != nil {
					errs <- resourceListErr
				}
				resourcesChan <- resourceList
			case <-opts.Ctx.Done():
				close(resourcesChan)
				close(errs)
				return
			}
		}
	}()

	return resourcesChan, errs, nil
}
