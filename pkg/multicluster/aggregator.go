package multicluster

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

type KubeWatchAggregator struct {
	clientSets []ClientSet
}

func NewKubeWatchAggregator(clientSets ...ClientSet) KubeWatchAggregator {
	return KubeWatchAggregator{
		clientSets: clientSets,
	}
}

func (w KubeWatchAggregator) AggregatedWatch(namespace string) clients.ResourceWatch {
	// spin up watches based on clients
	return func(ctx context.Context) (<-chan resources.ResourceList, <-chan error, error) {
		listsByWatcher := newResourcesByWatchIndex()
		out := make(chan resources.ResourceList)
		aggregatedErrs := make(chan error)
		sourceWatches := sync.WaitGroup{}

		go func() {
			for {
				for _, cs := range w.clientSets {
					select {
					case c := <-cs.Clients():
						sourceWatches.Add(1)
						key := c.Cluster
						lists, errs, err := c.Client.Watch(namespace, clients.WatchOpts{Ctx: ctx})
						if err != nil {
							// TODO what to do
							return
						}

						go func() {
							defer sourceWatches.Done()
							defer listsByWatcher.delete(key)
							for {
								select {
								case <-ctx.Done():
									return
								case err := <-errs:
									select {
									case <-ctx.Done():
										return
									case aggregatedErrs <- err:
									}
									// if the source starts returning errors, remove its list from the snapshot
									select {
									case <-ctx.Done():
										return
									case out <- listsByWatcher.delete(key).merge():
									}
								case list, ok := <-lists:
									if !ok {
										return
									}
									// add/update the list to the snapshot
									select {
									case <-ctx.Done():
										return
									case out <- listsByWatcher.set(key, list).merge():
									}
								}
							}
						}()
					}
				}
			}
		}()

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

// aggregate resources by the cluster they were read from
type resourcesByWatchIndex struct {
	access    sync.RWMutex
	resources map[string]resources.ResourceList
}

func newResourcesByWatchIndex() *resourcesByWatchIndex {
	return &resourcesByWatchIndex{resources: make(map[string]resources.ResourceList)}
}

func (rbw *resourcesByWatchIndex) set(key string, val resources.ResourceList) *resourcesByWatchIndex {
	rbw.access.Lock()
	rbw.resources[key] = val
	rbw.access.Unlock()
	return rbw
}

func (rbw *resourcesByWatchIndex) delete(key string) *resourcesByWatchIndex {
	rbw.access.Lock()
	delete(rbw.resources, key)
	rbw.access.Unlock()
	return rbw
}

func (rbw *resourcesByWatchIndex) merge() resources.ResourceList {
	rbw.access.RLock()
	var merged resources.ResourceList
	for _, list := range rbw.resources {
		merged = append(merged, list...)
	}
	rbw.access.RUnlock()
	return merged.Sort()
}
