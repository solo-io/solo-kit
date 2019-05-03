package wrapper

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

func ResourceWatch(rw clients.ResourceWatcher, namespace string, selector map[string]string) clients.ResourceWatch {
	return func(ctx context.Context) (<-chan resources.ResourceList, <-chan error, error) {
		return rw.Watch(namespace, clients.WatchOpts{
			Ctx:      ctx,
			Selector: selector,
		})
	}
}

func AggregatedWatch(watches ...clients.ResourceWatch) clients.ResourceWatch {
	return func(ctx context.Context) (<-chan resources.ResourceList, <-chan error, error) {
		listsByWatcher := make(resourcesByWatch)
		access := sync.Mutex{}
		out := make(chan resources.ResourceList)
		aggregatedErrs := make(chan error)
		sourceWatches := sync.WaitGroup{}

		for i, w := range watches {
			sourceWatches.Add(1)
			key := i
			lists, errs, err := w(ctx)
			if err != nil {
				return nil, nil, err
			}
			go func() {
				defer sourceWatches.Done()
				for {
					select {
					case <-ctx.Done():
						return
					case err := <-errs:
						// if the source starts returning errors, remove its list from the snasphot
						access.Lock()
						delete(listsByWatcher, key)
						mergedList := listsByWatcher.merge()
						access.Unlock()
						aggregatedErrs <- err
						select {
						case <-ctx.Done():
							return
						case out <- mergedList:
						}
					case list, ok := <-lists:
						if !ok {
							return
						}
						// add/update the list to the snapshot
						access.Lock()
						listsByWatcher[key] = list
						mergedList := listsByWatcher.merge()
						access.Unlock()
						select {
						case <-ctx.Done():
							return
						case out <- mergedList:
						}
					}
				}
			}()
		}

		go func() {
			// context is closed, clean up watch resources
			<-ctx.Done()
			// wait for source watches to be closed before closing the sink
			sourceWatches.Wait()
			close(out)
			close(aggregatedErrs)
		}()
		return out, aggregatedErrs, nil
	}
}

// aggregate resources by the channel they were read from
type resourcesByWatch map[int]resources.ResourceList

func (rbw resourcesByWatch) merge() resources.ResourceList {
	var merged resources.ResourceList
	for _, list := range rbw {
		merged = append(merged, list...)
	}
	return merged.Sort()
}
