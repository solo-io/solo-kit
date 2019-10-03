package multicluster

import (
	"context"
	"sync"

	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/factory"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/kube/crd"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
	"github.com/solo-io/solo-kit/test/mocks/v2alpha1"
	"k8s.io/client-go/rest"
)

type contextCancelTuple struct {
	ctx    context.Context
	cancel context.CancelFunc
}

type clusterClientTuple struct {
	cluster string
	client  clients.ResourceClient
}

// TODO update this to be a complete snapshot's worth of clients
type WatchAggregator struct {
	clients chan clusterClientTuple
	cancels map[string]contextCancelTuple
}

func (w WatchAggregator) ClusterAdded(cluster string, restConfig *rest.Config) {
	krc := &factory.KubeResourceClientFactory{
		Cluster: cluster,
		Crd:     crd.Crd{}, // TODO generate
		Cfg:     restConfig,
		// TODO Pass in through opts to aggregator
		SharedCache:        nil,
		SkipCrdCreation:    false,
		NamespaceWhitelist: nil,
		ResyncPeriod:       0,
	}
	client, err := v2alpha1.NewMockResourceClient(krc)
	if err != nil {
		return
	}
	if err := client.Register(); err != nil {
		return
	}
	// TODO more error handling, e.g. if cluster already exists with name
	w.clients <- clusterClientTuple{cluster: cluster, client: client.BaseClient()}
	ctx, cancel := context.WithCancel(context.Background())
	w.cancels[cluster] = contextCancelTuple{ctx: ctx, cancel: cancel}
}

func (w WatchAggregator) ClusterRemoved(cluster string, restConfig *rest.Config) {
	// cancel watch
	cc, ok := w.cancels[cluster]
	if !ok {
		return
	}
	cc.cancel()
	delete(w.cancels, cluster)
}

func (w WatchAggregator) AggregatedWatch(namespace string) clients.ResourceWatch {
	// spin up clients based on stream
	return func(ctx context.Context) (<-chan resources.ResourceList, <-chan error, error) {
		listsByWatcher := newResourcesByWatchIndex()
		out := make(chan resources.ResourceList)
		aggregatedErrs := make(chan error)
		sourceWatches := sync.WaitGroup{}

		go func() {
			for {
				select {
				case c := <-w.clients:
					sourceWatches.Add(1)
					key := c.cluster
					lists, errs, err := c.client.Watch(namespace, clients.WatchOpts{Ctx: ctx})
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

// aggregate resources by the channel they were read from
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
