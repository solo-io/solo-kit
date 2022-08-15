package common

import (
	"time"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/cache"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

type KubeCoreResourceClient struct {
	ResourceType resources.Resource
}

func (rc *KubeCoreResourceClient) Kind() string {
	return resources.Kind(rc.ResourceType)
}

func (rc *KubeCoreResourceClient) NewResource() resources.Resource {
	return resources.Clone(rc.ResourceType)
}

func (rc *KubeCoreResourceClient) Register() error {
	return nil
}

type ResourceListFunc func(namespace string, opts clients.ListOpts) (resources.ResourceList, error)

func KubeResourceWatch(cache cache.Cache, listFunc ResourceListFunc, namespace string,
	opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	opts = opts.WithDefaults()

	watch := cache.Subscribe()

	resourcesChan := make(chan resources.ResourceList)
	errs := make(chan error)
	var previous *resources.ResourceList
	updateResourceList := func() {
		list, err := listFunc(namespace, clients.ListOpts{
			Ctx:      opts.Ctx,
			Selector: opts.Selector,
		})
		if err != nil {
			errs <- err
			return
		}
		if previous != nil {
			if list.Equal(*previous) {
				// prevent flooding the channel with duplicates
				return
			}
		}
		previous = &list
		resourcesChan <- list
	}

	go func() {
		defer cache.Unsubscribe(watch)
		defer close(resourcesChan)
		defer close(errs)

		// intentionally rate-limited so that our sync loops have time to complete before the next snapshot is sent
		timer := time.NewTicker(time.Second)
		defer timer.Stop()

		// watch should open up with an initial read
		updateResourceList()
		update := false
		for {
			select {
			case _, ok := <-watch:
				if !ok {
					return
				}
				update = true
			case <-timer.C:
				if update {
					updateResourceList()
					update = false
				}
			case <-opts.Ctx.Done():
				return
			}
		}
	}()

	return resourcesChan, errs, nil
}
