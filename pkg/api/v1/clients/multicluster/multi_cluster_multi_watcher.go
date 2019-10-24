package multicluster

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
)

type aggregatedWatchClusterClientHandler struct {
	watch wrapper.WatchAggregator
}

var _ ClusterClientHandler = &aggregatedWatchClusterClientHandler{}

func NewAggregatedWatchClusterClientHandler(delegate wrapper.WatchAggregator) ClusterClientHandler {
	return &aggregatedWatchClusterClientHandler{watch: delegate}
}

func (aw *aggregatedWatchClusterClientHandler) HandleNewClusterClient(cluster string, client clients.ResourceClient) {
	aw.watch.AddWatch(client)
}

func (aw *aggregatedWatchClusterClientHandler) HandleRemovedClusterClient(cluster string, client clients.ResourceClient) {
	aw.watch.RemoveWatch(client)
}
