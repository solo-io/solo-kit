package multicluster

import (
	"github.com/solo-io/solo-kit/pkg/api/v1/clients"
	"github.com/solo-io/solo-kit/pkg/api/v1/clients/wrapper"
)

type clusterWatchAggregator struct {
	aggregator wrapper.WatchAggregator
}

var _ ClientForClusterHandler = &clusterWatchAggregator{}

// Provides a ClientForClusterHandler to sync an aggregated watch with clients available on a cluster.
func NewAggregatedWatchClusterClientHandler(aggregator wrapper.WatchAggregator) *clusterWatchAggregator {
	return &clusterWatchAggregator{aggregator: aggregator}
}

func (h *clusterWatchAggregator) HandleNewClusterClient(cluster string, client clients.ResourceClient) {
	h.aggregator.AddWatch(client)
}

func (h *clusterWatchAggregator) HandleRemovedClusterClient(cluster string, client clients.ResourceClient) {
	h.aggregator.RemoveWatch(client)
}
