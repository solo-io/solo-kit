package common

import (
	"fmt"
	"time"

	"github.com/solo-io/gloo/test/debugprint"
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
	// prevent flooding the channel with duplicates
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

		//fmt.Println(len(list.Names()))
		//
		if len(list.Names()) != 1 && len(list.Names()) != 56 && previous != nil && !list.Equal(*previous){
			og := debugprint.SprintAny(list)
			new := debugprint.SprintAny(*previous)
			fmt.Println(og)
			fmt.Println(new)

			fmt.Println("kdorosh")

			eq := list[0].Equal((*previous)[0])
			fmt.Println(eq)
			og2 := debugprint.SprintAny(list[0])
			new2 := debugprint.SprintAny((*previous)[0])
			eq2 := og2 == new2
			fmt.Println(eq2)
		}

		if previous != nil {
			og := debugprint.SprintAny(list)
			new := debugprint.SprintAny(*previous)
			if og == new {
				return
			}

			if list.Equal(*previous) {
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

		// watch should open up with an initial read
		timer := time.NewTicker(time.Second)
		defer timer.Stop()

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
