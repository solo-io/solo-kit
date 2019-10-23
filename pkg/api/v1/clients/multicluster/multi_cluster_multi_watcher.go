package multicluster

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
	"github.com/solo-io/solo-kit/pkg/api/v1/resources"
)

type MultiClusterAggregatedWatch interface {
	wrapper.WatchAggregator
	ClusterClientHandler
}

type multiClusterAggregatedWatch struct {
	delegate wrapper.WatchAggregator
}

func NewMultiClusterAggregatedWatch(delegate wrapper.WatchAggregator) MultiClusterAggregatedWatch {
	return &multiClusterAggregatedWatch{delegate: delegate}
}

func (aw *multiClusterAggregatedWatch) Watch(namespace string, opts clients.WatchOpts) (<-chan resources.ResourceList, <-chan error, error) {
	return aw.delegate.Watch(namespace, opts)
}

func (aw *multiClusterAggregatedWatch) AddWatch(w clients.ResourceWatcher) error {
	return aw.delegate.AddWatch(w)
}

func (aw *multiClusterAggregatedWatch) RemoveWatch(w clients.ResourceWatcher) {
	aw.delegate.RemoveWatch(w)
}

func (aw *multiClusterAggregatedWatch) HandleNewClusterClient(cluster string, client clients.ResourceClient) {
	aw.delegate.AddWatch(client)
}

func (aw *multiClusterAggregatedWatch) HandleRemovedClusterClient(cluster string, client clients.ResourceClient) {
	aw.delegate.RemoveWatch(client)
}
